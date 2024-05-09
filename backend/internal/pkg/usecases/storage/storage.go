package storage

import (
	"github.com/nikitaSstepanov/p2p-streaming-service/backend/internal/pkg/usecases/storage/repositories"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Storage struct {
	Movies    *repositories.Movies
	Users     *repositories.Users
	Playlists *repositories.Playlists
	Comments  *repositories.Comments
	Adapters  *repositories.Adapters
}

func New(db *pgxpool.Pool) *Storage {
	return &Storage{
		Movies:    repositories.NewMovies(db),
		Users:     repositories.NewUsers(db),
		Playlists: repositories.NewPlaylists(db),
		Comments:  repositories.NewComments(db),
		Adapters:  repositories.NewAdapters(db),
	}
}