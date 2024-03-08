package repositories

import (
	"context"
	"fmt"

	"github.com/nikitaSstepanov/p2p-streaming-service/backend/internal/pkg/storage/entities"
	"github.com/nikitaSstepanov/p2p-streaming-service/backend/libs/postgresql"
)

const (
	usersTable = "users"
)

type Users struct {
	db postgresql.Client
}

func NewUsers(db postgresql.Client) *Users {
	return &Users{
		db: db,
	}
}

func (u *Users) GetUser(ctx context.Context, username string) *entities.User {
	var user entities.User

	query := fmt.Sprintf("SELECT * FROM %s WHERE username = '%s'", usersTable, username)

	row := u.db.QueryRow(ctx, query)

	row.Scan(&user.Id, &user.Username, &user.Password)

	return &user
}

func (u *Users) Create(ctx context.Context, user *entities.User) {
	query := fmt.Sprintf("INSERT INTO %s (username, password) VALUES ('%s', '%s') ON CONFLICT DO NOTHING;", usersTable, user.Username, user.Password)

	u.db.QueryRow(ctx, query)
}
