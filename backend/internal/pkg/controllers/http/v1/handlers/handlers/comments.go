package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/nikitaSstepanov/p2p-streaming-service/backend/internal/pkg/types/dto/comments"
	"github.com/nikitaSstepanov/p2p-streaming-service/backend/internal/pkg/types/responses"
	"github.com/nikitaSstepanov/p2p-streaming-service/backend/internal/pkg/types/statuses"
	"github.com/nikitaSstepanov/p2p-streaming-service/backend/internal/pkg/usecases/services"
)

type Comments struct {
	services *services.Services
}

func NewComments(services *services.Services) *Comments {
	return &Comments{
		services: services,
	}
}

func (c *Comments) GetComments(ctx *gin.Context) {
	limit := ctx.DefaultQuery("limit", "5")

	offset := ctx.DefaultQuery("offset", "0")

	id := ctx.Param("id")

	result, status := c.services.Comments.GetComments(ctx, id, limit, offset)

	if status == statuses.OK {
		ctx.JSON(http.StatusOK, result)
	} else {
		c.handleError(ctx, status)
	}
}

func (c *Comments) CreateComment(ctx *gin.Context) {
	header := ctx.GetHeader("Authorization")

	id := ctx.Param("id")

	var data dto.CreateCommentDto

	if err := ctx.ShouldBindJSON(&data); err != nil {
		ctx.JSON(http.StatusBadRequest, "Incorrect data.")
		return
	}

	result, status := c.services.Comments.CreateComment(ctx, header, id, &data)

	if status == statuses.OK {
		ctx.JSON(http.StatusOK, result)
	} else {
		c.handleError(ctx, status)
	}
}

func (c *Comments) EditComment(ctx *gin.Context) {
	header := ctx.GetHeader("Authorization")

	id := ctx.Param("id")
	
	commentId := ctx.Param("commentId")

	var body dto.UpdateCommentDto

	if err := ctx.ShouldBindJSON(&body); err != nil {
		ctx.JSON(http.StatusBadRequest, "Incorrect data.")
		return
	}

	result, status := c.services.Comments.EditComment(ctx, header, id, commentId, &body)

	if status == statuses.OK {
		ctx.JSON(http.StatusOK, result)
	} else {
		c.handleError(ctx, status)
	}
}

func (c *Comments) DeleteComment(ctx *gin.Context) {
	header := ctx.GetHeader("Authorization")
	
	commentId := ctx.Param("commentId")

	result, status := c.services.Comments.DeleteComment(ctx, header, commentId)

	if status == statuses.OK {
		ctx.JSON(http.StatusOK, result)
	} else {
		c.handleError(ctx, status)
	}
}

func (c *Comments) handleError(ctx *gin.Context, status string) {
	e := &responses.Error{}

	switch status {

	case statuses.NotFound:
		e.Error = "This movie wasn`t found."
		ctx.JSON(http.StatusNotFound, e)

	case statuses.Unauthorize:
		e.Error = "Incorrect token."
		ctx.JSON(http.StatusUnauthorized, e)

	case statuses.Forbidden:
		e.Error = "Forbidden."
		ctx.JSON(http.StatusForbidden, e)

	case statuses.InternalError:
		e.Error = "Something going wrong....."
		ctx.JSON(http.StatusInternalServerError, e)

	}
}