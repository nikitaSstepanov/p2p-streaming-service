package services

import (
	"github.com/gin-gonic/gin"
	"github.com/nikitaSstepanov/p2p-streaming-service/backend/internal/pkg/storage"
)

type Comments struct {
	Storage *storage.Storage
}

func NewComments(storage *storage.Storage) *Comments {
	return &Comments{
		Storage: storage,
	}
}

func (c *Comments) GetComments(ctx *gin.Context) {

}

func (c *Comments) GetCommentById(ctx *gin.Context) {

}

func (c *Comments) CreateComment(ctx *gin.Context) {

}

func (c *Comments) EditComment(ctx *gin.Context) {

}

func (c *Comments) DeleteComment(ctx *gin.Context) {
	
}