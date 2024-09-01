package playlist

import (
	"context"
	"time"
	"fmt"

	"github.com/nikitaSstepanov/p2p-streaming-service/backend/pkg/client/postgresql"
	"github.com/nikitaSstepanov/p2p-streaming-service/backend/internal/entity"
	e "github.com/nikitaSstepanov/p2p-streaming-service/backend/pkg/errors"
	goredis "github.com/redis/go-redis/v9"
	"github.com/jackc/pgx/v5"
)

const (
	playlistsTable       = "playlists"
	playlistsMoviesTable = "movies_playlists"
	redisExpires         = 3 * time.Hour
)

var (
	internalErr = e.New("Something going wrong...", e.Internal)
	notFoundErr = e.New("This playlist wasn`t found", e.NotFound)
)

type Playlist struct {
	postgres postgresql.Client
	redis    *goredis.Client
}

func New(pgClient postgresql.Client, redisClient *goredis.Client) *Playlist {
	return &Playlist{
		postgres: pgClient,
		redis:    redisClient,
	}
}

func (p *Playlist) GetPlaylists(ctx context.Context, userId uint64, limit int, offset int) ([]*entity.Playlist, *e.Error) {
	var playlists []*entity.Playlist

	var playlist entity.Playlist
	
	query := fmt.Sprintf(
		"SELECT * FROM %s WHERE userId = %d LIMIT %d OFFSET %d;", 
		playlistsTable, userId, limit, offset,
	)

	rows, err := p.postgres.Query(ctx, query)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, notFoundErr
		} else {
			return nil, internalErr
		}
	}

	for rows.Next() {
		if err := playlist.Scan(rows); err != nil {
			return nil, internalErr
		}

		playlists = append(playlists, &playlist)
	}

	return playlists, nil
}

func (p *Playlist) GetPlaylist(ctx context.Context, id uint64) (*entity.Playlist, *e.Error) {
	var playlist entity.Playlist

	err := p.redis.Get(ctx, getRedisKey(id)).Scan(&playlist)
	if err != nil && err != goredis.Nil {
		return nil, internalErr
	}

	if playlist.Id != 0 {
		return &playlist, nil
	}

	query := fmt.Sprintf("SELECT * FROM %s WHERE id = %d;", playlistsTable, id)

	row := p.postgres.QueryRow(ctx, query)
	
	if err := playlist.Scan(row); err != nil {
		if err == pgx.ErrNoRows {
			return nil, notFoundErr
		} else {
			return nil, internalErr
		}
	}

	var movies []uint64

	var movie entity.PlaylistMovies

	query = fmt.Sprintf("SELECT * FROM %s WHERE playlistId = %d;", playlistsMoviesTable, id)

	rows, err := p.postgres.Query(ctx, query)
	
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, notFoundErr
		} else {
			return nil, internalErr
		}
	}

	for rows.Next() {
		if err := movie.Scan(rows); err != nil {
			return nil, internalErr
		}

		movies = append(movies, movie.MovieId)
	}

	playlist.MoviesIds = movies

	err = p.redis.Set(ctx, getRedisKey(id), &playlist, redisExpires).Err()
	if err != nil {
		return nil, internalErr
	}

	return &playlist, nil
}

func (p *Playlist) CreatePlaylist(ctx context.Context, playlist *entity.Playlist) *e.Error {
	query := fmt.Sprintf("INSERT INTO %s (userId, title) VALUES ($1, $2) RETURNING id;", playlistsTable)

	tx, err := p.postgres.Begin(ctx)
	if err != nil {
		return internalErr
	}
	defer tx.Rollback(ctx)
	
	row := tx.QueryRow(ctx, query, playlist.UserId, playlist.Title)

	if err = row.Scan(&playlist.Id); err != nil {
		return internalErr
	}

	if err := tx.Commit(ctx); err != nil {
		return internalErr
	}

	err = p.redis.Set(ctx, getRedisKey(playlist.Id), &playlist, redisExpires).Err()
	if err != nil {
		return internalErr
	}

	return nil
}

func (p *Playlist) UpdatePlaylist(ctx context.Context, playlist *entity.Playlist) *e.Error {
	query := fmt.Sprintf("UPDATE %s SET title = $1 WHERE id = $2;", playlistsTable)

	tx, err := p.postgres.Begin(ctx)
	if err != nil {
		return internalErr
	}
	defer tx.Rollback(ctx)

	_, err = tx.Exec(ctx, query, playlist.Title, playlist.Id)
	if err != nil {
		return internalErr
	}

	if err = tx.Commit(ctx); err != nil {
		return internalErr
	}

	err = p.redis.Del(ctx, getRedisKey(playlist.Id)).Err()
	if err != nil {
		return internalErr
	}

	err = p.redis.Set(ctx, getRedisKey(playlist.Id), &playlist, redisExpires).Err()
	if err != nil {
		return internalErr
	}

	return nil
}

func (p *Playlist) AddMovie(ctx context.Context, playlistId uint64, movieId uint64) *e.Error {
	query := fmt.Sprintf("SELECT * FROM %s WHERE playlistId = %d AND movieId = %d;", playlistsMoviesTable, playlistId, movieId)

	tx, err := p.postgres.Begin(ctx)
	if err != nil {
		return internalErr
	}
	defer tx.Rollback(ctx)
	
	row := tx.QueryRow(ctx, query)

	var movie entity.PlaylistMovies

	err = movie.Scan(row)
	if err != nil && err != pgx.ErrNoRows {
		return internalErr
	}

	if movie.MovieId != 0 {
		return nil
	}

	query = fmt.Sprintf("INSERT INTO %s (playlistId, movieId) VALUES ($1, $2) ON CONFLICT DO NOTHING;", playlistsMoviesTable)

	_, err = tx.Exec(ctx, query, playlistId, movieId)
	if err != nil {
		return internalErr
	}

	if err = tx.Commit(ctx); err != nil {
		return internalErr
	}

	err = p.redis.Del(ctx, getRedisKey(playlistId)).Err()
	if err != nil {
		return internalErr
	}

	return nil
}

func (p *Playlist) RemoveMovie(ctx context.Context, playlistId uint64, movieId uint64) *e.Error {
	query := fmt.Sprintf("DELETE FROM %s WHERE playlistId = $1 AND movieId = $2;", playlistsMoviesTable)

	tx, err := p.postgres.Begin(ctx)
	if err != nil {
		return internalErr
	}
	defer tx.Rollback(ctx)

	_, err = tx.Exec(ctx, query, playlistId, movieId)
	if err != nil {
		return internalErr
	}

	if err = tx.Commit(ctx); err != nil {
		return internalErr
	}

	err = p.redis.Del(ctx, getRedisKey(playlistId)).Err()
	if err != nil {
		return internalErr
	}

	return nil
}

func (p *Playlist) DeletePlaylist(ctx context.Context, id uint64) *e.Error {
	query := fmt.Sprintf("DELETE FROM %s WHERE id = $1;", playlistsTable)

	tx, err := p.postgres.Begin(ctx)
	if err != nil {
		return internalErr
	}
	defer tx.Rollback(ctx)

	_, err = tx.Exec(ctx, query, id)
	if err != nil {
		return internalErr
	}

	if err = tx.Commit(ctx); err != nil {
		return internalErr
	}

	err = p.redis.Del(ctx, getRedisKey(id)).Err()
	if err != nil {
		return internalErr
	}

	return nil
}

func getRedisKey(id uint64) string {
	return fmt.Sprintf("playlists:%d", id)
}