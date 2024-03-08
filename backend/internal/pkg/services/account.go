package services

import (
	"net/http"

	"github.com/nikitaSstepanov/p2p-streaming-service/backend/internal/pkg/storage/entities"
	"github.com/nikitaSstepanov/p2p-streaming-service/backend/internal/pkg/storage"
	"github.com/nikitaSstepanov/p2p-streaming-service/backend/internal/pkg/dto"
	"golang.org/x/crypto/bcrypt"
	"github.com/gin-gonic/gin"
)

type Account struct {
	Storage *storage.Storage
}

func NewAccount(storage *storage.Storage) *Account {
	return &Account{
		Storage: storage,
	}
}

func (a *Account) Create(ctx *gin.Context) {
	var dto dto.CreateUserDto

	if err := ctx.ShouldBindJSON(&dto); err != nil {
		return 
	}

	candidate := a.Storage.Users.GetUser(ctx, dto.Username)

	if candidate.Id != 0 {
		ctx.JSON(http.StatusBadRequest, "This username is taken.")
		return
	}

	password, err := hashPassword(dto.Password)

	if err != nil {
		return
	}

	user := &entities.User{
		Username: dto.Username,
		Password: password,
	}

	a.Storage.Users.Create(ctx, user)
}

func (a *Account) SignIn(ctx *gin.Context) {

}

func hashPassword(password string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), 10)

	if err != nil {
		return "", err
	}

	return string(hash), nil
}
