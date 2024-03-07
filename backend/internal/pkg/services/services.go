package services

import "github.com/nikitaSstepanov/p2p-streaming-service/backend/internal/pkg/storage"

type Services struct {
	Movies *Movies
	Account *Account
}

func New(storage *storage.Storage) *Services {

	return &Services {
		Movies: NewMovies(),
		Account: NewAccount(),
	}

}
