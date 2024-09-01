package playlist

import (
	"context"

	"github.com/nikitaSstepanov/p2p-streaming-service/backend/internal/entity"
	e "github.com/nikitaSstepanov/p2p-streaming-service/backend/pkg/errors"
)

var (
	forbiddenErr = e.New("Forbidden.", e.Forbidden)
)

type PlaylistStorage interface {
	GetPlaylists(ctx context.Context, userId uint64, limit int, offset int) ([]*entity.Playlist, *e.Error)
	GetPlaylist(ctx context.Context, id uint64) (*entity.Playlist, *e.Error)
	CreatePlaylist(ctx context.Context, playlist *entity.Playlist) *e.Error
	UpdatePlaylist(ctx context.Context, playlist *entity.Playlist) *e.Error
	AddMovie(ctx context.Context, playlistId uint64, movieId uint64) *e.Error
	RemoveMovie(ctx context.Context, playlistId uint64, movieId uint64) *e.Error
	DeletePlaylist(ctx context.Context, id uint64) *e.Error
}

type MovieStorage interface {
	GetMovieById(ctx context.Context, id uint64) (*entity.Movie, *e.Error)
}

type Playlist struct {
	playlistsStorage PlaylistStorage
	moviesStorage    MovieStorage
}

func New(playlists PlaylistStorage, movies MovieStorage) *Playlist {
	return &Playlist{
		playlistsStorage: playlists,
		moviesStorage:    movies,
	}
}

func (p *Playlist) GetPlaylists(ctx context.Context, id uint64, limit int, offset int) ([]*entity.Playlist, *e.Error) {
	return p.playlistsStorage.GetPlaylists(ctx, id, limit, offset)
}

func (p *Playlist) GetPlaylistById(ctx context.Context, userId uint64, id uint64) (*entity.Playlist, *e.Error) {
	playlist, err := p.playlistsStorage.GetPlaylist(ctx, id)
	if err != nil {
		return nil, err
	}

	if playlist.UserId != userId {
		return nil, forbiddenErr
	}

	return playlist, nil
}

func (p *Playlist) CreatePlaylist(ctx context.Context, playlist *entity.Playlist) *e.Error {
	return p.playlistsStorage.CreatePlaylist(ctx, playlist)
}

func (p *Playlist) EditPlaylist(ctx context.Context, updated *entity.Playlist, toAdd []uint64, toRemove []uint64) *e.Error {
	playlist, err := p.playlistsStorage.GetPlaylist(ctx, updated.Id)
	if err != nil {
		return err
	}

	if playlist.UserId != updated.UserId {
		return forbiddenErr
	}

	if updated.Title != "" {
		playlist.Title = updated.Title
	}

	if len(toAdd) != 0 {
		for i := 0; i < len(toAdd); i++ {
			movie, err := p.moviesStorage.GetMovieById(ctx, toAdd[i])
			if err != nil {
				err.Message += "Partially updated."
				return err
			}

			err = p.playlistsStorage.AddMovie(ctx, playlist.Id, movie.Id)
			if err != nil {
				err.Message += "Partially updated."
				return err
			}
		}
	}

	if len(toRemove) != 0 {
		for i := 0; i < len(toRemove); i++ {
			movie, err := p.moviesStorage.GetMovieById(ctx, toRemove[i])
			if err != nil {
				err.Message += "Partially updated."
				return err
			}

			err = p.playlistsStorage.RemoveMovie(ctx, playlist.Id, movie.Id)
			if err != nil {
				err.Message += "Partially updated."
				return err
			}
		}
	}

	return p.playlistsStorage.UpdatePlaylist(ctx, playlist)
}

func (p *Playlist) DeletePlaylist(ctx context.Context, userId uint64, id uint64) *e.Error {
	playlist, err := p.playlistsStorage.GetPlaylist(ctx, id)
	if err != nil {
		return err
	}

	if playlist.UserId != userId {
		return forbiddenErr
	}

	return p.playlistsStorage.DeletePlaylist(ctx, id)
}