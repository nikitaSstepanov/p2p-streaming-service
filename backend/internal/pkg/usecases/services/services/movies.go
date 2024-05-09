package services

import (
	"context"
	"fmt"
	"math"
	"os"
	"strings"
	"time"

	"github.com/nikitaSstepanov/p2p-streaming-service/backend/internal/pkg/types/dto/movies"
	"github.com/nikitaSstepanov/p2p-streaming-service/backend/internal/pkg/types/statuses"
	"github.com/nikitaSstepanov/p2p-streaming-service/backend/internal/pkg/usecases/state"
	"github.com/nikitaSstepanov/p2p-streaming-service/backend/internal/pkg/usecases/storage"
	"github.com/nikitaSstepanov/p2p-streaming-service/backend/internal/pkg/usecases/storage/entities"
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

func (m *Movies) GetMovies(ctx context.Context, limit string, offset string) (*[]dto.MovieDto, string) {
	movies := m.storage.Movies.GetAllMovies(ctx, limit, offset)

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
	
	return &result, statuses.OK
}

func (m *Movies) GetMovieById(ctx context.Context, movieId string) (*dto.MovieDto, string) {

	movie, err := m.findMovie(ctx, movieId)

	if err != nil {
		return nil, statuses.NotFound
	}

	result := dto.MovieDto{
		Id:   movie.Id,
		Name: movie.Name,
	}

	return &result, statuses.OK
}

func (m *Movies) StartWatch(ctx context.Context, movieId string) (*dto.ChunkDto, string) {
	movie, err := m.findMovie(ctx, movieId)

	if err != nil {
		return nil, statuses.NotFound
	}

	chunck, err := m.getChunck(ctx, movie, 0, true, -1)

	if err != nil {
		return nil, statuses.InternalError
	}

	return chunck, statuses.OK
}

func (m *Movies) GetMovieChunck(ctx context.Context, movieId string, fileId int64, chunkId int) (*dto.ChunkDto, string) {
	movie, err := m.findMovie(ctx, movieId)

	if err != nil {
		return nil, statuses.NotFound
	}

	chunk, err := m.getChunck(ctx, movie, chunkId, false, fileId)

	if err != nil {
		return nil, statuses.InternalError
	}

	return chunk, statuses.OK
}

func (m *Movies) findMovie(ctx context.Context, movieId string) (*entities.Movie, error) {
	var movie entities.Movie

	m.redis.Get(ctx, fmt.Sprintf("movies:%s", movieId)).Scan(&movie)

	if movie.Id == 0 {
		movie = *m.storage.Movies.GetMovieById(ctx, movieId)

		if movie.Id == 0 {
			return nil, fmt.Errorf("404")
		}

		m.redis.Set(ctx, fmt.Sprintf("movies:%s", movieId), movie, 4 * time.Hour)
	}

	return &movie, nil
}

func (m *Movies) getChunck(ctx context.Context, movie *entities.Movie, index int, changeExpires bool, fileId int64) (*dto.ChunkDto, error) {
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
			m.storage.Movies.UpdateMovie(ctx, movie)
			m.redis.Del(ctx, fmt.Sprintf("movies:%d", movie.Id))
			m.redis.Set(ctx, fmt.Sprintf("movies:%d", movie.Id), movie, 4 * time.Hour)
		}

		adapter := m.findAdapter(ctx, movie.Id, int64(movie.FileVersion))

		if adapter.Id == 0 {
			adapter = &entities.Adapter{
				MovieId:     movie.Id,
				Version:     movie.FileVersion,
				Length:      uint64(torrent.Length),
				PieceLength: uint64(torrent.PieceLength),
			}

			m.storage.Adapters.CreateAdapter(ctx, adapter)
			m.redis.Set(ctx, fmt.Sprintf("adapters:%d:%d", movie.Id, movie.FileVersion), adapter, 4 * time.Hour)
		}

		var toSave state.Movie

		toSave.Torrent = &torrent
		toSave.Expires = time.Now().Local().Add(4 * time.Hour)

		(*m.state.Mutex).Lock()

		(*m.state.Movies)[movie.Id] = toSave

		(*m.state.Mutex).Unlock()

		if fileId != int64(movie.FileVersion) && fileId != -1 {
			adapter := m.storage.Adapters.GetAdapter(ctx, movie.Id, uint64(fileId))
	
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

	m.redis.Get(ctx, fmt.Sprintf("adapters:%d:%d", movieId, fileId)).Scan(&adapter)

	if adapter.Id == 0 {
		adapter = *m.storage.Adapters.GetAdapter(ctx, movieId, uint64(fileId))

		if adapter.Id == 0 {
			return &entities.Adapter{}
		}

		m.redis.Set(ctx, fmt.Sprintf("adapters:%d:%d", movieId, fileId), adapter, 4 * time.Hour)
	}

	return &adapter
}
