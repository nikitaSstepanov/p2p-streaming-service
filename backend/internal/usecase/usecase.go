package usecase

import (
	"github.com/nikitaSstepanov/p2p-streaming-service/backend/internal/usecase/pkg/account"
	"github.com/nikitaSstepanov/p2p-streaming-service/backend/internal/usecase/pkg/admin"
	"github.com/nikitaSstepanov/p2p-streaming-service/backend/internal/usecase/pkg/auth"
	"github.com/nikitaSstepanov/p2p-streaming-service/backend/internal/usecase/pkg/comment"
	"github.com/nikitaSstepanov/p2p-streaming-service/backend/internal/usecase/pkg/movie"
	"github.com/nikitaSstepanov/p2p-streaming-service/backend/internal/usecase/pkg/playlist"
	"github.com/nikitaSstepanov/p2p-streaming-service/backend/internal/usecase/state"
	"github.com/nikitaSstepanov/p2p-streaming-service/backend/internal/usecase/storage"
)

type UseCase struct {
	Movies *movie.Movie
	Accounts *account.Account
	Admin    *admin.Admin
	Auth     *auth.Auth
	Comment  *comment.Comment
	Playlist *playlist.Playlist
}

func New(store *storage.Storage, state *state.State, jwt *auth.JwtUseCase) *UseCase {
	return &UseCase{
		Movies:   movie.New(store.Movies, store.Adapters, state),
		Accounts: account.New(store.Users, jwt),
		Admin:    admin.New(store.Users, store.Movies),
		Auth:     auth.New(jwt, store.Users, store.Tokens),
		Comment:  comment.New(store.Comments, store.Movies),
		Playlist: playlist.New(store.Playlists, store.Movies),
	}
}