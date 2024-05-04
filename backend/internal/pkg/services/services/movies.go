package services

import (
	"net/http"
	"context"
	"strings"
	"strconv"
	"math"
	"time"
	"fmt"
	"os"

	"github.com/nikitaSstepanov/p2p-streaming-service/backend/internal/pkg/services/dto/movies"
	"github.com/nikitaSstepanov/p2p-streaming-service/backend/internal/pkg/storage/entities"
	"github.com/nikitaSstepanov/p2p-streaming-service/backend/pkg/bittorrent/decode"
	"github.com/nikitaSstepanov/p2p-streaming-service/backend/internal/pkg/storage"
	"github.com/nikitaSstepanov/p2p-streaming-service/backend/pkg/bittorrent/p2p"
	"github.com/nikitaSstepanov/p2p-streaming-service/backend/internal/pkg/state"
	"github.com/redis/go-redis/v9"
	"github.com/gin-gonic/gin"
)

type Movies struct {
	Storage *storage.Storage
	Redis   *redis.Client
	State   *state.State
}

func NewMovies(storage *storage.Storage, state *state.State, redis *redis.Client) *Movies {
	return &Movies{
		Storage: storage,
		Redis:   redis,
		State:   state,
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
				Id:   movie.Id,
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
		Id:   movie.Id,
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

	chunck, err := m.getChunck(ctx, movie, 0, true, -1)

	if err != nil {
		ctx.JSON(http.StatusInternalServerError, "Something going wrong....")
		return
	}

	ctx.JSON(http.StatusOK, chunck)
}

func (m *Movies) GetMovieChunck(ctx *gin.Context) {
	movieId := ctx.Param("id")

	movie, err := m.findMovie(ctx, movieId)

	if err != nil {
		ctx.JSON(http.StatusNotFound, "This moive wasn`t found")
		return
	}

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

	chunk, err := m.getChunck(ctx, movie, chunkId, false, fileId)

	if err != nil {
		ctx.JSON(http.StatusInternalServerError, "Something going wrong....")
		return
	}

	ctx.JSON(http.StatusOK, chunk)
}

func (m *Movies) findMovie(ctx context.Context, movieId string) (*entities.Movie, error) {
	var movie entities.Movie

	m.Redis.Get(ctx, fmt.Sprintf("movies:%s", movieId)).Scan(&movie)

	if movie.Id == 0 {
		movie = *m.Storage.Movies.GetMovieById(ctx, movieId)

		if movie.Id == 0 {
			return nil, fmt.Errorf("404")
		}

		m.Redis.Set(ctx, fmt.Sprintf("movies:%s", movieId), movie, 4 * time.Hour)
	}

	return &movie, nil
}

func (m *Movies) getChunck(ctx context.Context, movie *entities.Movie, index int, changeExpires bool, fileId int64) (*dto.ChunkDto, error) {
	var result dto.ChunkDto

	(*m.State.Mutex).Lock()

	movieTorrent, isInMap := (*m.State.Movies)[movie.Id] 

	(*m.State.Mutex).Unlock()

	var piece p2p.Piece

	var err error

	if isInMap {
		if (changeExpires) {
			movieTorrent.Expires = time.Now().Local().Add(4 * time.Hour)

			(*m.State.Mutex).Lock()

			(*m.State.Movies)[movie.Id] = movieTorrent

			(*m.State.Mutex).Unlock()
		}

		if fileId != int64(movie.FileVersion) && fileId != -1 {
			adapter := m.findAdapter(ctx, movie.Id, fileId)

			if adapter.Id == 0 {
				return nil, fmt.Errorf("500")
			}
	
			index = int(math.Floor((float64(adapter.PieceLength) / float64(adapter.Length) * float64(index) * float64(100)) / (float64(movieTorrent.Torrent.PieceLength) / float64(movieTorrent.Torrent.Length) * float64(100))))
		}

		piece, err = p2p.Download(*movieTorrent.Torrent, index)

		if err != nil {
			return nil, fmt.Errorf("500")
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
			return nil, fmt.Errorf("500")
		}

		if firstIndex != 0 {
			paths = paths[firstIndex:]
			movie.Paths = strings.Join(paths, ";")
			movie.FileVersion += 1
			m.Storage.Movies.UpdateMovie(ctx, movie)
			m.Redis.Del(ctx, fmt.Sprintf("movies:%d", movie.Id))
			m.Redis.Set(ctx, fmt.Sprintf("movies:%d", movie.Id), movie, 4 * time.Hour)
		}

		adapter := m.findAdapter(ctx, movie.Id, int64(movie.FileVersion))

		if adapter.Id == 0 {
			adapter = &entities.Adapter{
				MovieId:     movie.Id,
				Version:     movie.FileVersion,
				Length:      uint64(torrent.Length),
				PieceLength: uint64(torrent.PieceLength),
			}

			m.Storage.Adapters.CreateAdapter(ctx, adapter)
			m.Redis.Set(ctx, fmt.Sprintf("adapters:%d:%d", movie.Id, movie.FileVersion), adapter, 4 * time.Hour)
		}

		var toSave state.Movie

		toSave.Torrent = &torrent
		toSave.Expires = time.Now().Local().Add(4 * time.Hour)

		(*m.State.Mutex).Lock()

		(*m.State.Movies)[movie.Id] = toSave

		(*m.State.Mutex).Unlock()

		if fileId != int64(movie.FileVersion) && fileId != -1 {
			adapter := m.Storage.Adapters.GetAdapter(ctx, movie.Id, uint64(fileId))
	
			index = int(math.Floor((float64(adapter.PieceLength) / float64(adapter.Length) * float64(index) * float64(100)) / (float64(torrent.PieceLength) / float64(torrent.Length) * float64(100))))
		}

		piece, err = p2p.Download(torrent, index)

		if err != nil {
			return nil, fmt.Errorf("500")
		}
	}

	result = dto.ChunkDto{
		Buffer:      piece.Buff,
		NextIndex:   1,
		FileVersion: movie.FileVersion,
	}
	return &result, nil
}

func (m *Movies) findAdapter(ctx context.Context, movieId uint64, fileId int64) *entities.Adapter {
	var adapter entities.Adapter

	m.Redis.Get(ctx, fmt.Sprintf("adapters:%d:%d", movieId, fileId)).Scan(&adapter)

	if adapter.Id == 0 {
		adapter = *m.Storage.Adapters.GetAdapter(ctx, movieId, uint64(fileId))

		if adapter.Id == 0 {
			return &entities.Adapter{}
		}

		m.Redis.Set(ctx, fmt.Sprintf("adapters:%d:%d", movieId, fileId), adapter, 4 * time.Hour)
	}

	return &adapter
}
