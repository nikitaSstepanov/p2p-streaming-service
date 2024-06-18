package handlers

import (
	"net/http"

	"github.com/nikitaSstepanov/p2p-streaming-service/backend/internal/usecases/services"
	"github.com/nikitaSstepanov/p2p-streaming-service/backend/internal/types/responses"
	"github.com/nikitaSstepanov/p2p-streaming-service/backend/internal/types/statuses"
	"github.com/gin-gonic/gin"
)

type Movies struct {
	services *services.Services
}

func NewMovies(services *services.Services) *Movies {
	return &Movies{
		services: services,
	}
}

func (m *Movies) GetMovies(ctx *gin.Context) {
	limit := ctx.DefaultQuery("limit", "5")

	offset := ctx.DefaultQuery("offset", "0")

	result, status, msg := m.services.Movies.GetMovies(ctx, limit, offset)

	if status == statuses.OK {
		ctx.JSON(http.StatusOK, result)
	} else {
		m.handleError(ctx, status, msg)
	}
}

func (m *Movies) GetMovieById(ctx *gin.Context) {
	movieId := ctx.Param("id")

	result, status, msg := m.services.Movies.GetMovieById(ctx, movieId)

	if status == statuses.OK {
		ctx.JSON(http.StatusOK, result)
	} else {
		m.handleError(ctx, status, msg)
	}
}

func (m *Movies) StartWatch(ctx *gin.Context) {
	movieId := ctx.Param("id")

	result, status, msg := m.services.Movies.StartWatch(ctx, movieId)

	if status == statuses.OK {
		ctx.JSON(http.StatusOK, result)
	} else {
		m.handleError(ctx, status, msg)
	}
}

func (m *Movies) GetMovieChunck(ctx *gin.Context) {
	movieId := ctx.Param("id")

	fileId := ctx.Param("fileId") 

	chunkId := ctx.Param("chunkId")

	result, status, msg := m.services.Movies.GetMovieChunck(ctx, movieId, fileId, chunkId)

	if status == statuses.OK {
		ctx.JSON(http.StatusOK, result)
	} else {
		m.handleError(ctx, status, msg)
	}
}

func (m *Movies) handleError(ctx *gin.Context, status string, message string) {
	e := &responses.Error{
		Error: message,
	}

	switch status {

	case statuses.NotFound:
		ctx.JSON(http.StatusNotFound, e)

	case statuses.InternalError:
		ctx.JSON(http.StatusInternalServerError, e)

	case statuses.BadRequest:
		ctx.JSON(http.StatusBadRequest, e)
	}
}