package storage

import (
	"github.com/nikitaSstepanov/p2p-streaming-service/backend/internal/usecase/storage/adapter"
	"github.com/nikitaSstepanov/p2p-streaming-service/backend/internal/usecase/storage/comment"
	"github.com/nikitaSstepanov/p2p-streaming-service/backend/internal/usecase/storage/movie"
	"github.com/nikitaSstepanov/p2p-streaming-service/backend/internal/usecase/storage/playlist"
	"github.com/nikitaSstepanov/p2p-streaming-service/backend/internal/usecase/storage/token"
	"github.com/nikitaSstepanov/p2p-streaming-service/backend/internal/usecase/storage/user"
	"github.com/nikitaSstepanov/p2p-streaming-service/backend/pkg/client/postgresql"
	"github.com/redis/go-redis/v9"
)

type Storage struct {
	Movies    *movie.Movie
	Users     *user.User
	Comments  *comment.Comment
	Playlists *playlist.Playlist
	Tokens    *token.Token
	Adapters  *adapter.Adapter
}

func New(postgres postgresql.Client, redis *redis.Client) *Storage {
	return &Storage{
		Movies:    movie.New(postgres, redis),
		Users:     user.New(postgres, redis),
		Comments:  comment.New(postgres, redis),
		Playlists: playlist.New(postgres, redis),
		Tokens:    token.New(postgres),
		Adapters:  adapter.New(postgres, redis),		
	}
}