package controllers

import (
	"net/http"

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

	router.GET("ping", pingFunc)
	
	api := router.Group("/api")
	{
		movies := api.Group("/movies")
		{
			movies.GET("/", c.Services.Movies.GetMovies)
			movies.GET("/:id", c.Services.Movies.GetMovieById)
			movies.GET("/:id/start", c.Services.Movies.StartWatch)
			movies.GET("/:id/:fileId/:chunkId", c.Services.Movies.GetMovieChunck)
		}

		account := api.Group("/account")
		{
			account.GET("/", c.Services.Account.GetAccount)
			account.POST("/new", c.Services.Account.Create)
			account.POST("/sign-in", c.Services.Account.SignIn)
		}

		admin := api.Group("/admin")
		{
			admin.POST("/add-admin", c.Services.Admin.AddAdmin)

			movies := admin.Group("/movies") 
			{
				movies.POST("/new", c.Services.Admin.CreateMovie)
			}
		}
	}

	return router
}

func pingFunc(ctx *gin.Context) {
	ctx.JSON(http.StatusOK, "ok")
}