package services

import (
	"net/http"
	"strconv"
	"strings"
	"context"
	"time"
	"fmt"

	"github.com/nikitaSstepanov/p2p-streaming-service/backend/internal/pkg/services/dto/playlists"
	"github.com/nikitaSstepanov/p2p-streaming-service/backend/internal/pkg/storage/entities"
	"github.com/nikitaSstepanov/p2p-streaming-service/backend/internal/pkg/storage"
	"github.com/redis/go-redis/v9"
	"github.com/gin-gonic/gin"
)

type Playlists struct {
	Storage *storage.Storage
	Redis   *redis.Client
	Auth    *Auth
}

func NewPlaylists(storage *storage.Storage, redis *redis.Client, auth *Auth) *Playlists {
	return &Playlists{
		Storage: storage,
		Redis: redis,
		Auth: auth,
	}
}

func (p *Playlists) GetPlaylists(ctx *gin.Context) {
	user := p.getUser(ctx)

	if user.Id == 0 {
		return
	}

	limit := ctx.DefaultQuery("limit", "5")

	offset := ctx.DefaultQuery("offset", "0")

	playlists := p.Storage.Playlists.GetPlaylists(ctx, user.Id, limit, offset)

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

	ctx.JSON(http.StatusOK, result)
}

func (p *Playlists) GetPlaylistById(ctx *gin.Context) {
	user := p.getUser(ctx)

	if user.Id == 0 {
		return
	}

	id := ctx.Param("id")

	playlist := p.findPlaylist(ctx, id)

	if playlist.Id == 0 {
		ctx.JSON(http.StatusNotFound, "This playlist wasn`t found.")
		return
	}
	
	if user.Id != playlist.UserId {
		ctx.JSON(http.StatusForbidden, "Forbidden.")
		return
	}

	result := dto.Playlist{
		Id: playlist.Id,
		Title: playlist.Title,
		Movies: playlist.MoviesIds,
	}

	ctx.JSON(http.StatusOK, result)
}

func (p *Playlists) CreatePlaylist(ctx *gin.Context) {
	user := p.getUser(ctx)

	if user.Id == 0 {
		return
	}

	var body dto.CreatePlaylistDto

	if err := ctx.ShouldBindJSON(&body); err != nil {
		ctx.JSON(http.StatusBadRequest, "Incorrect data.")
		return
	}

	playlist := entities.Playlist{
		UserId: user.Id,
		Title: body.Title,
	}

	p.Storage.Playlists.CreatePlaylist(ctx, playlist)

	ctx.JSON(http.StatusCreated, "Created.")
}

func (p *Playlists) EditPlaylist(ctx *gin.Context) {
	user := p.getUser(ctx)

	if user.Id == 0 {
		return
	}

	id := ctx.Param("id")

	playlist := p.findPlaylist(ctx, id)

	if playlist.Id == 0 {
		ctx.JSON(http.StatusNotFound, "This playlist wasn`t found.")
		return
	}

	if user.Id != playlist.UserId {
		ctx.JSON(http.StatusForbidden, "Forbidden.")
		return
	}

	var body dto.UpdatePlaylistDto

	if err := ctx.ShouldBindJSON(&body); err != nil {
		ctx.JSON(http.StatusBadRequest, "Incorrect data.")
		return
	}

	if body.Title != "" && len(body.Title) != 0 {
		playlist.Title = body.Title

		p.Storage.Playlists.UpdatePlaylist(ctx, *playlist)
		p.Redis.Del(ctx, fmt.Sprintf("playlists:%s", id))
		p.Redis.Set(ctx, fmt.Sprintf("playlists:%s", id), playlist, 1 * time.Hour)
	}

	if body.MoviesToAdd != nil && len(body.MoviesToAdd) != 0 {
		for i := 0; i < len(body.MoviesToAdd); i++ {
			movie := p.Storage.Movies.GetMovieById(ctx, strconv.FormatUint(body.MoviesToAdd[i], 10))

			if movie.Id == 0 {
				ctx.JSON(http.StatusNotFound, fmt.Sprintf("Movie %d wasn`t found", body.MoviesToAdd[i]))
				return
			}

			p.Storage.Playlists.AddMovie(ctx, playlist.Id, movie.Id)
		}
	}

	if body.MoviesToRemove != nil && len(body.MoviesToRemove) != 0 {
		for i := 0; i < len(body.MoviesToRemove); i++ {
			movie := p.Storage.Movies.GetMovieById(ctx, strconv.FormatUint(body.MoviesToRemove[i], 10))

			if movie.Id == 0 {
				continue
			}

			p.Storage.Playlists.RemoveMovie(ctx, playlist.Id, movie.Id)
		}
	}

	ctx.JSON(http.StatusOK, "Updated.")
}

func (p *Playlists) DeletePlaylist(ctx *gin.Context) {
	user := p.getUser(ctx)

	if user.Id == 0 {
		return
	}

	id := ctx.Param("id")

	playlist := p.findPlaylist(ctx, id)

	if playlist.Id == 0 {
		ctx.JSON(http.StatusNotFound, "This playlist wasn`t found.")
		return
	}

	if user.Id != playlist.UserId {
		ctx.JSON(http.StatusForbidden, "Forbidden.")
		return
	}

	p.Storage.Playlists.DeletePlaylist(ctx, id)
	p.Redis.Del(ctx, fmt.Sprintf("playlists:%s", id))

	ctx.JSON(http.StatusNoContent, "No content.")
}

func (p *Playlists) findPlaylist(ctx context.Context, id string) *entities.Playlist {
	var playlist entities.Playlist

	p.Redis.Get(ctx, fmt.Sprintf("playlists:%s", id)).Scan(&playlist)

	if playlist.Id == 0 {
		playlist = *p.Storage.Playlists.GetPlaylist(ctx, id)

		p.Redis.Set(ctx, fmt.Sprintf("playlists:%s", id), playlist, 1 * time.Hour)
	}

	return &playlist
}

func (p *Playlists) getUser(ctx *gin.Context) *entities.User {
	header := strings.Split(ctx.GetHeader("Authorization"), " ")
	bearer := header[0]
	token := header[1]

	if bearer != "Bearer" {
		ctx.JSON(http.StatusUnauthorized, "Incorrect token.")
		return nil
	}

	claims, err := p.Auth.ValidateToken(token)

	if err != nil {
		ctx.JSON(http.StatusUnauthorized, "Incorrecct token.")
		return nil
	}

	user := p.findUser(ctx, claims.Username)

	if user.Id == 0 {
		ctx.JSON(http.StatusUnauthorized, "Incorrecct token.")
		return nil
	}

	return &user
}

func (p *Playlists) findUser(ctx context.Context, username string) entities.User {
	var user entities.User

	p.Redis.Get(ctx, fmt.Sprintf("users:%s", username)).Scan(&user)

	if user.Id == 0 {
		user = *(p.Storage.Users.GetUser(ctx, username))

		if user.Id == 0 {
			return entities.User{}
		}

		p.Redis.Set(ctx, fmt.Sprintf("users:%s", username), user, 1 * time.Hour)
	}

	return user
}