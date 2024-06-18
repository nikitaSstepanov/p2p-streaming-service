package repositories

import (
	"context"
	"fmt"

	"github.com/nikitaSstepanov/p2p-streaming-service/backend/internal/usecases/storage/entities"
	"github.com/nikitaSstepanov/p2p-streaming-service/backend/pkg/postgresql"
)

const (
	adaptersTable = "adapters"
)

type Adapters struct {
	db postgresql.Client
}

func NewAdapters(db postgresql.Client) *Adapters {
	return &Adapters{
		db: db,
	}
}

func (a *Adapters) GetAdapter(ctx context.Context, movieId uint64, version uint64) (entities.Adapter, error) {
	var adapter entities.Adapter

	query := fmt.Sprintf("SELECT * FROM %s WHERE movieId = %d AND version = %d ;", adaptersTable, movieId, version)

	row := a.db.QueryRow(ctx, query)

	err := row.Scan(&adapter.Id, &adapter.MovieId, &adapter.Version, &adapter.Length, &adapter.PieceLength)

	if err != nil {
		return entities.Adapter{}, err
	}

	return adapter, nil
}

func (a *Adapters) CreateAdapter(ctx context.Context, adapter *entities.Adapter) error {
	query := fmt.Sprintf("INSERT INTO %s (movieId, version, length, pieceLength) VALUES ($1, $2, $3, $4) ON CONFLICT DO NOTHING;", adaptersTable)

	tx, err := a.db.Begin(ctx)

	if err != nil {
		return err
	}

	_, err = tx.Exec(ctx, query, adapter.MovieId, adapter.Version, adapter.Length, adapter.PieceLength)

	if err != nil {
		return err
	}

	err = tx.Commit(ctx)

	if err != nil {
		return err
	}

	return nil
}