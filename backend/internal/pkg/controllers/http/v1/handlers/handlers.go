package handlers

import (
	"github.com/nikitaSstepanov/p2p-streaming-service/backend/internal/pkg/controllers/http/v1/handlers/handlers"
	"github.com/nikitaSstepanov/p2p-streaming-service/backend/internal/pkg/usecases/services"
)

type Handlers struct {
	Movies    *handlers.Movies
	Account   *handlers.Account
	Playlists *handlers.Playlists
	Comments  *handlers.Comments
	Admin     *handlers.Admin
	Default   *handlers.Default
}

func New(services *services.Services) *Handlers {
	return &Handlers{
		Movies:    handlers.NewMovies(services),
		Account:   handlers.NewAccount(services),
		Playlists: handlers.NewPlaylists(services),
		Comments:  handlers.NewComments(services),
		Admin:     handlers.NewAdmin(services),
		Default:   handlers.NewDefault(),
	}
}