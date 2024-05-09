package services

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/nikitaSstepanov/p2p-streaming-service/backend/internal/pkg/types/dto/users"
	"github.com/nikitaSstepanov/p2p-streaming-service/backend/internal/pkg/types/statuses"
	"github.com/nikitaSstepanov/p2p-streaming-service/backend/internal/pkg/usecases/storage"
	"github.com/nikitaSstepanov/p2p-streaming-service/backend/internal/pkg/usecases/storage/entities"
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

func (a *Account) GetAccount(ctx context.Context, header string) (*dto.AccountDto, string) {
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

	user := a.findUser(ctx, claims.Username)

	if user.Id == 0 {
		return nil, statuses.Unauthorize
	}

	dto := &dto.AccountDto{
		Username: user.Username,
	}

	return dto, statuses.OK
}

func (a *Account) Create(ctx context.Context, data *dto.CreateUserDto) (*dto.AccountDto, string) {
	candidate := a.findUser(ctx, data.Username)

	if candidate.Id != 0 {
		return nil, statuses.Conflict
	}

	password, err := a.auth.HashPassword(data.Password)

	if err != nil {
		return nil, statuses.BadRequest
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

	return result, statuses.OK
}

func (a *Account) SignIn(ctx context.Context, body *dto.SignInDto) (*dto.TokenResult, string) {
	user := a.findUser(ctx, body.Username)

	if user.Id == 0 {
		return nil, statuses.Unauthorize
	}

	err := a.auth.CheckPassword(user.Password, body.Password)

	if err != nil {
		return nil, statuses.Unauthorize
	}

	token, err := a.auth.GenerateToken(user.Username)

	if err != nil {
		return nil, statuses.Unauthorize
	}

	result := &dto.TokenResult{
		Token: token,
	}

	return result, statuses.OK
}

func (a *Account) findUser(ctx context.Context, username string) entities.User {
	var user entities.User

	a.redis.Get(ctx, fmt.Sprintf("users:%s", username)).Scan(&user)

	if user.Id == 0 {
		user = *(a.storage.Users.GetUser(ctx, username))

		if user.Id == 0 {
			return entities.User{}
		}

		a.redis.Set(ctx, fmt.Sprintf("users:%s", username), user, 1 * time.Hour)
	}

	return user
}
