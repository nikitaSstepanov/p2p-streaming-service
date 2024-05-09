package repositories

import (
	"context"
	"fmt"

	"github.com/nikitaSstepanov/p2p-streaming-service/backend/internal/pkg/usecases/storage/entities"
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

func (a *Adapters) GetAdapter(ctx context.Context, movieId uint64, version uint64) *entities.Adapter {
	var adapter entities.Adapter

	query := fmt.Sprintf("SELECT * FROM %s WHERE movieId = %d AND version = %d ;", adaptersTable, movieId, version)

	row := a.db.QueryRow(ctx, query)

	row.Scan(&adapter.Id, &adapter.MovieId, &adapter.Version, &adapter.Length, &adapter.PieceLength)

	return &adapter
}

func (a *Adapters) CreateAdapter(ctx context.Context, adapter *entities.Adapter) {
	query := fmt.Sprintf("INSERT INTO %s (movieId, version, length, pieceLength) VALUES (%d, %d, %d, %d) ON CONFLICT DO NOTHING;", adaptersTable, adapter.MovieId, adapter.Version, adapter.Length, adapter.PieceLength)

	a.db.QueryRow(ctx, query)
}