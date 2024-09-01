package comment

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
	commentsTable = "comments"
	redisExpires  = 3 * time.Hour
)

var (
	internalErr = e.New("Something going wrong...", e.Internal)
	notFoundErr = e.New("This comment or comments wasn`t found", e.NotFound)
)

type Comment struct {
	postgres postgresql.Client
	redis    *goredis.Client
}

func New(pgClient postgresql.Client, redisClient *goredis.Client) *Comment {
	return &Comment{
		postgres: pgClient,
		redis:    redisClient,
	}
}

func (c *Comment) GetComments(ctx context.Context, movieId uint64, limit int, offset int) ([]*entity.Comment, *e.Error) {
	var comments []*entity.Comment

	var comment entity.Comment

	query := fmt.Sprintf("SELECT * FROM %s WHERE movieId = %d LIMIT %d OFFSET %d;", commentsTable, movieId, limit, offset)

	rows, err := c.postgres.Query(ctx, query)
	
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, notFoundErr
		} else {
			return nil, internalErr
		}
	}

	for rows.Next() {
		err = rows.Scan(&comment.Id, &comment.MovieId, &comment.UserId, &comment.Text)
		if err != nil {
			return nil, internalErr
		}

		comments = append(comments, &comment)
	}

	return comments, nil
}

func (c *Comment) GetCommentById(ctx context.Context, id uint64) (*entity.Comment, *e.Error) {
	var comment entity.Comment

	err := c.redis.Get(ctx, getRedisKey(id)).Scan(&comment)
	if err != nil && err != goredis.Nil {
		return nil, internalErr
	}

	if comment.Id != 0 {
		return &comment, nil
	}

	query := fmt.Sprintf("SELECT * FROM %s WHERE id = %d;", commentsTable, id)

	row := c.postgres.QueryRow(ctx, query)

	err = row.Scan(&comment.Id, &comment.MovieId, &comment.UserId, &comment.Text)

	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, notFoundErr
		} else {
			return nil, internalErr
		}
	}

	return &comment, nil
}

func (c *Comment) CreateComment(ctx context.Context, comment *entity.Comment) *e.Error {
	query := fmt.Sprintf("INSERT INTO %s (movieId, userId, text) VALUES ($1, $2, $3);", commentsTable)

	tx, err := c.postgres.Begin(ctx)
	if err != nil {
		return internalErr
	}
	defer tx.Rollback(ctx)

	_, err = tx.Exec(ctx, query, comment.MovieId, comment.UserId, comment.Text)
	if err != nil {
		return internalErr
	}

	if err = tx.Commit(ctx); err != nil {
		return internalErr
	}

	return nil
}

func (c *Comment) UpdateComment(ctx context.Context, comment *entity.Comment) *e.Error {
	query := fmt.Sprintf("UPDATE %s SET text = $1 WHERE id = $2;", commentsTable)

	tx, err := c.postgres.Begin(ctx)
	if err != nil {
		return internalErr
	}
	defer tx.Rollback(ctx)

	_, err = tx.Exec(ctx, query, comment.Text, comment.Id)
	if err != nil {
		return internalErr
	}

	if err = tx.Commit(ctx); err != nil {
		return internalErr
	}

	err = c.redis.Del(ctx, getRedisKey(comment.Id)).Err()
	if err != nil {
		return internalErr
	}

	err = c.redis.Set(ctx, getRedisKey(comment.Id), &comment, redisExpires).Err()
	if err != nil {
		return internalErr
	}

	return nil
}

func (c *Comment) DeleteComment(ctx context.Context, id uint64) *e.Error {
	query := fmt.Sprintf("DELETE FROM %s WHERE id = $1;", commentsTable)

	tx, err := c.postgres.Begin(ctx)
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

	err = c.redis.Del(ctx, getRedisKey(id)).Err()
	if err != nil {
		return internalErr
	}

	return nil
}

func getRedisKey(id uint64) string {
	return fmt.Sprintf("comments:%d", id)
}