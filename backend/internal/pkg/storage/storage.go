package storage

import (
	"github.com/nikitaSstepanov/p2p-streaming-service/backend/internal/pkg/storage/repositories"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Storage struct {
	Movies *repositories.Movies
	Users *repositories.Users
}

func New(db *pgxpool.Pool) *Storage {
	return &Storage{
		Movies: repositories.NewMovies(db),
		Users: repositories.NewUsers(db),
	}
}