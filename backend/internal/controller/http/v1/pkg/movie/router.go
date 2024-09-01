package movie

import "github.com/gin-gonic/gin"

type MovieHandler interface {
	GetMovies(ctx *gin.Context)
	GetMovieById(ctx *gin.Context)
	StartWatch(ctx *gin.Context)
	GetMovieChunck(ctx *gin.Context)
}

func InitRoutes(handler *gin.RouterGroup, movie MovieHandler) *gin.RouterGroup {
	router := handler.Group("/movies")
	{
		router.GET("/", movie.GetMovies)
		router.GET("/:id", movie.GetMovieById)
		router.GET("/:id/start", movie.StartWatch)
		router.GET("/:id/:fileId/:chunkId", movie.GetMovieChunck)
	}

	return router
}