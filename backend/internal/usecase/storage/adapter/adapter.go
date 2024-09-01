package adapter

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
	adaptersTable = "adapters"
	redisExpires  = 3 * time.Hour
)

var (
	internalErr = e.New("Something going wrong...", e.Internal)
	notFoundErr = e.New("This adapter wasn`t found", e.NotFound)
)

type Adapter struct {
	postgres postgresql.Client
	redis    *goredis.Client
}

func New(pgClient postgresql.Client, redisClient *goredis.Client) *Adapter {
	return &Adapter{
		postgres: pgClient,
		redis:    redisClient,
	}
}

func (a *Adapter) GetAdapter(ctx context.Context, movieId uint64, version int) (*entity.Adapter, *e.Error) {
	var adapter entity.Adapter

	err := a.redis.Get(ctx, getRedisKey(movieId, version)).Scan(&adapter)
	if err != nil && err != goredis.Nil {
		return nil, internalErr
	}

	if adapter.Id != 0 {
		return &adapter, nil
	}

	query := fmt.Sprintf("SELECT * FROM %s WHERE movieId = %d AND version = %d;", adaptersTable, movieId, version)

	row := a.postgres.QueryRow(ctx, query)

	err = row.Scan(&adapter.Id, &adapter.MovieId, &adapter.Version, &adapter.Length, &adapter.PieceLength)
	
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, notFoundErr
		} else {
			return nil, internalErr
		}
	}

	err = a.redis.Set(ctx, getRedisKey(movieId, version), &adapter, redisExpires).Err()
	if err != nil {
		return nil, internalErr
	}

	return &adapter, nil
}

func (a *Adapter) CreateAdapter(ctx context.Context, adapter *entity.Adapter) *e.Error {
	query := fmt.Sprintf("INSERT INTO %s (movieId, version, length, pieceLength) VALUES ($1, $2, $3, $4) ON CONFLICT DO NOTHING;", adaptersTable)

	tx, err := a.postgres.Begin(ctx)
	if err != nil {
		return internalErr
	}
	defer tx.Rollback(ctx)

	_, err = tx.Exec(ctx, query, adapter.MovieId, adapter.Version, adapter.Length, adapter.PieceLength)
	if err != nil {
		return internalErr
	}

	if err = tx.Commit(ctx); err != nil {
		return internalErr
	}

	return nil
}

func getRedisKey(movieId uint64, fileId int) string {
	return fmt.Sprintf("adapters:%d:%d", movieId, fileId)
}