package token

import (
	"context"
	"fmt"

	"github.com/nikitaSstepanov/p2p-streaming-service/backend/internal/entity"
	"github.com/nikitaSstepanov/p2p-streaming-service/backend/pkg/client/postgresql"
	e "github.com/nikitaSstepanov/p2p-streaming-service/backend/pkg/errors"
)

const (
	tokensTable = "tokens"
)

var (
	internalErr = e.New("Something going wrong...", e.Internal)
)

type Token struct {
	postgres postgresql.Client
}

func New(postgres postgresql.Client) *Token {
	return &Token{
		postgres,
	}
}

func (t *Token) GetUserByRefresh(ctx context.Context, refresh string) (*entity.User, *e.Error) {
	var user entity.User
	
	query := fmt.Sprintf("SELECT * FROM %s WHERE token = '%s'", tokensTable, refresh)

	row := t.postgres.QueryRow(ctx, query)

	if err := user.Scan(row); err != nil {
		return nil, internalErr
	}

	return &user, nil
}

func (t *Token) SetRefresh(ctx context.Context, userId uint64, token string) *e.Error {
	query := fmt.Sprintf("INSERT INTO %s (userId, token) VALUES ($1, $2);", tokensTable)

	tx, err := t.postgres.Begin(ctx)
	if err != nil {
		return internalErr
	}
	defer tx.Rollback(ctx)

	_, err = tx.Exec(ctx, query, userId, token)
	if err != nil {
		return internalErr
	}
	
	if err := tx.Commit(ctx); err != nil {
		return internalErr
	}

	return nil
}

func (t *Token) DeleteRefresh(ctx context.Context, userId uint64) *e.Error {
	query := fmt.Sprintf("DELETE FROM %s WHERE userId = %d;", tokensTable, userId)

	tx, err := t.postgres.Begin(ctx)
	if err != nil {
		return internalErr
	}
	defer tx.Rollback(ctx)

	if _, err = tx.Exec(ctx, query); err != nil {
		return internalErr
	}

	if err := tx.Commit(ctx); err != nil {
		return internalErr
	}

	return nil
}