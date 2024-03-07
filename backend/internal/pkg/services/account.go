package services

import (
	"github.com/gin-gonic/gin"
)

type Account struct {}

func NewAccount() *Account {
	return &Account{}
}

func (a *Account) Create(ctx *gin.Context) {

}

func (a *Account) SignIn(ctx *gin.Context) {

}