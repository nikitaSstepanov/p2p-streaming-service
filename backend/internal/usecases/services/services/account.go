package services

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/nikitaSstepanov/p2p-streaming-service/backend/internal/types/dto/users"
	"github.com/nikitaSstepanov/p2p-streaming-service/backend/internal/types/statuses"
	"github.com/nikitaSstepanov/p2p-streaming-service/backend/internal/usecases/storage"
	"github.com/nikitaSstepanov/p2p-streaming-service/backend/internal/usecases/storage/entities"
	"github.com/redis/go-redis/v9"
)

type Account struct {
	storage *storage.Storage
	redis   *redis.Client
	auth    *Auth
}

func NewAccount(storage *storage.Storage, redis *redis.Client, auth *Auth) *Account {
	return &Account{
		storage: storage,
		redis:   redis,
		auth:    auth,
	}
}

func (a *Account) GetAccount(ctx context.Context, header string) (*dto.AccountDto, string, string) {
	parts := strings.Split(header, " ")
	bearer := parts[0]
	token := parts[1]

	if bearer != "Bearer" {
		return nil, statuses.Unauthorize, "Token must be Bearer."
	}

	claims, err := a.auth.ValidateToken(token)

	if err != nil {
		return nil, statuses.Unauthorize, "Invalid token."
	}

	user, err := a.findUser(ctx, claims.Username)
	
	if err != nil {
		return nil, statuses.InternalError, "Something going wrong."
	}

	if user.Id == 0 {
		return nil, statuses.Unauthorize, "Invalid token."
	}

	dto := &dto.AccountDto{
		Username: user.Username,
	}

	return dto, statuses.OK, ""
}

func (a *Account) Create(ctx context.Context, data *dto.CreateUserDto) (*dto.AccountDto, string, string) {
	candidate, err := a.findUser(ctx, data.Username)

	if err != nil {
		return nil, statuses.InternalError, "Something going wrong..."
	}

	if candidate.Id != 0 {
		return nil, statuses.Conflict, "This username is taken."
	}

	password, err := a.auth.HashPassword(data.Password)

	if err != nil {
		return nil, statuses.BadRequest, "Invalid password."
	}

	user := &entities.User{
		Username: data.Username,
		Password: password,
		Role:     "USER",
	}

	a.storage.Users.Create(ctx, user)

	result := &dto.AccountDto{
		Username: user.Username,
	}

	return result, statuses.OK, ""
}

func (a *Account) SignIn(ctx context.Context, body *dto.SignInDto) (*dto.TokenResult, string, string) {
	user, err := a.findUser(ctx, body.Username)
	
	if err != nil {
		return nil, statuses.InternalError, "Something going wrong...."
	}

	if user.Id == 0 {
		return nil, statuses.Unauthorize, "Invalid username or password."
	}

	err = a.auth.CheckPassword(user.Password, body.Password)

	if err != nil {
		return nil, statuses.Unauthorize, "Invalid username or password."
	}

	token, err := a.auth.GenerateToken(user.Username)

	if err != nil {
		return nil, statuses.Unauthorize, "Can`t generate token."
	}

	result := &dto.TokenResult{
		Token: token,
	}

	return result, statuses.OK, ""
}

func (a *Account) findUser(ctx context.Context, username string) (*entities.User, error) {
	var user entities.User

	a.redis.Get(ctx, fmt.Sprintf("users:%s", username)).Scan(&user)

	if user.Id != 0 {
		return &user, nil
	}
	
	user, err := a.storage.Users.GetUser(ctx, username)

	if err != nil {
		return nil, err
	}

	a.redis.Set(ctx, fmt.Sprintf("users:%s", username), user, 1 * time.Hour)

	return &user, nil
}

func (a *Account) getUser(ctx context.Context, header string) (*entities.User, string) {
	parts := strings.Split(header, " ")
	bearer := parts[0]
	token := parts[1]

	if bearer != "Bearer" {
		return nil, statuses.Unauthorize
	}

	claims, err := a.auth.ValidateToken(token)

	if err != nil {
		return nil, statuses.Unauthorize
	}

	user, err := a.findUser(ctx, claims.Username)

	if err != nil {
		return nil, statuses.InternalError
	}

	if user.Id == 0 {
		return nil, statuses.Unauthorize
	}

	return user, statuses.OK
}
