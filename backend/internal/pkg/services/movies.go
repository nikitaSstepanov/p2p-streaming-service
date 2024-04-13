package services

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/nikitaSstepanov/p2p-streaming-service/backend/internal/pkg/dto/movies"
	"github.com/nikitaSstepanov/p2p-streaming-service/backend/internal/pkg/storage"
	"github.com/nikitaSstepanov/p2p-streaming-service/backend/pkg/bittorrent/decode"
	"github.com/nikitaSstepanov/p2p-streaming-service/backend/pkg/bittorrent/p2p"
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

	var result []dto.MovieDto

	if len(*movies) != 0 {
		for i := 0; i < len(*movies); i++ {
			movie := (*movies)[i]
			toAdd := dto.MovieDto{
				Id: movie.Id,
				Name: movie.Name,
			}
			result = append(result, toAdd)
		}
	} else {
		result = make([]dto.MovieDto, 0)
	}
	
	ctx.JSON(http.StatusOK, result)
}

func (m *Movies) GetMovieById(ctx *gin.Context) {
	movieId := ctx.Param("id")

	movie := m.Storage.Movies.GetMovieById(ctx, movieId)

	if movie.Id == 0 {
		ctx.JSON(http.StatusNotFound, "This moive wasn`t found")
		return
	}

	result := dto.MovieDto{
		Id: movie.Id,
		Name: movie.Name,
	}

	ctx.JSON(http.StatusOK, result)
}

func (m *Movies) GetMovieChunck(ctx *gin.Context) {
	movieId := ctx.Param("id")

	chunkId := ctx.Param("chunkId")

	movie := m.Storage.Movies.GetMovieById(ctx, movieId)

	if movie.Id == 0 {
		ctx.JSON(http.StatusNotFound, "This moive wasn`t found")
		return
	}
	
	tf, err := decode.Open(movie.Path)

	if err != nil {
		ctx.JSON(http.StatusInternalServerError, "Something going wrong.....")
		return 
	}

	torrent, err := tf.GetTorrentFile()

	if err != nil {
		ctx.JSON(http.StatusInternalServerError, "Something going wrong.....")
		return 
	}
	
	index, err := strconv.Atoi(chunkId)

	if err != nil {
		ctx.JSON(http.StatusBadRequest, "chunkId must be integer")
		return
	} 

	piece, err := p2p.Download(torrent, index)

	if err != nil {
		ctx.JSON(http.StatusInternalServerError, "Something going wrong.....")
		return 
	}

	result := dto.ChunkDto{
		Buffer: piece.Buff,
		NextIndex: uint64(index) + 1,
	}

	ctx.JSON(http.StatusOK, result)
}

func (m *Movies) StartWatch(ctx *gin.Context) {
	movieId := ctx.Param("id")

	movie := m.Storage.Movies.GetMovieById(ctx, movieId)

	if movie.Id == 0 {
		ctx.JSON(http.StatusNotFound, "This moive wasn`t found")
		return
	}
	
	tf, err := decode.Open(movie.Path)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, "Something going wrong.....")
		return 
	}

	torrent, err := tf.GetTorrentFile()

	if err != nil {
		ctx.JSON(http.StatusInternalServerError, "Something going wrong.....")
		return 
	}

	index := 0
	piece, err := p2p.Download(torrent, index)

	if err != nil {
		ctx.JSON(http.StatusInternalServerError, "Something going wrong.....")
		return 
	}

	result := dto.ChunkDto{
		Buffer: piece.Buff,
		NextIndex: 1,
	}

	ctx.JSON(http.StatusOK, result)
}