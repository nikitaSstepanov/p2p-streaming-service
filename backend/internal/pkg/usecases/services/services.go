package services

import (
	"github.com/nikitaSstepanov/p2p-streaming-service/backend/internal/pkg/usecases/services/services"
	"github.com/nikitaSstepanov/p2p-streaming-service/backend/internal/pkg/usecases/storage"
	"github.com/nikitaSstepanov/p2p-streaming-service/backend/internal/pkg/usecases/state"
	"github.com/redis/go-redis/v9"
)

type Services struct {
	Movies    *services.Movies
	Account   *services.Account
	Comments  *services.Comments
	Playlists *services.Playlists
	Admin     *services.Admin
}

func New(storage *storage.Storage, state *state.State, redis *redis.Client) *Services {
	auth := services.NewAuth()
	
	return &Services {
		Movies:    services.NewMovies(storage, state, redis),
		Account:   services.NewAccount(storage, redis, auth),
		Comments:  services.NewComments(storage, redis, auth),
		Playlists: services.NewPlaylists(storage, redis, auth),
		Admin:     services.NewAdmin(storage, redis, auth),
	}
}
