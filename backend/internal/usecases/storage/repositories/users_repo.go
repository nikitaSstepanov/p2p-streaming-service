package repositories

import (
	"context"
	"fmt"

	"github.com/nikitaSstepanov/p2p-streaming-service/backend/internal/usecases/storage/entities"
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

func (u *Users) GetUser(ctx context.Context, username string) (entities.User, error) {
	var user entities.User

	query := fmt.Sprintf("SELECT * FROM %s WHERE username = '%s'", usersTable, username)

	row := u.db.QueryRow(ctx, query)

	err := row.Scan(&user.Id, &user.Username, &user.Password, &user.Role)

	if err != nil {
		return entities.User{}, err
	}

	return user, nil
}

func (u *Users) GetUsersByRole(ctx context.Context, role string) (*[]entities.User, error) {
	var users []entities.User

	var user entities.User

	query := fmt.Sprintf("SELECT * FROM %s WHERE role = '%s';", usersTable, role)

	rows, err := u.db.Query(ctx, query)

	if err != nil {
		return nil, err
	}

	for rows.Next() {
		err = rows.Scan(&user.Id, &user.Username, &user.Password, &user.Role)

		if err != nil {
			return nil, err
		}

		users = append(users, user)
	}

	return &users, nil
}

func (u *Users) Create(ctx context.Context, user *entities.User) error {
	query := fmt.Sprintf("INSERT INTO %s (username, password, role) VALUES ($1, $2, $3) ON CONFLICT DO NOTHING;", usersTable)

	tx, err := u.db.Begin(ctx)

	if err != nil {
		return err
	}

	_, err = tx.Exec(ctx, query, user.Username, user.Password, user.Role)

	if err != nil {
		return err
	}

	err = tx.Commit(ctx)

	if err != nil {
		return err
	}

	return nil
}

func (u *Users) Update(ctx context.Context, user *entities.User) error {
	query := fmt.Sprintf("UPDATE %s SET	role = $1 WHERE id = $2;", usersTable)

	tx, err := u.db.Begin(ctx)

	if err != nil {
		return err
	}

	_, err = tx.Exec(ctx, query, user.Role, user.Id)

	if err != nil {
		return err
	}

	err = tx.Commit(ctx)

	if err != nil {
		return err
	}

	return nil
}
