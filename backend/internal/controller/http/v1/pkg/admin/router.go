package admin

import "github.com/gin-gonic/gin"

type AdminHandler interface {
	GetAdmins(ctx *gin.Context)
	AddAdmin(ctx *gin.Context)
	RemoveAdmin(ctx *gin.Context)
	CreateMovie(ctx *gin.Context)
	EditMovie(ctx *gin.Context)
}

type Middleware interface {
	CheckAccess(roles ...string) gin.HandlerFunc
}

func InitRoutes(handler *gin.RouterGroup, admin AdminHandler, mid Middleware) *gin.RouterGroup {
	router := handler.Group("/admin")
	{
		movies := router.Group("/movies")

		movies.Use(mid.CheckAccess("ADMIN"))
		{
			movies.POST("/new", admin.CreateMovie)
			movies.PATCH("/edit", admin.EditMovie)
		}

		admins := router.Group("/admins")

		admins.Use(mid.CheckAccess("SUPER_ADMIN"))
		{
			admins.GET("/", admin.GetAdmins)
			admins.PATCH("/add", admin.AddAdmin)
			admins.PATCH("/remove", admin.RemoveAdmin)
		}
	}

	return router
}