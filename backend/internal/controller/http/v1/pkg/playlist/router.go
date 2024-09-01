package playlist

import "github.com/gin-gonic/gin"

type PlaylistHandler interface {
	GetPlaylists(ctx *gin.Context)
	GetPlaylistById(ctx *gin.Context)
	CreatePlaylist(ctx *gin.Context) 
	EditPlaylist(ctx *gin.Context)
	DeletePlaylist(ctx *gin.Context)
}

type Middleware interface {
	CheckAccess(roles ...string) gin.HandlerFunc
}

func InitRoutes(handler *gin.RouterGroup, playlist PlaylistHandler, mid Middleware) *gin.RouterGroup {
	router := handler.Group("/playlists")

	router.Use(mid.CheckAccess())
	{
		router.GET("/", playlist.GetPlaylists)
		router.GET("/:id", playlist.GetPlaylistById)
		router.POST("/new", playlist.CreatePlaylist)
		router.PATCH("/:id/edit", playlist.EditPlaylist)
		router.DELETE("/:id/del", playlist.DeletePlaylist)
	}

	return router
}