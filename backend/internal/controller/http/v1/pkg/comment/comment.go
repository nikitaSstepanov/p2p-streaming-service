package comment

import (
	"context"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/nikitaSstepanov/p2p-streaming-service/backend/internal/controller/http/v1/dto"
	"github.com/nikitaSstepanov/p2p-streaming-service/backend/internal/controller/http/v1/responses"
	"github.com/nikitaSstepanov/p2p-streaming-service/backend/internal/entity"
	e "github.com/nikitaSstepanov/p2p-streaming-service/backend/pkg/errors"
)

const (
	ok = http.StatusOK
	badReq = http.StatusBadRequest
	created = http.StatusCreated
	deleted = http.StatusNoContent
)

var (
	badReqErr  = e.New("Incorrect data.", e.BadInput)

	createdMsg = responses.NewMessage("Created.")
	updatedMsg = responses.NewMessage("Updated.")
	deletedMsg = responses.NewMessage("Deleted.")
)

type CommentUseCase interface {
	GetComments(ctx context.Context, movieId uint64, limit int, offset int) ([]*entity.Comment, *e.Error)
	CreateComment(ctx context.Context, comment *entity.Comment) *e.Error
	EditComment(ctx context.Context, updated *entity.Comment) *e.Error
	DeleteComment(ctx context.Context, comment *entity.Comment) *e.Error
}

type Comment struct {
	usecase CommentUseCase
}

func New(usecase CommentUseCase) *Comment {
	return &Comment{
		usecase,
	}
}

func (c *Comment) GetComments(ctx *gin.Context) {
	id, err := strconv.ParseUint(ctx.Param("id"), 10, 64)
	if err != nil {
		ctx.AbortWithStatusJSON(badReq, badReqErr)
		return
	}

	limit, err := strconv.Atoi(ctx.DefaultQuery("limit", "5"))
	if err != nil {
		ctx.AbortWithStatusJSON(badReq, badReqErr)
		return
	}

	offset, err := strconv.Atoi(ctx.DefaultQuery("offset", "0"))
	if err != nil {
		ctx.AbortWithStatusJSON(badReq, badReqErr)
		return
	}

	comments, commErr := c.usecase.GetComments(ctx, id, limit, offset)
	if commErr != nil {
		ctx.AbortWithStatusJSON(commErr.ToHttpCode(), commErr)
		return
	}

	result := make([]dto.CommentDto, 0)

	for i := 0; i < len(comments); i++ {
		result = append(result, *dto.CommentToDto(comments[i]))
	}

	ctx.JSON(ok, result)
}

func (c *Comment) CreateComment(ctx *gin.Context) {
	userId := ctx.GetUint64("userId")

	movieId, err := strconv.ParseUint(ctx.Param("id"), 10, 64)
	if err != nil {
		ctx.AbortWithStatusJSON(badReq, badReqErr)
		return
	}

	var body dto.CreateCommentDto

	if err := ctx.ShouldBindJSON(&body); err != nil {
		ctx.AbortWithStatusJSON(badReq, badReqErr)
		return
	}

	comment := &entity.Comment{
		MovieId: movieId,
		UserId:  userId,
		Text:    body.Text,
	}

	commErr := c.usecase.CreateComment(ctx, comment)
	if commErr != nil {
		ctx.AbortWithStatusJSON(commErr.ToHttpCode(), commErr)
		return 
	}

	ctx.JSON(created, createdMsg)
}

func (c *Comment) EditComment(ctx *gin.Context) {
	userId := ctx.GetUint64("userId")

	movieId, err := strconv.ParseUint(ctx.Param("id"), 10, 64)
	if err != nil {
		ctx.AbortWithStatusJSON(badReq, badReqErr)
		return
	}

	id, err := strconv.ParseUint(ctx.Param("commentId"), 10, 64)
	if err != nil {
		ctx.AbortWithStatusJSON(badReq, badReqErr)
		return
	}

	var body dto.UpdateCommentDto

	if err := ctx.ShouldBindJSON(&body); err != nil {
		ctx.AbortWithStatusJSON(badReq, badReqErr)
		return
	}

	comment := &entity.Comment{
		Id: id,
		MovieId: movieId,
		UserId: userId,
		Text:   body.Text,
	}

	commErr := c.usecase.EditComment(ctx, comment)
	if commErr != nil {
		ctx.AbortWithStatusJSON(commErr.ToHttpCode(), commErr)
		return 
	} 

	ctx.JSON(ok, updatedMsg)
}

func (c *Comment) DeleteComment(ctx *gin.Context) {
	userId := ctx.GetUint64("userId")

	commentId, err := strconv.ParseUint(ctx.Param("commentId"), 10, 64)
	if err != nil {
		ctx.AbortWithStatusJSON(badReq, badReqErr)
		return
	}

	comment := &entity.Comment{
		Id: commentId,
		UserId: userId,
	}

	commErr := c.usecase.DeleteComment(ctx, comment)
	if commErr != nil {
		ctx.AbortWithStatusJSON(commErr.ToHttpCode(), commErr)
		return 
	}

	ctx.JSON(deleted, deletedMsg)
}