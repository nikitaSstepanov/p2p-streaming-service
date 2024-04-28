package services

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/nikitaSstepanov/p2p-streaming-service/backend/internal/pkg/services/dto/users"
	"github.com/nikitaSstepanov/p2p-streaming-service/backend/internal/pkg/storage"
	"github.com/nikitaSstepanov/p2p-streaming-service/backend/internal/pkg/storage/entities"
)

type Account struct {
	Storage *storage.Storage
	Auth *Auth
}



func NewAccount(storage *storage.Storage, auth *Auth) *Account {
	return &Account{
		Storage: storage,
		Auth: auth,
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

	user := a.Storage.Users.GetUser(ctx, claims.Username)

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

	candidate := a.Storage.Users.GetUser(ctx, data.Username)

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
		Role: "USER",
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

	user := a.Storage.Users.GetUser(ctx, body.Username)

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

