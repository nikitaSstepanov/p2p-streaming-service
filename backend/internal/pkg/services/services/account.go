package services

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/nikitaSstepanov/p2p-streaming-service/backend/internal/pkg/services/dto/users"
	"github.com/nikitaSstepanov/p2p-streaming-service/backend/internal/pkg/storage"
	"github.com/nikitaSstepanov/p2p-streaming-service/backend/internal/pkg/storage/entities"
	"github.com/redis/go-redis/v9"
)

type Account struct {
	Storage *storage.Storage
	Redis   *redis.Client
	Auth    *Auth
}

func NewAccount(storage *storage.Storage, redis *redis.Client, auth *Auth) *Account {
	return &Account{
		Storage: storage,
		Redis:   redis,
		Auth:    auth,
	}
}

func (a *Account) GetAccount(ctx *gin.Context) {
	header := strings.Split(ctx.GetHeader("Authorization"), " ")
	bearer := header[0]
	token := header[1]

	if bearer != "Bearer" {
		ctx.JSON(http.StatusUnauthorized, "Incorrect token.")
		return
	}

	claims, err := a.Auth.ValidateToken(token)

	if err != nil {
		ctx.JSON(http.StatusUnauthorized, "Incorrecct token.")
		return
	}

	user := a.findUser(ctx, claims.Username)

	if user.Id == 0 {
		ctx.JSON(http.StatusUnauthorized, "Incorrecct token.")
		return
	}

	dto := &dto.AccountDto{
		Username: user.Username,
	}

	ctx.JSON(http.StatusOK, dto)
}

func (a *Account) Create(ctx *gin.Context) {
	var data dto.CreateUserDto

	if err := ctx.ShouldBindJSON(&data); err != nil {
		ctx.JSON(http.StatusBadRequest, "Incorrect data.")
		return
	}

	candidate := a.findUser(ctx, data.Username)

	if candidate.Id != 0 {
		ctx.JSON(http.StatusConflict, "This username is taken.")
		return
	}

	password, err := a.Auth.HashPassword(data.Password)

	if err != nil {
		ctx.JSON(http.StatusBadRequest, "Incorrect data.")
		return
	}

	user := &entities.User{
		Username: data.Username,
		Password: password,
		Role:     "USER",
	}

	a.Storage.Users.Create(ctx, user)

	result := dto.AccountDto{
		Username: user.Username,
	}

	ctx.JSON(http.StatusCreated, result)
}

func (a *Account) SignIn(ctx *gin.Context) {
	var body dto.SignInDto

	if err := ctx.ShouldBindJSON(&body); err != nil {
		ctx.JSON(http.StatusBadRequest, "Incorrect data.")
		return
	}

	user := a.findUser(ctx, body.Username)

	if user.Id == 0 {
		ctx.JSON(http.StatusUnauthorized, "Incorrect username or password.")
		return
	}

	err := a.Auth.CheckPassword(user.Password, body.Password)

	if err != nil {
		ctx.JSON(http.StatusUnauthorized, "Incorrect username or password.")
		return
	}

	token, err := a.Auth.GenerateToken(user.Username)

	if err != nil {
		ctx.JSON(http.StatusUnauthorized, "Incorrect username or password.")
		return
	}

	result := &dto.TokenResult{
		Token: token,
	}

	ctx.JSON(http.StatusOK, result)
}

func (a *Account) findUser(ctx context.Context, username string) entities.User {
	var user entities.User

	a.Redis.Get(ctx, fmt.Sprintf("users:%s", username)).Scan(&user)

	if user.Id == 0 {
		user = *(a.Storage.Users.GetUser(ctx, username))

		if user.Id == 0 {
			return entities.User{}
		}

		a.Redis.Set(ctx, fmt.Sprintf("users:%s", username), user, 1 * time.Hour)
	}

	return user
}
