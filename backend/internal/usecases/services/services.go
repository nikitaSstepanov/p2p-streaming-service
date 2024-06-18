package services

import (
	"github.com/nikitaSstepanov/p2p-streaming-service/backend/internal/usecases/services/services"
	"github.com/nikitaSstepanov/p2p-streaming-service/backend/internal/usecases/storage"
	"github.com/nikitaSstepanov/p2p-streaming-service/backend/internal/usecases/state"
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
	account := services.NewAccount(storage, redis, auth)
	
	return &Services {
		Movies:    services.NewMovies(storage, state, redis),
		Account:   account,
		Comments:  services.NewComments(storage, redis, account),
		Playlists: services.NewPlaylists(storage, redis, account),
		Admin:     services.NewAdmin(storage, redis, account, auth),
	}
}
