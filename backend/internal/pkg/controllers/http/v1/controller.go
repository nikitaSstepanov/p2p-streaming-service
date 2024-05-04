package controllers

import (
	"github.com/nikitaSstepanov/p2p-streaming-service/backend/internal/pkg/services"
	"github.com/gin-gonic/gin"
)

type Controller struct {
	Services *services.Services
}

func New(services *services.Services) *Controller {
	return &Controller{
		Services: services,
	}
}

func (c *Controller) InitRoutes() *gin.Engine {
	router := gin.New()

	router.GET("ping", c.Services.PingFunc)
	
	api := router.Group("/api/v1")
	{
		movies := api.Group("/movies")
		{
			movies.GET("/", c.Services.Movies.GetMovies)
			movies.GET("/:id", c.Services.Movies.GetMovieById)
			movies.GET("/:id/start", c.Services.Movies.StartWatch)
			movies.GET("/:id/:fileId/:chunkId", c.Services.Movies.GetMovieChunck)

			comments := movies.Group("/:id/comments")
			{
				comments.GET("/", c.Services.Comments.GetComments)
				comments.GET("/:commentId", c.Services.Comments.GetCommentById)
				comments.POST("/new", c.Services.Comments.CreateComment)
				comments.PATCH("/:commentId/edit", c.Services.Comments.EditComment)
				comments.DELETE("/:commentId/del", c.Services.Comments.DeleteComment)
			}
		}

		account := api.Group("/account")
		{
			account.GET("/", c.Services.Account.GetAccount)
			account.POST("/new", c.Services.Account.Create)
			account.POST("/sign-in", c.Services.Account.SignIn)

			playlists := account.Group("/playlists")
			{
				playlists.GET("/", c.Services.Playlists.GetPlaylists)
				playlists.GET("/:id", c.Services.Playlists.GetPlaylistById)
				playlists.POST("/new", c.Services.Playlists.CreatePlaylist)
				playlists.PATCH("/:id/edit", c.Services.Playlists.EditPlaylist)
				playlists.DELETE("/:id/del", c.Services.Playlists.DeletePlaylist)
			}
		}

		admin := api.Group("/admin")
		{
			movies := admin.Group("/movies") 
			{
				movies.POST("/new", c.Services.Admin.CreateMovie)
				movies.PATCH("/edit", c.Services.Admin.EditMovie)
			}

			admins := admin.Group("/admins")
			{
				admins.GET("/", c.Services.Admin.GetAdmins)
				admins.PATCH("/add", c.Services.Admin.AddAdmin)
				admins.PATCH("/remove", c.Services.Admin.RemoveAdmin)
			}
		}
	}

	return router
}
