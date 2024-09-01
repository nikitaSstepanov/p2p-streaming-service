package comment

import (
	"context"

	"github.com/nikitaSstepanov/p2p-streaming-service/backend/internal/entity"
	e "github.com/nikitaSstepanov/p2p-streaming-service/backend/pkg/errors"
)

var (
	forbiddenErr = e.New("Forbidden.", e.Forbidden)
)

type CommentStorage interface {
	GetComments(ctx context.Context, movieId uint64, limit int, offset int) ([]*entity.Comment, *e.Error)
	GetCommentById(ctx context.Context, id uint64) (*entity.Comment, *e.Error)
	CreateComment(ctx context.Context, comment *entity.Comment) *e.Error
	UpdateComment(ctx context.Context, comment *entity.Comment) *e.Error
	DeleteComment(ctx context.Context, id uint64) *e.Error
}

type MovieStorage interface {
	GetMovieById(ctx context.Context, id uint64) (*entity.Movie, *e.Error)
}

type Comment struct {
	commentsStorage CommentStorage
	movieStorage    MovieStorage
}

func New(comments CommentStorage, movies MovieStorage) *Comment {
	return &Comment{
		commentsStorage: comments,
		movieStorage:    movies,
	}
}

func (c *Comment) GetComments(ctx context.Context, movieId uint64, limit int, offset int) ([]*entity.Comment, *e.Error) {
	return c.commentsStorage.GetComments(ctx, movieId, limit, offset)
}

func (c *Comment) CreateComment(ctx context.Context, comment *entity.Comment) *e.Error {
	_, err := c.movieStorage.GetMovieById(ctx, comment.MovieId)
	if err != nil {
		return err
	}

	return c.commentsStorage.CreateComment(ctx, comment)
}

func (c *Comment) EditComment(ctx context.Context, updated *entity.Comment) *e.Error {
	_, err := c.movieStorage.GetMovieById(ctx, updated.MovieId)
	if err != nil {
		return err
	}

	comment, err := c.commentsStorage.GetCommentById(ctx, updated.Id)
	if err != nil {
		return err
	}

	if updated.UserId != comment.UserId {
		return forbiddenErr
	}

	if len(updated.Text) != 0 {
		comment.Text = updated.Text
	}

	return c.commentsStorage.UpdateComment(ctx, comment)
}

func (c *Comment) DeleteComment(ctx context.Context, comment *entity.Comment) *e.Error {
	toDel, err := c.commentsStorage.GetCommentById(ctx, comment.Id)
	if err != nil {
		return err
	}

	if toDel.UserId != comment.UserId{
		return forbiddenErr
	}

	return c.commentsStorage.DeleteComment(ctx, toDel.Id)
}