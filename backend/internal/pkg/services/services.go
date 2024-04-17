package services

import (
	"github.com/nikitaSstepanov/p2p-streaming-service/backend/internal/pkg/state"
	"github.com/nikitaSstepanov/p2p-streaming-service/backend/internal/pkg/storage"
)

type Services struct {
	Movies *Movies
	Account *Account
	State *state.State
}

func New(storage *storage.Storage, state *state.State) *Services {
	return &Services {
		Movies: NewMovies(storage, state),
		Account: NewAccount(storage),
	}
}
