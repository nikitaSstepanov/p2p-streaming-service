package app

import (
	"os"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"github.com/nikitaSstepanov/p2p-streaming-service/backend/internal/pkg/controllers"
	"github.com/nikitaSstepanov/p2p-streaming-service/backend/internal/pkg/services"
	"github.com/nikitaSstepanov/p2p-streaming-service/backend/internal/pkg/storage"
	"github.com/nikitaSstepanov/p2p-streaming-service/backend/libs/server"
)

type App struct {
	Controller *controllers.Controller
	Services   *services.Services
	Storage    *storage.Storage
	Server     *server.Server
}

func New() *App {

	gin.SetMode(gin.TestMode)

	godotenv.Load("../../.env")

	port := os.Getenv("URL")

	app := &App{}

	app.Storage = storage.New()

	app.Services = services.New(app.Storage)

	app.Controller = controllers.New(app.Services)

	handler := app.Controller.InitRoutes()

	app.Server = server.New(handler, port)

	return app
}

func (a *App) Run() {
	a.Server.Start()
}
