package controllers

import (
	"github.com/nikitaSstepanov/p2p-streaming-service/backend/internal/pkg/controllers/http/v1/handlers"
	"github.com/nikitaSstepanov/p2p-streaming-service/backend/internal/pkg/usecases/services"
	"github.com/gin-gonic/gin"
)

type Controller struct {
	handlers *handlers.Handlers
}

func New(services *services.Services) *Controller {
	return &Controller{
		handlers: handlers.New(services),
	}
}

func (c *Controller) InitRoutes() *gin.Engine {
	router := gin.New()

	router.GET("ping", c.handlers.Default.Ping)
	
	api := router.Group("/api/v1")
	{
		movies := api.Group("/movies")
		{
			movies.GET("/", c.handlers.Movies.GetMovies)
			movies.GET("/:id", c.handlers.Movies.GetMovieById)
			movies.GET("/:id/start", c.handlers.Movies.StartWatch)
			movies.GET("/:id/:fileId/:chunkId", c.handlers.Movies.GetMovieChunck)

			comments := movies.Group("/:id/comments")
			{
				comments.GET("/", c.handlers.Comments.GetComments)
				comments.POST("/new", c.handlers.Comments.CreateComment)
				comments.PATCH("/:commentId/edit", c.handlers.Comments.EditComment)
				comments.DELETE("/:commentId/del", c.handlers.Comments.DeleteComment)
			}
		}

		account := api.Group("/account")
		{
			account.GET("/", c.handlers.Account.GetAccount)
			account.POST("/new", c.handlers.Account.Create)
			account.POST("/sign-in", c.handlers.Account.SignIn)

			playlists := account.Group("/playlists")
			{
				playlists.GET("/", c.handlers.Playlists.GetPlaylists)
				playlists.GET("/:id", c.handlers.Playlists.GetPlaylistById)
				playlists.POST("/new", c.handlers.Playlists.CreatePlaylist)
				playlists.PATCH("/:id/edit", c.handlers.Playlists.EditPlaylist)
				playlists.DELETE("/:id/del", c.handlers.Playlists.DeletePlaylist)
			}
		}

		admin := api.Group("/admin")
		{
			movies := admin.Group("/movies") 
			{
				movies.POST("/new", c.handlers.Admin.CreateMovie)
				movies.PATCH("/edit", c.handlers.Admin.EditMovie)
			}

			admins := admin.Group("/admins")
			{
				admins.GET("/", c.handlers.Admin.GetAdmins)
				admins.PATCH("/add", c.handlers.Admin.AddAdmin)
				admins.PATCH("/remove", c.handlers.Admin.RemoveAdmin)
			}
		}
	}

	return router
}
