package services

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	//"github.com/nikitaSstepanov/p2p-streaming-service/backend/internal/pkg/dto"
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
	
	ctx.JSON(http.StatusOK, movies)
}

func (m *Movies) GetMovieById(ctx *gin.Context) {

}

func (m *Movies) GetMovieChunck(ctx *gin.Context) {

}

func (m *Movies) StartWatch(ctx *gin.Context) {
	movieId := ctx.Param("id")

	movie := m.Storage.Movies.GetMovieById(ctx, movieId)
	
	tf, err := decode.Open((movie.Path))
	
	if err != nil {
		fmt.Println(err)
		return 
	}

	torrent, err := tf.GetTorrentFile()

	if err != nil {
		ctx.JSON(http.StatusInternalServerError, "err")
		return 
	}

	index := 0
	piece, err := p2p.Download(torrent, index)

	if err != nil {
		ctx.JSON(http.StatusInternalServerError, "err")
		return 
	}

	fmt.Println(len(piece.Buff))

	ctx.JSON(http.StatusOK, &piece.Buff)
}