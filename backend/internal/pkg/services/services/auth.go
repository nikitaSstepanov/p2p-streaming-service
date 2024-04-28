package services

import (
	"time"
	"fmt"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
	"github.com/spf13/viper"
)

type Auth struct {}

type Claims struct {
	Username string `json:"usr"`
	jwt.RegisteredClaims
}

func NewAuth() *Auth {
	return &Auth{}
}

func (a *Auth) ValidateToken(jwtString string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(jwtString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		return []byte(viper.GetString("jwt.key")), nil
	})

	if err != nil || !token.Valid {
		return &Claims{}, fmt.Errorf("401")
	}

	return token.Claims.(*Claims), nil
}

func (a *Auth) GenerateToken(username string) (string, error) {
	claims := Claims{
		username,
		jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(24 * time.Hour)),
			Issuer:    viper.GetString("jwt.issuer"),
			Audience:  []string{viper.GetString("jwt.audience")},
		},
	}
	
	key := viper.GetString("jwt.key")

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	tokenString, err := token.SignedString([]byte(key))

	if err != nil {
		return "", err
	}

	return tokenString, nil
}

func (a *Auth) CheckPassword(hash string, password string) error {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))

	if err != nil {
		return err
	}

	return nil
}

func (a *Auth) HashPassword(password string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), 10)

	if err != nil {
		return "", err
	}

	return string(hash), nil
}