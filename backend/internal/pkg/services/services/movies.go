package services

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"
	"math"

	"github.com/gin-gonic/gin"
	"github.com/nikitaSstepanov/p2p-streaming-service/backend/internal/pkg/services/dto/movies"
	"github.com/nikitaSstepanov/p2p-streaming-service/backend/internal/pkg/state"
	"github.com/nikitaSstepanov/p2p-streaming-service/backend/internal/pkg/storage"
	"github.com/nikitaSstepanov/p2p-streaming-service/backend/internal/pkg/storage/entities"
	"github.com/nikitaSstepanov/p2p-streaming-service/backend/pkg/bittorrent/decode"
	"github.com/nikitaSstepanov/p2p-streaming-service/backend/pkg/bittorrent/p2p"
	"github.com/redis/go-redis/v9"
)

type Movies struct {
	Storage *storage.Storage
	Redis   *redis.Client
	State   *state.State
}

func NewMovies(storage *storage.Storage, state *state.State, redis *redis.Client) *Movies {
	return &Movies{
		Storage: storage,
		Redis: redis,
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

	movie, err := m.findMovie(ctx, movieId)

	if err != nil {
		ctx.JSON(http.StatusNotFound, "This moive wasn`t found")
		return
	}

	result := dto.MovieDto{
		Id: movie.Id,
		Name: movie.Name,
	}

	ctx.JSON(http.StatusOK, result)
}

func (m *Movies) StartWatch(ctx *gin.Context) {
	movieId := ctx.Param("id")

	movie, err := m.findMovie(ctx, movieId)

	if err != nil {
		ctx.JSON(http.StatusNotFound, "This moive wasn`t found")
		return
	}

	var result dto.ChunkDto

	(*m.State.Mutex).Lock()

	movieTorrent, isInMap := (*m.State.Movies)[movie.Id] 

	(*m.State.Mutex).Unlock()

	var piece p2p.Piece

	if isInMap {
		movieTorrent.Expires = time.Now().Local().Add(4 * time.Hour)

		(*m.State.Mutex).Lock()

		(*m.State.Movies)[movie.Id] = movieTorrent

		(*m.State.Mutex).Unlock()

		piece, err = p2p.Download(*movieTorrent.Torrent, 0)

		if err != nil {
			ctx.JSON(http.StatusInternalServerError, "Something going wrong.....")
			return 
		}
	} else {
		paths := strings.Split(movie.Paths, ";")

		var torrent decode.Torrent

		firstIndex := 0

		for i := 0; i < len(paths); i++ {
			path := paths[i]
			tf, err := decode.Open("files/" + path)

			if err != nil {
				os.Remove("files/" + path)
				continue
			}
			
			torrent, err = tf.GetTorrentFile()

			if err != nil {
				os.Remove("files/" + path)
				continue
			}

			firstIndex = i
			break
		}

		if torrent.Length == 0 {
			ctx.JSON(http.StatusInternalServerError, "Hm... Something going wrong.")
			return
		}

		if firstIndex != 0 {
			paths = paths[firstIndex:]
			movie.Paths = strings.Join(paths, ";")
			movie.FileVersion += 1
			m.Storage.Movies.UpdateMovie(ctx, movie)
			m.Redis.Del(ctx, movieId)
			m.Redis.Set(ctx, movieId, movie, 4 * time.Hour)
		}

		adapter := m.Storage.Adapters.GetAdapter(ctx, movieId, movie.FileVersion)

		if adapter.Id == 0 {
			adapter = &entities.Adapter{
				MovieId: movie.Id,
				Version: movie.FileVersion,
				Length: uint64(torrent.Length),
				PieceLength: uint64(torrent.PieceLength),
			}

			m.Storage.Adapters.CreateAdapter(ctx, adapter)
		}

		var toSave state.Movie

		toSave.Torrent = &torrent
		toSave.Expires = time.Now().Local().Add(4 * time.Hour)

		(*m.State.Mutex).Lock()

		(*m.State.Movies)[movie.Id] = toSave

		(*m.State.Mutex).Unlock()

		piece, err = p2p.Download(torrent, 0)

		if err != nil {
			ctx.JSON(http.StatusInternalServerError, "Something going wrong.....")
			return 
		}
	}

	result = dto.ChunkDto{
		Buffer: piece.Buff,
		NextIndex: 1,
	}

	ctx.JSON(http.StatusOK, result)
}

func (m *Movies) GetMovieChunck(ctx *gin.Context) {
	movieId := ctx.Param("id")

	movie, err := m.findMovie(ctx, movieId)

	if err != nil {
		ctx.JSON(http.StatusNotFound, "This moive wasn`t found")
		return
	}

	fileId, err := strconv.ParseUint(ctx.Param("fileId"), 10, 64)

	if err != nil {
		ctx.JSON(http.StatusBadRequest, "fileId must be integer")
		return
	}

	chunkId, err := strconv.Atoi(ctx.Param("chunkId"))

	if err != nil {
		ctx.JSON(http.StatusBadRequest, "chunkId must be integer")
		return
	}

	var index int

	var result dto.ChunkDto

	(*m.State.Mutex).Lock()

	movieTorrent, isInMap := (*m.State.Movies)[movie.Id]

	(*m.State.Mutex).Unlock()

	var piece p2p.Piece

	if isInMap {
		if fileId == movie.FileVersion {
			index = chunkId
		} else {
			adapter := m.Storage.Adapters.GetAdapter(ctx, movieId, fileId)
	
			index = int(math.Floor((float64(adapter.PieceLength) / float64(adapter.Length) * float64(chunkId) * float64(100)) / (float64(movieTorrent.Torrent.PieceLength) / float64(movieTorrent.Torrent.Length) * float64(100))))
		}

		piece, err = p2p.Download(*movieTorrent.Torrent, index)
	
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, "Something going wrong.....")
			return 
		}
	} else {	
		paths := strings.Split(movie.Paths, ";")

		var torrent decode.Torrent

		firstIndex := 0

		for i := 0; i < len(paths); i++ {
			path := paths[i]
			tf, err := decode.Open("files/" + path)

			if err != nil {
				os.Remove("files/" + path)
				continue
			}
			
			torrent, err = tf.GetTorrentFile()

			if err != nil {
				os.Remove("files/" + path)
				continue
			}

			firstIndex = i
			break
		}

		if torrent.Length == 0 {
			ctx.JSON(http.StatusInternalServerError, "Hm... Something going wrong.")
			return
		}

		if firstIndex != 0 {
			paths = paths[firstIndex:]
			movie.Paths = strings.Join(paths, ";")
			movie.FileVersion += 1
			m.Storage.Movies.UpdateMovie(ctx, movie)
			m.Redis.Del(ctx, movieId)
			m.Redis.Set(ctx, movieId, movie, 4 * time.Hour)
		}

		adapter := m.Storage.Adapters.GetAdapter(ctx, movieId, movie.FileVersion)

		if adapter.Id == 0 {
			adapter = &entities.Adapter{
				MovieId: movie.Id,
				Version: movie.FileVersion,
				Length: uint64(torrent.Length),
				PieceLength: uint64(torrent.PieceLength),
			}

			m.Storage.Adapters.CreateAdapter(ctx, adapter)
		}

		var toSave state.Movie

		toSave.Torrent = &torrent
		toSave.Expires = time.Now().Local().Add(4 * time.Hour)

		(*m.State.Mutex).Lock()

		(*m.State.Movies)[movie.Id] = toSave

		(*m.State.Mutex).Unlock()

		if fileId == movie.FileVersion {
			index = chunkId
		} else {
			adapter = m.Storage.Adapters.GetAdapter(ctx, movieId, fileId)
	
			index = int(math.Floor((float64(adapter.PieceLength) / float64(adapter.Length) * float64(chunkId) * float64(100)) / (float64(torrent.PieceLength) / float64(torrent.Length) * float64(100))))
		}
	
		piece, err = p2p.Download(torrent, index)
	
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, "Something going wrong.....")
			return 
		}
	}

	result = dto.ChunkDto{
		Buffer: piece.Buff,
		NextIndex: uint64(index) + 1,
	}

	ctx.JSON(http.StatusOK, result)
}

func (m *Movies) findMovie(ctx context.Context, movieId string) (*entities.Movie, error) {
	var movie entities.Movie

	m.Redis.Get(ctx, movieId).Scan(&movie)

	if movie.Id == 0 {
		movie = *m.Storage.Movies.GetMovieById(ctx, movieId)

		if movie.Id == 0 {
			return nil, fmt.Errorf("404")
		}

		m.Redis.Set(ctx, movieId, movie, 4 * time.Hour)
	}

	return &movie, nil
}