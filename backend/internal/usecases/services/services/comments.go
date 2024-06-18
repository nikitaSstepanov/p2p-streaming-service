package services

import (
	"context"

	"github.com/nikitaSstepanov/p2p-streaming-service/backend/internal/usecases/storage/entities"
	"github.com/nikitaSstepanov/p2p-streaming-service/backend/internal/types/dto/comments"
	"github.com/nikitaSstepanov/p2p-streaming-service/backend/internal/usecases/storage"
	"github.com/nikitaSstepanov/p2p-streaming-service/backend/internal/types/responses"
	"github.com/nikitaSstepanov/p2p-streaming-service/backend/internal/types/statuses"
	"github.com/redis/go-redis/v9"
)

type Comments struct {
	storage *storage.Storage
	redis   *redis.Client
	account *Account
}

func NewComments(storage *storage.Storage, redis *redis.Client, account *Account) *Comments {
	return &Comments{
		storage: storage,
		redis: redis,
		account: account,
	}
}

func (c *Comments) GetComments(ctx context.Context, movieId string, limit string, offset string) (*[]dto.CommentDto, string) {
	comments, err := c.storage.Comments.GetComments(ctx, movieId, limit, offset)

	if err != nil {
		return nil, statuses.InternalError
	}

	var result []dto.CommentDto

	if len(*comments) != 0 {
		for i := 0; i < len(*comments); i++ {
			comment := (*comments)[i]
			toAdd := dto.CommentDto{
				Id: comment.Id,
				UserId: comment.UserId,
				Text: comment.Text,
			}
			result = append(result, toAdd)
		}
	} else {
		result = make([]dto.CommentDto, 0)
	}

	return &result, statuses.OK
}

func (c *Comments) CreateComment(ctx context.Context, header string, movieId string, body *dto.CreateCommentDto) (*responses.Message, string) {
	user, status := c.account.getUser(ctx, header)

	if status != statuses.OK {
		return nil, status
	}

	movie, err := c.storage.Movies.GetMovieById(ctx, movieId)

	if err != nil {
		return nil, statuses.InternalError
	}

	if movie.Id == 0 {
		return nil, statuses.NotFound
	}

	comment := &entities.Comment{
		MovieId: movie.Id,
		UserId: user.Id,
		Text: body.Text,
	}

	c.storage.Comments.CreateComment(ctx, comment)

	result := &responses.Message{
		Message: "Created.",
	}

	return result, statuses.OK
}

func (c *Comments) EditComment(ctx context.Context, header string, movieId string, commentId string, body *dto.UpdateCommentDto) (*responses.Message, string) {
	user, status := c.account.getUser(ctx, header)

	if status != statuses.OK {
		return nil, status
	}

	movie, err := c.storage.Movies.GetMovieById(ctx, movieId)

	if err != nil {
		return nil, statuses.InternalError
	}

	if movie.Id == 0 {
		return nil, statuses.NotFound
	}

	comment, err := c.storage.Comments.GetCommentById(ctx, commentId)

	if err != nil {
		return nil, statuses.InternalError
	}

	if comment.Id == 0 {
		return nil, statuses.NotFound
	}

	if user.Id != comment.UserId {
		return nil, statuses.Forbidden
	}

	if len(body.Text) != 0 {
		comment.Text = body.Text
	}

	c.storage.Comments.UpdateComment(ctx, comment)

	result := &responses.Message{
		Message: "Updated.",
	}

	return result, statuses.OK
}

func (c *Comments) DeleteComment(ctx context.Context, header string, commentId string) (*responses.Message, string) {
	user, status := c.account.getUser(ctx, header)

	if status != statuses.OK {
		return nil, status
	}

	comment, err := c.storage.Comments.GetCommentById(ctx, commentId)

	if err != nil {
		return nil, statuses.InternalError
	}

	if comment.Id == 0 {
		return nil, statuses.NotFound
	}

	if user.Id != comment.UserId {
		return nil, statuses.Forbidden
	}

	c.storage.Comments.DeleteComment(ctx, commentId)

	result := &responses.Message{
		Message: "Deleted.",
	}

	return result, statuses.OK
}