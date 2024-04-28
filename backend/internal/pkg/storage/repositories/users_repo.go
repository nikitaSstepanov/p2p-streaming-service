package repositories

import (
	"context"
	"fmt"

	"github.com/nikitaSstepanov/p2p-streaming-service/backend/internal/pkg/storage/entities"
	"github.com/nikitaSstepanov/p2p-streaming-service/backend/pkg/postgresql"
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

	row.Scan(&user.Id, &user.Username, &user.Password, &user.Role)

	return &user
}

func (u *Users) GetUsersByRole(ctx context.Context, role string) *[]entities.User {
	var users []entities.User

	var user entities.User

	query := fmt.Sprintf("SELECT * FROM %s WHERE role = '%s';", usersTable, role)

	rows, _ := u.db.Query(ctx, query)

	for rows.Next() {
		rows.Scan(&user.Id, &user.Username, &user.Password, &user.Role)

		users = append(users, user)
	}

	return &users
}

func (u *Users) Create(ctx context.Context, user *entities.User) {
	query := fmt.Sprintf("INSERT INTO %s (username, password, role) VALUES ('%s', '%s', '%s') ON CONFLICT DO NOTHING;", usersTable, user.Username, user.Password, user.Role)

	u.db.QueryRow(ctx, query)
}

func (u *Users) Update(ctx context.Context, user *entities.User) {
	query := fmt.Sprintf("UPDATE %s SET	role = '%s' WHERE id = %d;", usersTable, user.Role, user.Id)

	u.db.QueryRow(ctx, query)
}
