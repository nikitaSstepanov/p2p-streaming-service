package controllers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/nikitaSstepanov/p2p-streaming-service/backend/internal/pkg/services"
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
			movies.GET("/:id/:chunkId", c.Services.Movies.GetMovieChunck)
		}

		account := api.Group("/account")
		{
			account.POST("/new", c.Services.Account.Create)
			account.POST("/sign-in", c.Services.Account.SignIn)
		}
	}

	return router
}

func pingFunc(ctx *gin.Context) {
	ctx.JSON(http.StatusOK, "ok")
}