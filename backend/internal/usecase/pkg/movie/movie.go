package movie

import (
	"context"
	"math"
	"os"
	"strings"
	"time"

	"github.com/nikitaSstepanov/p2p-streaming-service/backend/internal/entity"
	"github.com/nikitaSstepanov/p2p-streaming-service/backend/pkg/bittorrent/decode"
	"github.com/nikitaSstepanov/p2p-streaming-service/backend/pkg/bittorrent/p2p"
	e "github.com/nikitaSstepanov/p2p-streaming-service/backend/pkg/errors"
)

const (
	expires = 4 * time.Hour
)

var (
	internalErr = e.New("Something going wrong...", e.Internal)
)

type State interface {
	Get(id uint64) *decode.Torrent
	Add(id uint64, torrent *decode.Torrent, expires time.Duration)
	ChangeExpires(id uint64, expires time.Duration)
}

type MovieStorage interface {
	GetAllMovies(ctx context.Context, limit int, offset int) ([]*entity.Movie, *e.Error)
	GetMovieById(ctx context.Context, id uint64) (*entity.Movie, *e.Error)
	CreateMovie(ctx context.Context, movie *entity.Movie) *e.Error
	UpdateMovie(ctx context.Context, movie *entity.Movie) *e.Error
}

type AdapterStorage interface {
	GetAdapter(ctx context.Context, movieId uint64, version int) (*entity.Adapter, *e.Error)
	CreateAdapter(ctx context.Context, adapter *entity.Adapter) *e.Error
}

type Movie struct {
	movies   MovieStorage
	adapters AdapterStorage
	state    State
}

func New(movies MovieStorage, adapters AdapterStorage, state State) *Movie {
	return &Movie{
		movies,
		adapters,
		state,
	}
}

func (m *Movie) GetMovies(ctx context.Context, limit int, offset int) ([]*entity.Movie, *e.Error) {
	return m.movies.GetAllMovies(ctx, limit, offset)
}

func (m *Movie) GetMovieById(ctx context.Context, id uint64) (*entity.Movie, *e.Error) {
	return m.movies.GetMovieById(ctx, id)
}

func (m *Movie) StartWatch(ctx context.Context, movieId uint64) (*entity.Chunk, *e.Error) {
	movie, err := m.movies.GetMovieById(ctx, movieId)
	if err != nil {
		return nil, err
	}

	torrent := m.state.Get(movieId)

	if torrent != nil {
		m.state.ChangeExpires(movieId, expires)
	} else {
		torrent, err = m.openTorrent(ctx, movie)
		if err != nil {
			return nil, err
		}

		m.state.Add(movieId, torrent, expires)
	}

	piece, getChunkErr := p2p.Download(*torrent, 0)
	if getChunkErr != nil {
		return nil, internalErr
	}

	chunk := &entity.Chunk{
		Buffer:      piece.Buff,
		NextIndex:   1,
		FileVersion: movie.FileVersion,
	}

	return chunk, nil
}

func (m *Movie) GetMovieChunck(ctx context.Context, movieId uint64, fileId int, index int) (*entity.Chunk, *e.Error) {
	movie, err := m.movies.GetMovieById(ctx, movieId)
	if err != nil {
		return nil, err
	}

	torrent := m.state.Get(movieId)

	if torrent == nil {
		torrent, err = m.openTorrent(ctx, movie)
		if err != nil {
			return nil, err
		}

		m.state.Add(movieId, torrent, expires)
	}

	if fileId != movie.FileVersion {
		adapter, err := m.adapters.GetAdapter(ctx, movieId, fileId)
		if err != nil {
			return nil, err
		}

		index = countIndex(index, adapter, torrent)
	}

	piece, getChunkErr := p2p.Download(*torrent, index)
	if getChunkErr != nil {
		return nil, internalErr
	}

	chunk := &entity.Chunk{
		Buffer:      piece.Buff,
		NextIndex:   index + 1,
		FileVersion: movie.FileVersion,
	}

	return chunk, nil
}

func (m *Movie) openTorrent(ctx context.Context, movie *entity.Movie) (*decode.Torrent, *e.Error) {
	var torrent decode.Torrent

	firstIndex := 0

	paths := strings.Split(movie.Paths, ";")

	for i := 0; i < len(paths); i++ {
		path := paths[i]

		tf, err := decode.Open("files/" + path)
		if err != nil {
			if err = os.Remove("files/" + path); err != nil {
				return nil, internalErr
			}

			continue
		}
		
		torrent, err = tf.GetTorrentFile()
		if err != nil {
			if err = os.Remove("files/" + path); err != nil {
				return nil, internalErr
			}

			continue
		}

		firstIndex = i
		break
	}

	if torrent.Length == 0 {
		return nil, internalErr
	}

	if firstIndex != 0 {
		movie.Paths = strings.Join(paths[firstIndex:], ";")
		movie.FileVersion += 1

		err := m.movies.UpdateMovie(ctx, movie)
		if err != nil {
			return nil, err
		}
	}

	adapter, err := m.adapters.GetAdapter(ctx, movie.Id, movie.FileVersion)
	if err != nil && err.Code != e.NotFound {
		return nil, err
	}

	if adapter.Id == 0 {
		new := &entity.Adapter{
			MovieId:     movie.Id,
			Version:     movie.FileVersion,
			Length:      torrent.Length,
			PieceLength: torrent.PieceLength,
		}

		err = m.adapters.CreateAdapter(ctx, new)
		if err != nil {
			return nil, err
		}
	}

	return &torrent, nil
}

func countIndex(index int, adapter *entity.Adapter, torrent *decode.Torrent) int {
	return int(math.Floor((float64(adapter.PieceLength) / float64(adapter.Length) * float64(index) * float64(100)) / (float64(torrent.PieceLength) / float64(torrent.Length) * float64(100))))
}