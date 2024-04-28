package services

import (
	"os"

	"github.com/nikitaSstepanov/p2p-streaming-service/backend/internal/pkg/services/services"
	"github.com/nikitaSstepanov/p2p-streaming-service/backend/internal/pkg/storage"
	"github.com/nikitaSstepanov/p2p-streaming-service/backend/internal/pkg/state"
	"github.com/redis/go-redis/v9"
)

type Services struct {
	Movies   *services.Movies
	Account  *services.Account
	Admin    *services.Admin
	Auth     *services.Auth
	State    *state.State
}

func New(storage *storage.Storage, state *state.State, redis *redis.Client) *Services {
	auth := services.NewAuth()

	initFilesDirectory()
	
	return &Services {
		Movies:  services.NewMovies(storage, state, redis),
		Account: services.NewAccount(storage, auth),
		Admin:   services.NewAdmin(storage, auth),
		Auth:    auth,
	}
}

func initFilesDirectory() {
	err := os.MkdirAll("files", 0777)
	if err != nil {
		panic(err)
	}
}