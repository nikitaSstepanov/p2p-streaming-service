package services

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/nikitaSstepanov/p2p-streaming-service/backend/internal/pkg/storage"
	"github.com/nikitaSstepanov/p2p-streaming-service/backend/internal/pkg/storage/entities"
	"golang.org/x/crypto/bcrypt"
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
	type CreateUserDto struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}

	var dto CreateUserDto

	if err := ctx.BindJSON(&dto); err != nil {
        return
    }

	candidate := a.Storage.Users.GetUser(ctx, dto.Username)

	if (candidate.Id != 0) {
		ctx.JSON(http.StatusBadRequest, "This username is taken")
		return
	}

	user := &entities.User{
		Username: dto.Username,
		Password: HashPassword(dto.Password),
	}

	a.Storage.Users.Create(ctx, user)
}

func (a *Account) SignIn(ctx *gin.Context) {

}

func HashPassword(password string) string {
	hash, _ := bcrypt.GenerateFromPassword([]byte(password), 10)

	return string(hash)
}
