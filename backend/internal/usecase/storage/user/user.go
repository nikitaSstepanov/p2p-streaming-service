package user

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
	redisExpires = 3 * time.Hour
	usersTable   = "users"
)

var (
	internalErr = e.New("Something going wrong...", e.Internal)
	conflictErr = e.New("User with this email already exist", e.Conflict)
	notFoundErr = e.New("This user wasn`t found", e.NotFound)
)

type User struct {
	postgres postgresql.Client
	redis    *goredis.Client
}

func New(pgClient postgresql.Client, redisClient *goredis.Client) *User {
	return &User{
		postgres: pgClient,
		redis:    redisClient,
	}
}

func (u *User) GetUser(ctx context.Context, id uint64) (*entity.User, *e.Error) {
	var user entity.User

	err := u.redis.Get(ctx, getRedisKey(id)).Scan(&user)
	if err != nil && err != goredis.Nil {
		return nil, internalErr
	}

	if user.Id != 0 {
		return &user, nil
	}

	query := fmt.Sprintf("SELECT * FROM %s WHERE id = %d", usersTable, id)

	tx, err := u.postgres.Begin(ctx)
	if err != nil {
		return nil, internalErr
	}
	defer tx.Rollback(ctx)

	row := tx.QueryRow(ctx, query)
	
	if err = user.Scan(row); err != nil {
		if err == pgx.ErrNoRows {
			return nil, notFoundErr
		} else {
			return nil, internalErr
		}
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, internalErr
	}

	err = u.redis.Set(ctx, getRedisKey(id), &user, redisExpires).Err()
	if err != nil {
		return nil, internalErr
	}

	return &user, nil
}

func (u *User) GetUserByUsername(ctx context.Context, username string) (*entity.User, *e.Error) {
	var user entity.User

	query := fmt.Sprintf("SELECT * FROM %s WHERE username = '%s';", usersTable, username)

	tx, err := u.postgres.Begin(ctx)
	if err != nil {
		return nil, internalErr
	}
	defer tx.Rollback(ctx)

	row := tx.QueryRow(ctx, query)

	if err = user.Scan(row); err != nil {
		if err == pgx.ErrNoRows {
			return nil, notFoundErr
		} else {
			return nil, internalErr
		}
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, internalErr
	}

	return &user, nil
}

func (u *User) GetUsersByRole(ctx context.Context, role string) ([]*entity.User, *e.Error) {
	var users []*entity.User

	var user entity.User

	query := fmt.Sprintf("SELECT * FROM %s WHERE role = '%s';", usersTable, role)

	tx, err := u.postgres.Begin(ctx)
	if err != nil {
		return nil, internalErr
	}
	defer tx.Rollback(ctx)

	rows, err := tx.Query(ctx, query)

	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, notFoundErr
		} else {
			return nil, internalErr
		}
	}

	for rows.Next() {
		if err = user.Scan(rows); err != nil {
			return nil, internalErr
		}

		users = append(users, &user)
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, internalErr
	}

	return users, nil
}

func (u *User) Create(ctx context.Context, user *entity.User) *e.Error {
	candidate, checkUsernameError := u.GetUserByUsername(ctx, user.Username)
	if checkUsernameError != nil {
		return checkUsernameError
	}

	if candidate.Id != 0 {
		return conflictErr
	}
	
	query := fmt.Sprintf(
		"INSERT INTO %s (username, password, role) VALUES ('%s', '%s', '%s') RETURNING id;", 
		usersTable, user.Username, user.Password, user.Password,
	)

	tx, err := u.postgres.Begin(ctx)
	if err != nil {
		return internalErr
	}
	defer tx.Rollback(ctx)

	row := tx.QueryRow(ctx, query)

	err = row.Scan(&user.Id)
	if err != nil {
		return internalErr
	}

	if err := tx.Commit(ctx); err != nil {
		return internalErr
	}

	err = u.redis.Set(ctx, getRedisKey(user.Id), user, redisExpires).Err()
	if err != nil {
		return internalErr
	}

	return nil
}

func (u *User) Update(ctx context.Context, user *entity.User) *e.Error {
	_, checkUsernameError := u.GetUserByUsername(ctx, user.Username)
	if (checkUsernameError != nil) {
		return checkUsernameError
	}

	query := fmt.Sprintf("UPDATE %s SET	role = $1 WHERE id = $2;", usersTable)

	tx, err := u.postgres.Begin(ctx)
	if err != nil {
		return internalErr
	}
	defer tx.Rollback(ctx)

	_, err = tx.Exec(ctx, query, user.Role, user.Id)
	if err != nil {
		return internalErr
	}

	if err := tx.Commit(ctx); err != nil {
		return internalErr
	}

	return nil
}

func getRedisKey(id uint64) string {
	return fmt.Sprintf("users:%d", id)
}