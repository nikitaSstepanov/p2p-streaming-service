package services

import (
	"net/http"

	"github.com/nikitaSstepanov/p2p-streaming-service/backend/internal/pkg/services/services"
	"github.com/nikitaSstepanov/p2p-streaming-service/backend/internal/pkg/storage"
	"github.com/nikitaSstepanov/p2p-streaming-service/backend/internal/pkg/state"
	"github.com/redis/go-redis/v9"
	"github.com/gin-gonic/gin"
)

type Services struct {
	Movies    *services.Movies
	Account   *services.Account
	Comments  *services.Comments
	Playlists *services.Playlists
	Admin     *services.Admin
	Auth      *services.Auth
	State     *state.State
}

func New(storage *storage.Storage, state *state.State, redis *redis.Client) *Services {
	auth := services.NewAuth()
	
	return &Services {
		Movies:    services.NewMovies(storage, state, redis),
		Account:   services.NewAccount(storage, redis, auth),
		Comments:  services.NewComments(storage),
		Playlists: services.NewPlaylists(storage, redis, auth),
		Admin:     services.NewAdmin(storage, redis, auth),
		Auth:      auth,
	}
}

func (s *Services) PingFunc(ctx *gin.Context) {
	ctx.JSON(http.StatusOK, "pong")
}