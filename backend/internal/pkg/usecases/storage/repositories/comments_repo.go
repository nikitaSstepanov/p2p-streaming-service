package repositories

import (
	"context"
	"fmt"

	"github.com/nikitaSstepanov/p2p-streaming-service/backend/internal/pkg/usecases/storage/entities"
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

func (c *Comments) GetComments(ctx context.Context, movieId string, limit string, offset string) *[]entities.Comment {
	var comments []entities.Comment

	var comment entities.Comment

	query := fmt.Sprintf("SELECT * FROM %s WHERE movieId = %s LIMIT %s OFFSET %s;", commentsTable, movieId, limit, offset)

	rows, _ := c.db.Query(ctx, query)

	for rows.Next() {
		rows.Scan(&comment.Id, &comment.MovieId, &comment.UserId, &comment.Text)

		comments = append(comments, comment)
	}

	return &comments
}

func (c *Comments) GetCommentById(ctx context.Context, id string) *entities.Comment {
	var comment entities.Comment

	query := fmt.Sprintf("SELECT * FROM %s WHERE id = %s;", commentsTable, id)

	row := c.db.QueryRow(ctx, query)

	row.Scan(&comment.Id, &comment.MovieId, &comment.UserId, &comment.Text)

	return &comment
}

func (c *Comments) CreateComment(ctx context.Context, comment *entities.Comment) {
	query := fmt.Sprintf("INSERT INTO %s (movieId, userId, text) VALUES (%d, %d, '%s');", commentsTable, comment.MovieId, comment.UserId, comment.Text)

	c.db.QueryRow(ctx, query)
}

func (c *Comments) UpdateComment(ctx context.Context, comment *entities.Comment) {
	query := fmt.Sprintf("UPDATE %s SET text = '%s' WHERE id = %d;", commentsTable, comment.Text, comment.Id)

	c.db.QueryRow(ctx, query)
}

func (c *Comments) DeleteComment(ctx context.Context, id string) {
	query := fmt.Sprintf("DELETE FROM %s WHERE id = %s;", commentsTable, id)

	c.db.QueryRow(ctx, query)
}