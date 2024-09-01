package comment

import "github.com/gin-gonic/gin"

type CommentHandler interface {
	GetComments(ctx *gin.Context)
	CreateComment(ctx *gin.Context)
	EditComment(ctx *gin.Context)
	DeleteComment(ctx *gin.Context)
}

type Middleware interface {
	CheckAccess(roles ...string) gin.HandlerFunc
}

func InitRoutes(handler *gin.RouterGroup, comment CommentHandler, mid Middleware) *gin.RouterGroup {
	router := handler.Group("/:id/comments")
	{
		router.GET("/", comment.GetComments)
		router.POST("/new", mid.CheckAccess(), comment.CreateComment)
		router.PATCH("/:commentId/edit", mid.CheckAccess(), comment.EditComment)
		router.DELETE("/:commentId/del", mid.CheckAccess(), comment.DeleteComment)
	}

	return router
}