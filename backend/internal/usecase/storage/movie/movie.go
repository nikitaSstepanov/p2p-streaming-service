package movie

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
	moviesTable  = "movies"
	redisExpires = 3 * time.Hour
)

var (
	internalErr = e.New("Something going wrong...", e.Internal)
	notFoundErr = e.New("This movie wasn`t found", e.NotFound)
)

type Movie struct {
	postgres postgresql.Client
	redis    *goredis.Client
}

func New(pgClient postgresql.Client, redisClient *goredis.Client) *Movie {
	return &Movie{
		postgres: pgClient,
		redis:    redisClient,
	}
}

func (m *Movie) GetAllMovies(ctx context.Context, limit int, offset int) ([]*entity.Movie, *e.Error) {
	query := fmt.Sprintf("SELECT * FROM %s LIMIT %d OFFSET %d;", moviesTable, limit, offset)

	rows, err := m.postgres.Query(ctx, query)
	
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, notFoundErr
		} else {
			return nil, internalErr
		}
	}

	var movies []*entity.Movie

	var movie entity.Movie

	for rows.Next() {
		err = movie.Scan(rows)
		if err != nil {
			return nil, internalErr
		}

		movies = append(movies, &movie)
	}

	return movies, nil
}

func (m *Movie) GetMovieById(ctx context.Context, id uint64) (*entity.Movie, *e.Error) {
	var movie entity.Movie

	err := m.redis.Get(ctx, getRedisKey(id)).Scan(&movie)
	if err != nil && err != goredis.Nil {
		return nil, internalErr
	}

	if movie.Id != 0 {
		return &movie, nil
	}
	
	query := fmt.Sprintf("SELECT * FROM %s WHERE id = %d;", moviesTable, id)

	row := m.postgres.QueryRow(ctx, query)

	err = movie.Scan(row)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, notFoundErr
		} else {
			return nil, internalErr
		}
	}

	err = m.redis.Set(ctx, getRedisKey(id), &movie, redisExpires).Err()
	if err != nil {
		return nil, internalErr
	}

	return &movie, nil
}

func (m *Movie) CreateMovie(ctx context.Context, movie *entity.Movie) *e.Error {
	query := fmt.Sprintf("INSERT INTO %s (name, paths, fileVersion) VALUES ($1, $2, $3);", moviesTable)

	tx, err := m.postgres.Begin(ctx)
	if err != nil {
		return internalErr
	}
	defer tx.Rollback(ctx)

	_, err = tx.Exec(ctx, query,  movie.Name, movie.Paths, movie.FileVersion)
	if err != nil {
		return internalErr
	}

	if err = tx.Commit(ctx); err != nil {
		return internalErr
	}

	err = m.redis.Set(ctx, getRedisKey(movie.Id), &movie, redisExpires).Err()
	if err != nil {
		return internalErr
	}

	return nil
}

func (m *Movie) UpdateMovie(ctx context.Context, movie *entity.Movie) *e.Error {
	query := fmt.Sprintf("UPDATE %s SET fileVersion = $1, paths = $2, name = $3 WHERE id = $4;", moviesTable)

	tx, err := m.postgres.Begin(ctx)
	if err != nil {
		return internalErr
	}
	defer tx.Rollback(ctx)

	_, err = tx.Exec(ctx, query, movie.FileVersion, movie.Paths, movie.Name, movie.Id)
	if err != nil {
		return internalErr
	}

	if err = tx.Commit(ctx); err != nil {
		return internalErr
	}

	err = m.redis.Del(ctx, getRedisKey(movie.Id)).Err()
	if err != nil {
		return internalErr
	}

	err = m.redis.Set(ctx, getRedisKey(movie.Id), &movie, redisExpires).Err()
	if err != nil {
		return internalErr
	}

	return nil
}

func getRedisKey(id uint64) string {
	return fmt.Sprintf("movies:%d", id)
}