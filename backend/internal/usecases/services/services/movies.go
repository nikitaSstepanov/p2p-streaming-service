package services

import (
	"context"
	"fmt"
	"math"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/nikitaSstepanov/p2p-streaming-service/backend/internal/types/dto/movies"
	"github.com/nikitaSstepanov/p2p-streaming-service/backend/internal/types/statuses"
	"github.com/nikitaSstepanov/p2p-streaming-service/backend/internal/usecases/state"
	"github.com/nikitaSstepanov/p2p-streaming-service/backend/internal/usecases/storage"
	"github.com/nikitaSstepanov/p2p-streaming-service/backend/internal/usecases/storage/entities"
	"github.com/nikitaSstepanov/p2p-streaming-service/backend/pkg/bittorrent/decode"
	"github.com/nikitaSstepanov/p2p-streaming-service/backend/pkg/bittorrent/p2p"
	"github.com/redis/go-redis/v9"
)

type Movies struct {
	storage *storage.Storage
	redis   *redis.Client
	state   *state.State
}

func NewMovies(storage *storage.Storage, state *state.State, redis *redis.Client) *Movies {
	return &Movies{
		storage: storage,
		redis:   redis,
		state:   state,
	}
}

func (m *Movies) GetMovies(ctx context.Context, limit string, offset string) (*[]dto.MovieDto, string, string) {
	lim, err := strconv.Atoi(limit)
	if err != nil {
		return nil, statuses.BadRequest, "Limit must be integer."
	} else if (lim < 0 || lim > 50) {
		return nil, statuses.BadRequest, "Limit must be in range [0; 50]"
	}

	off, err := strconv.Atoi(offset)
	if err != nil {
		return nil, statuses.BadRequest, "Limit must be integer."
	} else if off < 0 {
		return nil, statuses.BadRequest, "Offset must be greater than zero."
	}
	
	movies, err := m.storage.Movies.GetAllMovies(ctx, limit, offset)

	if err != nil {
		return nil, statuses.InternalError, "Something going wrong...."
	}

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
	
	return &result, statuses.OK, ""
}

func (m *Movies) GetMovieById(ctx context.Context, movieId string) (*dto.MovieDto, string, string) {
	_, err := strconv.ParseUint(movieId, 10, 64)
	if err != nil {
		return nil, statuses.BadRequest, "Id must be integer."
	}

	movie, status, msg := m.findMovie(ctx, movieId)

	if status != statuses.OK {
		return nil, status, msg
	}

	result := dto.MovieDto{
		Id:   movie.Id,
		Name: movie.Name,
	}

	return &result, statuses.OK, ""
}

func (m *Movies) StartWatch(ctx context.Context, movieId string) (*dto.ChunkDto, string, string) {
	movie, status, msg := m.findMovie(ctx, movieId)

	if status != statuses.OK {
		return nil, status, msg
	}

	chunck, status, msg := m.getChunck(ctx, movie, 0, true, -1)

	if status != statuses.OK {
		return nil, status, msg
	}

	return chunck, statuses.OK, ""
}

func (m *Movies) GetMovieChunck(ctx context.Context, movieId string, fileId string, chunkId string) (*dto.ChunkDto, string, string) {
	flId, err := strconv.ParseInt(fileId, 10, 64)
	if err != nil {
		return nil, statuses.BadRequest, "fileId must be integer"
	}

	chnkId, err := strconv.Atoi(chunkId)
	if err != nil {
		return nil, statuses.BadRequest, "chunkId must be integer"
	}
	
	movie, status, msg := m.findMovie(ctx, movieId)

	if status != statuses.OK {
		return nil, status, msg
	}

	chunk, status, msg := m.getChunck(ctx, movie, chnkId, false, flId)

	if status != statuses.OK {
		return nil, status, msg
	}

	return chunk, statuses.OK, ""
}

func (m *Movies) findMovie(ctx context.Context, movieId string) (*entities.Movie, string, string) {
	var movie entities.Movie

	err := m.redis.Get(ctx, fmt.Sprintf("movies:%s", movieId)).Scan(&movie)

	if err != nil && err != redis.Nil {
		return nil, statuses.InternalError, "Something going wrong...."
	}

	if movie.Id == 0 {
		movie, err = m.storage.Movies.GetMovieById(ctx, movieId)

		if err != nil {
			return nil, statuses.InternalError, "Something going wrong...."
		}

		if movie.Id == 0 {
			return nil, statuses.NotFound, "This movie wasn`t found."
		}

		err = m.redis.Set(ctx, fmt.Sprintf("movies:%s", movieId), movie, 4 * time.Hour).Err()

		if err != nil {
			return nil, statuses.InternalError, "Something going wrong...."
		}
	}

	return &movie, statuses.OK, ""
}

func (m *Movies) getChunck(ctx context.Context, movie *entities.Movie, index int, changeExpires bool, fileId int64) (*dto.ChunkDto, string, string) {
	var result dto.ChunkDto

	(*m.state.Mutex).Lock()

	movieTorrent, isInMap := (*m.state.Movies)[movie.Id] 

	(*m.state.Mutex).Unlock()

	var piece p2p.Piece

	var err error

	if isInMap {
		if (changeExpires) {
			movieTorrent.Expires = time.Now().Local().Add(4 * time.Hour)

			(*m.state.Mutex).Lock()

			(*m.state.Movies)[movie.Id] = movieTorrent

			(*m.state.Mutex).Unlock()
		}

		if fileId != int64(movie.FileVersion) && fileId != -1 {
			adapter, status, msg := m.findAdapter(ctx, movie.Id, fileId)

			if status == statuses.InternalError {
				return nil, status, msg
			}

			if adapter.Id == 0 {
				return nil, statuses.NotFound, "This file wasn`t found."
			}
	
			index = int(math.Floor((float64(adapter.PieceLength) / float64(adapter.Length) * float64(index) * float64(100)) / (float64(movieTorrent.Torrent.PieceLength) / float64(movieTorrent.Torrent.Length) * float64(100))))
		}

		piece, err = p2p.Download(*movieTorrent.Torrent, index)

		if err != nil {
			return nil, statuses.InternalError, "Something going wrong...."
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
			return nil, statuses.InternalError, "Something going wrong...."
		}

		if firstIndex != 0 {
			paths = paths[firstIndex:]
			movie.Paths = strings.Join(paths, ";")
			movie.FileVersion += 1
			m.storage.Movies.UpdateMovie(ctx, movie)
			m.redis.Del(ctx, fmt.Sprintf("movies:%d", movie.Id))
			m.redis.Set(ctx, fmt.Sprintf("movies:%d", movie.Id), movie, 4 * time.Hour)
		}

		adapter, status, msg := m.findAdapter(ctx, movie.Id, int64(movie.FileVersion))

		if status == statuses.InternalError {
			return nil, status, msg
		}

		if adapter.Id == 0 {
			adapter = &entities.Adapter{
				MovieId:     movie.Id,
				Version:     movie.FileVersion,
				Length:      uint64(torrent.Length),
				PieceLength: uint64(torrent.PieceLength),
			}

			err = m.storage.Adapters.CreateAdapter(ctx, adapter)
			if err != nil {
				return nil, statuses.InternalError, "Something going wrong...."
			}

			err = m.redis.Set(ctx, fmt.Sprintf("adapters:%d:%d", movie.Id, movie.FileVersion), adapter, 4 * time.Hour).Err()
			if err != nil {
				return nil, statuses.InternalError, "Something going wrong...."
			}
		}

		var toSave state.Movie

		toSave.Torrent = &torrent
		toSave.Expires = time.Now().Local().Add(4 * time.Hour)

		(*m.state.Mutex).Lock()

		(*m.state.Movies)[movie.Id] = toSave

		(*m.state.Mutex).Unlock()

		if fileId != int64(movie.FileVersion) && fileId != -1 {
			adapter, err := m.storage.Adapters.GetAdapter(ctx, movie.Id, uint64(fileId))

			if err != nil {
				return nil, statuses.NotFound, "This file wasn`t found."
			}
	
			index = int(math.Floor((float64(adapter.PieceLength) / float64(adapter.Length) * float64(index) * float64(100)) / (float64(torrent.PieceLength) / float64(torrent.Length) * float64(100))))
		}

		piece, err = p2p.Download(torrent, index)

		if err != nil {
			return nil, statuses.InternalError, "Something going wrong."
		}
	}

	result = dto.ChunkDto{
		Buffer:      piece.Buff,
		NextIndex:   uint64(index + 1),
		FileVersion: movie.FileVersion,
	}
	
	return &result, statuses.OK, ""
}

func (m *Movies) findAdapter(ctx context.Context, movieId uint64, fileId int64) (*entities.Adapter, string, string) {
	var adapter entities.Adapter

	err := m.redis.Get(ctx, fmt.Sprintf("adapters:%d:%d", movieId, fileId)).Scan(&adapter)
	
	if err != nil && err != redis.Nil {
		return nil, statuses.InternalError, "Something going wrong..."
	}

	if adapter.Id == 0 {
		adapter, err = m.storage.Adapters.GetAdapter(ctx, movieId, uint64(fileId))

		if err != nil {
			return nil, statuses.InternalError, "Something going wrong..."
		}

		if adapter.Id == 0 {
			return &entities.Adapter{}, statuses.NotFound, "Adapter wasn`t found."
		}

		err = m.redis.Set(ctx, fmt.Sprintf("adapters:%d:%d", movieId, fileId), adapter, 4 * time.Hour).Err()
		if err != nil {
			return nil, statuses.InternalError, "Something going wrong..."
		}
	}

	return &adapter, statuses.OK, ""
}
