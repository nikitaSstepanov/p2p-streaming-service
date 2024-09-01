package movie

import (
	"context"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/nikitaSstepanov/p2p-streaming-service/backend/internal/controller/http/v1/dto"
	"github.com/nikitaSstepanov/p2p-streaming-service/backend/internal/entity"
	e "github.com/nikitaSstepanov/p2p-streaming-service/backend/pkg/errors"
)

const (
	ok = http.StatusOK
	badReq = http.StatusBadRequest
)

var (
	badReqErr  = e.New("Incorrect data.", e.BadInput)
)

type MovieUseCase interface {
	GetMovies(ctx context.Context, limit int, offset int) ([]*entity.Movie, *e.Error)
	GetMovieById(ctx context.Context, id uint64) (*entity.Movie, *e.Error)
	StartWatch(ctx context.Context, movieId uint64) (*entity.Chunk, *e.Error)
	GetMovieChunck(ctx context.Context, movieId uint64, fileId int, index int) (*entity.Chunk, *e.Error)
}

type Movies struct {
	usecase MovieUseCase
}

func New(uc MovieUseCase) *Movies {
	return &Movies{
		usecase: uc,
	}
}

func (m *Movies) GetMovies(ctx *gin.Context) {
	limit, err := strconv.Atoi(ctx.DefaultQuery("limit", "5"))
	if err != nil {
		ctx.AbortWithStatusJSON(badReq, badReqErr)
		return
	}

	offset, err := strconv.Atoi(ctx.DefaultQuery("offset", "0"))
	if err != nil {
		ctx.AbortWithStatusJSON(badReq, badReqErr)
		return
	}

	movies, movieErr := m.usecase.GetMovies(ctx, limit, offset)
	if movieErr != nil {
		ctx.AbortWithStatusJSON(movieErr.ToHttpCode(), movieErr)
		return
	}

	result := make([]dto.MovieDto, 0)

	for i := 0; i < len(movies); i++ {
		result = append(result, dto.MovieToDto(movies[i]))
	}

	ctx.JSON(ok, result)
}

func (m *Movies) GetMovieById(ctx *gin.Context) {
	id, err := strconv.ParseUint(ctx.Param("id"), 10, 64)
	if err != nil {
		ctx.AbortWithStatusJSON(badReq, badReqErr)
		return
	}

	movie, movieErr := m.usecase.GetMovieById(ctx, id)
	if movieErr != nil {
		ctx.AbortWithStatusJSON(movieErr.ToHttpCode(), movieErr)
		return
	}

	ctx.JSON(ok, dto.MovieToDto(movie))
}

func (m *Movies) StartWatch(ctx *gin.Context) {
	movieId, err := strconv.ParseUint(ctx.Param("id"), 10, 64)
	if err != nil {
		ctx.AbortWithStatusJSON(badReq, badReqErr)
		return
	}

	chunk, movieErr := m.usecase.StartWatch(ctx, movieId)
	if movieErr != nil {
		ctx.AbortWithStatusJSON(movieErr.ToHttpCode(), movieErr)
		return
	}

	ctx.JSON(ok, dto.ChunkToDto(chunk))
}

func (m *Movies) GetMovieChunck(ctx *gin.Context) {
	movieId, err := strconv.ParseUint(ctx.Param("id"), 10, 64)
	if err != nil {
		ctx.AbortWithStatusJSON(badReq, badReqErr)
		return
	}

	fileId, err := strconv.Atoi(ctx.Param("fileId"))
	if err != nil {
		ctx.AbortWithStatusJSON(badReq, badReqErr)
		return
	}

	chunkId, err := strconv.Atoi(ctx.Param("chunkId"))
	if err != nil {
		ctx.AbortWithStatusJSON(badReq, badReqErr)
		return
	}

	chunk, movieErr := m.usecase.GetMovieChunck(ctx, movieId, fileId, chunkId)
	if movieErr != nil {
		ctx.AbortWithStatusJSON(movieErr.ToHttpCode(), movieErr)
		return
	}

	ctx.JSON(ok, dto.ChunkToDto(chunk))
}