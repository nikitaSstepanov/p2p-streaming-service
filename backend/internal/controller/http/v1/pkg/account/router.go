package account

import "github.com/gin-gonic/gin"

type AccountHandler interface {
	GetAccount(ctx *gin.Context)
	Create(ctx *gin.Context)
}

type Middleware interface {
	CheckAccess(roles ...string) gin.HandlerFunc
}

func InitRoutes(handler *gin.RouterGroup, acc AccountHandler, mid Middleware) *gin.RouterGroup {
	router := handler.Group("/account")
	{
		router.GET("/", mid.CheckAccess(), acc.GetAccount)
		router.POST("/new", acc.Create)
	}

	return router
}