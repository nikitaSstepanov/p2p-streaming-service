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
	
	api := router.Group("/api")
	{
		movies := api.Group("/movies")
		{
			movies.GET("/", c.Services.Movies.GetMovies)
			movies.GET("/:id", c.Services.Movies.GetMovieById)
		}

		account := api.Group("/account")
		{
			account.POST("/new", c.Services.Account.Create)
			account.POST("/sign-in", c.Services.Account.SignIn)
		}
	}

	return router
}
