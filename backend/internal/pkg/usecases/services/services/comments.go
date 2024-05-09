package services

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/nikitaSstepanov/p2p-streaming-service/backend/internal/pkg/types/dto/comments"
	"github.com/nikitaSstepanov/p2p-streaming-service/backend/internal/pkg/types/responses"
	"github.com/nikitaSstepanov/p2p-streaming-service/backend/internal/pkg/types/statuses"
	"github.com/nikitaSstepanov/p2p-streaming-service/backend/internal/pkg/usecases/storage"
	"github.com/nikitaSstepanov/p2p-streaming-service/backend/internal/pkg/usecases/storage/entities"
	"github.com/redis/go-redis/v9"
)

type Comments struct {
	storage *storage.Storage
	redis   *redis.Client
	auth    *Auth
}

func NewComments(storage *storage.Storage, redis *redis.Client, auth *Auth) *Comments {
	return &Comments{
		storage: storage,
		redis: redis,
		auth: auth,
	}
}

func (c *Comments) GetComments(ctx context.Context, movieId string, limit string, offset string) (*[]dto.CommentDto, string) {
	comments := c.storage.Comments.GetComments(ctx, movieId, limit, offset)

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
	user, status := c.getUser(ctx, header)

	if status != statuses.OK {
		return nil, status
	}

	movie := c.storage.Movies.GetMovieById(ctx, movieId)

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
	user, status := c.getUser(ctx, header)

	if status != statuses.OK {
		return nil, status
	}

	movie := c.storage.Movies.GetMovieById(ctx, movieId)

	if movie.Id == 0 {
		return nil, statuses.NotFound
	}

	comment := c.storage.Comments.GetCommentById(ctx, commentId)

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
	user, status := c.getUser(ctx, header)

	if status != statuses.OK {
		return nil, status
	}

	comment := c.storage.Comments.GetCommentById(ctx, commentId)

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

func (c *Comments) getUser(ctx context.Context, header string) (*entities.User, string) {
	parts := strings.Split(header, " ")
	bearer := parts[0]
	token := parts[1]

	if bearer != "Bearer" {
		return nil, statuses.Unauthorize
	}

	claims, err := c.auth.ValidateToken(token)

	if err != nil {
		return nil, statuses.Unauthorize
	}

	user := c.findUser(ctx, claims.Username)

	if user.Id == 0 {
		return nil, statuses.Unauthorize
	}

	return &user, statuses.OK
}

func (c *Comments) findUser(ctx context.Context, username string) entities.User {
	var user entities.User

	c.redis.Get(ctx, fmt.Sprintf("users:%s", username)).Scan(&user)

	if user.Id == 0 {
		user = *(c.storage.Users.GetUser(ctx, username))

		if user.Id == 0 {
			return entities.User{}
		}

		c.redis.Set(ctx, fmt.Sprintf("users:%s", username), user, 1 * time.Hour)
	}

	return user
}