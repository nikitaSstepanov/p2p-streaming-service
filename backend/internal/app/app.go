package app

import (
	"context"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"github.com/nikitaSstepanov/p2p-streaming-service/backend/internal/pkg/controllers"
	"github.com/nikitaSstepanov/p2p-streaming-service/backend/internal/pkg/services"
	"github.com/nikitaSstepanov/p2p-streaming-service/backend/internal/pkg/storage"
	"github.com/nikitaSstepanov/p2p-streaming-service/backend/libs/postgresql"
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

	ctx := context.TODO()

	db := postgresql.ConnectToDb(ctx, postgresql.Config{
		Username: os.Getenv("POSTGRES_USERNAME"),
		Password: os.Getenv("POSTGRES_PASSWORD"),
		DBName: os.Getenv("POSTGRES_DATABASE"),
		Port: os.Getenv("POSTGRES_PORT"),
		Host: os.Getenv("POSTGRES_HOST"),
		SSLMode: os.Getenv("POSTGRES_SSl"),
	})

	app.Storage = storage.New(db)

	app.Services = services.New(app.Storage)

	app.Controller = controllers.New(app.Services)

	handler := app.Controller.InitRoutes()

	app.Server = server.New(handler, port)

	return app
}

func (a *App) Run() {
	a.Server.Start()
}
