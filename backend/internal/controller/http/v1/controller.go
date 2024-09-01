package controllers

import (
	"github.com/gin-gonic/gin"
	"github.com/nikitaSstepanov/p2p-streaming-service/backend/internal/controller/http/v1/middleware"
	"github.com/nikitaSstepanov/p2p-streaming-service/backend/internal/controller/http/v1/pkg/account"
	"github.com/nikitaSstepanov/p2p-streaming-service/backend/internal/controller/http/v1/pkg/admin"
	"github.com/nikitaSstepanov/p2p-streaming-service/backend/internal/controller/http/v1/pkg/auth"
	"github.com/nikitaSstepanov/p2p-streaming-service/backend/internal/controller/http/v1/pkg/comment"
	"github.com/nikitaSstepanov/p2p-streaming-service/backend/internal/controller/http/v1/pkg/movie"
	"github.com/nikitaSstepanov/p2p-streaming-service/backend/internal/controller/http/v1/pkg/playlist"
	"github.com/nikitaSstepanov/p2p-streaming-service/backend/internal/usecase"
)

type Controller struct {
	account   *account.Account
	admin     *admin.Admin
	auth      *auth.Auth
	comment   *comment.Comment
	movie     *movie.Movies
	playlist  *playlist.Playlist
	middleware *middleware.Middleware
}

func New(uc *usecase.UseCase) *Controller {
	return &Controller{
		account:  account.New(uc.Accounts),
		admin:    admin.New(uc.Admin),
		auth:     auth.New(uc.Auth),
		comment:  comment.New(uc.Comment),
		movie:    movie.New(uc.Movies),
		playlist: playlist.New(uc.Playlist),
	}
}

func (c *Controller) InitRoutes() *gin.Engine {
	router := gin.New()
	
	api := router.Group("/api/v1")
	{
		movieGroup := movie.InitRoutes(api, c.movie)

		comment.InitRoutes(movieGroup, c.comment, c.middleware)

		accountGroup := account.InitRoutes(api, c.account, c.middleware)

		auth.InitRoutes(accountGroup, c.auth, c.middleware)

		playlist.InitRoutes(accountGroup, c.playlist, c.middleware)

		admin.InitRoutes(api, c.admin, c.middleware)
	}

	return router
}