package repositories

import (
	"context"
	"fmt"

	"github.com/nikitaSstepanov/p2p-streaming-service/backend/internal/usecases/storage/entities"
	"github.com/nikitaSstepanov/p2p-streaming-service/backend/pkg/postgresql"
)

const (
	commentsTable = "comments"
)

type Comments struct {
	db postgresql.Client
}

func NewComments(db postgresql.Client) *Comments {
	return &Comments{
		db: db,
	}
}

func (c *Comments) GetComments(ctx context.Context, movieId string, limit string, offset string) (*[]entities.Comment, error) {
	var comments []entities.Comment

	var comment entities.Comment

	query := fmt.Sprintf("SELECT * FROM %s WHERE movieId = %s LIMIT %s OFFSET %s;", commentsTable, movieId, limit, offset)

	rows, err := c.db.Query(ctx, query)

	if err != nil {
		return nil, err
	}

	for rows.Next() {
		err = rows.Scan(&comment.Id, &comment.MovieId, &comment.UserId, &comment.Text)

		if err != nil {
			return nil, err
		}

		comments = append(comments, comment)
	}

	return &comments, nil
}

func (c *Comments) GetCommentById(ctx context.Context, id string) (*entities.Comment, error) {
	var comment entities.Comment

	query := fmt.Sprintf("SELECT * FROM %s WHERE id = %s;", commentsTable, id)

	row := c.db.QueryRow(ctx, query)

	err := row.Scan(&comment.Id, &comment.MovieId, &comment.UserId, &comment.Text)

	if err != nil {
		return nil, err
	}

	return &comment, nil
}

func (c *Comments) CreateComment(ctx context.Context, comment *entities.Comment) error {
	query := fmt.Sprintf("INSERT INTO %s (movieId, userId, text) VALUES ($1, $2, $3);", commentsTable)

	tx, err := c.db.Begin(ctx)

	if err != nil {
		return err
	}

	_, err = tx.Exec(ctx, query, comment.MovieId, comment.UserId, comment.Text)

	if err != nil {
		return err
	}

	err = tx.Commit(ctx)

	if err != nil {
		return err
	}

	return nil
}

func (c *Comments) UpdateComment(ctx context.Context, comment *entities.Comment) error {
	query := fmt.Sprintf("UPDATE %s SET text = $1 WHERE id = $2;", commentsTable)

	tx, err := c.db.Begin(ctx)

	if err != nil {
		return err
	}

	_, err = tx.Exec(ctx, query, comment.Text, comment.Id)

	if err != nil {
		return err
	}

	err = tx.Commit(ctx)

	if err != nil {
		return err
	}

	return nil
}

func (c *Comments) DeleteComment(ctx context.Context, id string) error {
	query := fmt.Sprintf("DELETE FROM %s WHERE id = $1;", commentsTable)

	tx, err := c.db.Begin(ctx)

	if err != nil {
		return err
	}

	_, err = tx.Exec(ctx, query, id)

	if err != nil {
		return err
	}

	err = tx.Commit(ctx)

	if err != nil {
		return err
	}

	return nil
}