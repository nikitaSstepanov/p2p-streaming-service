package services

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/nikitaSstepanov/p2p-streaming-service/backend/internal/pkg/dto/users"
	"github.com/nikitaSstepanov/p2p-streaming-service/backend/internal/pkg/storage"
	"github.com/nikitaSstepanov/p2p-streaming-service/backend/internal/pkg/storage/entities"
	"github.com/spf13/viper"
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
	var body dto.SignInDto

	if err := ctx.ShouldBindJSON(&body); err != nil {
		return
	}

	user := a.Storage.Users.GetUser(ctx, body.Username)

	if user.Id == 0 {
		return
	}

	err := checkPassword(user.Password, body.Password)

	if err != nil {
		return
	}

	token, err := generateToken(user.Username)

	if err != nil {
		return
	}

	result := &dto.TokenResult{
		Token: token,
	}

	ctx.JSON(200, result)
}

func hashPassword(password string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), 10)

	if err != nil {
		return "", err
	}

	return string(hash), nil
}

func checkPassword(hash string, password string) error {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))

	if err != nil {
		return err
	}

	return nil
}

func generateToken(username string) (string, error) {
	claims := jwt.MapClaims{
		"exp": time.Now().Add(1 * time.Hour),
		"username": username,
	}

	key := viper.GetString("jwt.key")

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	tokenString, err := token.SignedString([]byte(key))

	if err != nil {
		return "", err
	}

	return tokenString, nil
}
