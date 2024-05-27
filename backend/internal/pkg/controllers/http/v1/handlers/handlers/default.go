package handlers

import (
	"net/http"

	"github.com/nikitaSstepanov/p2p-streaming-service/backend/internal/pkg/types/responses"
	"github.com/gin-gonic/gin"
)

type Default struct {}

func NewDefault() *Default {
	return &Default{}
}

func (d *Default) Ping(ctx *gin.Context) {
	result := &responses.Message{
		Message: "pong",
	}

	ctx.JSON(http.StatusOK, result)
}