package handlers

import (
	"net/http"
	"strconv"

	"github.com/nikitaSstepanov/p2p-streaming-service/backend/internal/pkg/usecases/services"
	"github.com/nikitaSstepanov/p2p-streaming-service/backend/internal/pkg/types/responses"
	"github.com/nikitaSstepanov/p2p-streaming-service/backend/internal/pkg/types/statuses"
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

	result, status := m.services.Movies.GetMovies(ctx, limit, offset)

	if status == statuses.OK {
		ctx.JSON(http.StatusOK, result)
	} else {
		err := &responses.Error{
			Error: "Something going wrong.....",
		}
		ctx.JSON(http.StatusInternalServerError, err)
	}
}

func (m *Movies) GetMovieById(ctx *gin.Context) {
	movieId := ctx.Param("id")

	result, status := m.services.Movies.GetMovieById(ctx, movieId)

	if status == statuses.OK {
		ctx.JSON(http.StatusOK, result)
	} else if status == statuses.NotFound {
		err := &responses.Error{
			Error: "This movie wasn`t found.",
		}

		ctx.JSON(http.StatusNotFound, err)
	}
}

func (m *Movies) StartWatch(ctx *gin.Context) {
	movieId := ctx.Param("id")

	result, status := m.services.Movies.StartWatch(ctx, movieId)

	if status == statuses.OK {
		ctx.JSON(http.StatusOK, result)
	} else {
		m.handleError(ctx, status)
	}
}

func (m *Movies) GetMovieChunck(ctx *gin.Context) {
	movieId := ctx.Param("id")

	fileId, err := strconv.ParseInt(ctx.Param("fileId"), 10, 64)

	if err != nil {
		ctx.JSON(http.StatusBadRequest, "fileId must be integer")
		return
	}

	chunkId, err := strconv.Atoi(ctx.Param("chunkId"))
	
	if err != nil {
		ctx.JSON(http.StatusBadRequest, "chunkId must be integer")
		return
	}

	result, status := m.services.Movies.GetMovieChunck(ctx, movieId, fileId, chunkId)

	if status == statuses.OK {
		ctx.JSON(http.StatusOK, result)
	} else {
		m.handleError(ctx, status)
	}
}

func (m *Movies) handleError(ctx *gin.Context, status string) {
	e := &responses.Error{}

	switch status {

	case statuses.NotFound:
		e.Error = "This movie wasn`t found."
		ctx.JSON(http.StatusNotFound, e)

	case statuses.InternalError:
		e.Error = "Something going wrong....."
		ctx.JSON(http.StatusInternalServerError, e)

	}
}