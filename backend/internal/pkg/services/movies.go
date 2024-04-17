package services

import (
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/nikitaSstepanov/p2p-streaming-service/backend/internal/pkg/dto/movies"
	"github.com/nikitaSstepanov/p2p-streaming-service/backend/internal/pkg/state"
	"github.com/nikitaSstepanov/p2p-streaming-service/backend/internal/pkg/storage"
	"github.com/nikitaSstepanov/p2p-streaming-service/backend/pkg/bittorrent/decode"
	"github.com/nikitaSstepanov/p2p-streaming-service/backend/pkg/bittorrent/p2p"
)

type Movies struct {
	Storage *storage.Storage
	State   *state.State
}

func NewMovies(storage *storage.Storage, state *state.State) *Movies {
	return &Movies{
		Storage: storage,
		State: state,
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

	var result dto.ChunkDto

	movieTorrent, isInMap := (*m.State.Movies)[movie.Id] 

	if isInMap {
		index, err := strconv.Atoi(chunkId)

		if err != nil {
			ctx.JSON(http.StatusBadRequest, "chunkId must be integer")
			return
		} 

		piece, err := p2p.Download(*movieTorrent.Torrent, index)

		if err != nil {
			ctx.JSON(http.StatusInternalServerError, "Something going wrong.....")
			return 
		}

		result = dto.ChunkDto{
			Buffer: piece.Buff,
			NextIndex: uint64(index) + 1,
		}
	} else {	
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
	
		result = dto.ChunkDto{
			Buffer: piece.Buff,
			NextIndex: uint64(index) + 1,
		}

		var toSave state.Movie

		toSave.Torrent = &torrent
		toSave.Expires = time.Now().Local().Add(4 * time.Hour)

		(*m.State.Movies)[movie.Id] = toSave
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

	var result dto.ChunkDto

	movieTorrent, isInMap := (*m.State.Movies)[movie.Id] 

	if isInMap {
		piece, err := p2p.Download(*movieTorrent.Torrent, 0)

		if err != nil {
			ctx.JSON(http.StatusInternalServerError, "Something going wrong.....")
			return 
		}

		result = dto.ChunkDto{
			Buffer: piece.Buff,
			NextIndex: 1,
		}

		movieTorrent.Expires = time.Now().Local().Add(4 * time.Hour)

		(*m.State.Movies)[movie.Id] = movieTorrent
	} else {
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

		result = dto.ChunkDto{
			Buffer: piece.Buff,
			NextIndex: 1,
		}

		var toSave state.Movie

		toSave.Torrent = &torrent
		toSave.Expires = time.Now().Local().Add(4 * time.Hour)

		(*m.State.Movies)[movie.Id] = toSave
	}

	ctx.JSON(http.StatusOK, result)
}