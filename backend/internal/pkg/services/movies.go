package services

import (
	"net/http"

	"github.com/nikitaSstepanov/p2p-streaming-service/backend/internal/pkg/storage"
	"github.com/gin-gonic/gin"
)

type Movies struct {
	Storage *storage.Storage
}

func NewMovies(storage *storage.Storage) *Movies {
	return &Movies{
		Storage: storage,
	}
}

func (m *Movies) GetMovies(ctx *gin.Context) {
	limit := ctx.DefaultQuery("limit", "5")

	offset := ctx.DefaultQuery("offset", "0")

	movies := m.Storage.Movies.GetAllMovies(ctx, limit, offset)

	ctx.JSON(http.StatusOK, movies)
}

func (m *Movies) GetMovieById(ctx *gin.Context) {

}

func (m *Movies) GetMovieChunck() {

}
