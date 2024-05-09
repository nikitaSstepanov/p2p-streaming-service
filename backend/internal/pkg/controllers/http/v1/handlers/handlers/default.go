package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

type Default struct {}

func NewDefault() *Default {
	return &Default{}
}

func (d *Default) Ping(ctx *gin.Context) {
	ctx.JSON(http.StatusOK, "pong")
}