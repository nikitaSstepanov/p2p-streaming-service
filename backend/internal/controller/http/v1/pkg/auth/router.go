package auth

import "github.com/gin-gonic/gin"

type AuthHandler interface {
	Login(ctx *gin.Context)
	Logout(ctx *gin.Context)
	Refresh(ctx *gin.Context)
}

type Middleware interface {
	CheckAccess(roles ...string) gin.HandlerFunc
}

func InitRoutes(handler *gin.RouterGroup, auth AuthHandler, mid Middleware) *gin.RouterGroup {
	router := handler.Group("/auth")
	{
		router.POST("/login", auth.Login)
		router.POST("/logout", mid.CheckAccess(), auth.Logout)
		router.GET("/refresh", auth.Refresh)
	}

	return router
}