package services

import (
	"context"
	"strconv"
	"time"
	"fmt"

	"github.com/nikitaSstepanov/p2p-streaming-service/backend/internal/usecases/storage/entities"
	"github.com/nikitaSstepanov/p2p-streaming-service/backend/internal/types/dto/playlists"
	"github.com/nikitaSstepanov/p2p-streaming-service/backend/internal/usecases/storage"
	"github.com/nikitaSstepanov/p2p-streaming-service/backend/internal/types/responses"
	"github.com/nikitaSstepanov/p2p-streaming-service/backend/internal/types/statuses"
	"github.com/redis/go-redis/v9"
)

type Playlists struct {
	storage *storage.Storage
	redis   *redis.Client
	account *Account
}

func NewPlaylists(storage *storage.Storage, redis *redis.Client, account *Account) *Playlists {
	return &Playlists{
		storage: storage,
		redis:   redis,
		account: account,
	}
}

func (p *Playlists) GetPlaylists(ctx context.Context, header string, limit string, offset string) (*[]dto.PlaylistForList, string) {
	user, status := p.account.getUser(ctx, header)

	if status != statuses.OK {
		return nil, status
	}

	playlists, err := p.storage.Playlists.GetPlaylists(ctx, user.Id, limit, offset)

	if err != nil {
		return nil, statuses.InternalError
	}

	var result []dto.PlaylistForList

	if len(*playlists) != 0 {
		for i := 0; i < len(*playlists); i++ {
			playlist := (*playlists)[i]
			toAdd := dto.PlaylistForList{
				Id:   playlist.Id,
				Title: playlist.Title,
			}
			result = append(result, toAdd)
		}
	} else {
		result = make([]dto.PlaylistForList, 0)
	}

	return &result, statuses.OK
}

func (p *Playlists) GetPlaylistById(ctx context.Context, header string, id string) (*dto.Playlist, string) {
	user, status := p.account.getUser(ctx, header)

	if status != statuses.OK {
		return nil, status
	}

	playlist, err := p.findPlaylist(ctx, id)

	if err != nil {
		return nil, statuses.InternalError
	}

	if playlist.Id == 0 {
		return nil, statuses.NotFound
	}
	
	if user.Id != playlist.UserId {
		return nil, statuses.Forbidden
	}

	var movies []uint64

	if playlist.MoviesIds != nil && len(playlist.MoviesIds) != 0 {
		movies = playlist.MoviesIds
	} else {
		movies = make([]uint64, 0)
	}

	result := &dto.Playlist{
		Id: playlist.Id,
		Title: playlist.Title,
		Movies: movies,
	}

	return result, statuses.OK
}

func (p *Playlists) CreatePlaylist(ctx context.Context, header string, body *dto.CreatePlaylistDto) (*responses.Message, string) {
	user, status := p.account.getUser(ctx, header)

	if status != statuses.OK {
		return nil, status
	}

	playlist := &entities.Playlist{
		UserId: user.Id,
		Title: body.Title,
	}

	p.storage.Playlists.CreatePlaylist(ctx, playlist)

	result := &responses.Message{
		Message: "Created.",
	}

	return result, statuses.OK
}

func (p *Playlists) EditPlaylist(ctx context.Context, header string, id string, body *dto.UpdatePlaylistDto) (*responses.Message, string) {
	user, status := p.account.getUser(ctx, header)

	if status != statuses.OK {
		return nil, status
	}

	playlist, err := p.findPlaylist(ctx, id)

	if err != nil {
		return nil, statuses.InternalError
	}

	if playlist.Id == 0 {
		return nil, statuses.NotFound
	}

	if user.Id != playlist.UserId {
		return nil, statuses.NotFound
	}

	if body.Title != "" && len(body.Title) != 0 {
		playlist.Title = body.Title

		p.storage.Playlists.UpdatePlaylist(ctx, *playlist)
		p.redis.Del(ctx, fmt.Sprintf("playlists:%s", id))
		p.redis.Set(ctx, fmt.Sprintf("playlists:%s", id), playlist, 1 * time.Hour)
	}

	result := &responses.Message{}

	if body.MoviesToAdd != nil && len(body.MoviesToAdd) != 0 {
		for i := 0; i < len(body.MoviesToAdd); i++ {
			movie, err := p.storage.Movies.GetMovieById(ctx, strconv.FormatUint(body.MoviesToAdd[i], 10))
			
			if err != nil {
				return nil, statuses.InternalError
			}

			if movie.Id == 0 {
				result.Message = "Partially updated."
				return result, statuses.NotFound
			}

			p.storage.Playlists.AddMovie(ctx, playlist.Id, movie.Id)

			p.redis.Del(ctx, fmt.Sprintf("playlists:%s", id))
		}
	}

	if body.MoviesToRemove != nil && len(body.MoviesToRemove) != 0 {
		for i := 0; i < len(body.MoviesToRemove); i++ {
			movie, err := p.storage.Movies.GetMovieById(ctx, strconv.FormatUint(body.MoviesToRemove[i], 10))

			if err != nil {
				return nil, statuses.InternalError
			}

			if movie.Id == 0 {
				continue
			}

			p.storage.Playlists.RemoveMovie(ctx, playlist.Id, movie.Id)

			p.redis.Del(ctx, fmt.Sprintf("playlists:%s", id))
		}
	}

	result.Message = "Updated."

	return result, statuses.OK
}

func (p *Playlists) DeletePlaylist(ctx context.Context, header string, id string) (*responses.Message, string) {
	user, status := p.account.getUser(ctx, header)

	if status != statuses.OK {
		return nil, status
	}

	playlist, err := p.findPlaylist(ctx, id)

	if err != nil {
		return nil, statuses.InternalError
	}

	if playlist.Id == 0 {
		return nil, statuses.NotFound
	}

	if user.Id != playlist.UserId {
		return nil, statuses.Forbidden
	}

	p.storage.Playlists.DeletePlaylist(ctx, id)
	p.redis.Del(ctx, fmt.Sprintf("playlists:%s", id))

	result := &responses.Message{
		Message: "No content.",
	}

	return result, statuses.OK
}

func (p *Playlists) findPlaylist(ctx context.Context, id string) (*entities.Playlist, error) {
	var playlist *entities.Playlist

	err := p.redis.Get(ctx, fmt.Sprintf("playlists:%s", id)).Scan(&playlist)

	if err != nil {
		return nil, err
	}

	if playlist.Id == 0 {
		playlist, err = p.storage.Playlists.GetPlaylist(ctx, id)

		if err != nil {
			return nil, err
		}

		err = p.redis.Set(ctx, fmt.Sprintf("playlists:%s", id), playlist, 1 * time.Hour).Err()

		if err != nil {
			return nil, err
		}
	}

	return playlist, nil
}
