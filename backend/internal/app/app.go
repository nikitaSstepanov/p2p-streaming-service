package app

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"github.com/nikitaSstepanov/p2p-streaming-service/backend/internal/pkg/controllers"
	"github.com/nikitaSstepanov/p2p-streaming-service/backend/internal/pkg/services"
	"github.com/nikitaSstepanov/p2p-streaming-service/backend/internal/pkg/storage"
	"github.com/nikitaSstepanov/p2p-streaming-service/backend/libs/logger"
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
	logger := logger.New()

	slog.SetDefault(logger.Logger)

	gin.SetMode(gin.TestMode)

	if err := godotenv.Load("../../.env"); err != nil {
		slog.Error("Can`t load env. Error:", err)
	}

	port := os.Getenv("URL")

	app := &App{}

	ctx := context.TODO()

	db, err := postgresql.ConnectToDb(ctx, postgresql.Config{
		Username: os.Getenv("POSTGRES_USERNAME"),
		Password: os.Getenv("POSTGRES_PASSWORD"),
		DBName:   os.Getenv("POSTGRES_DATABASE"),
		Port:     os.Getenv("POSTGRES_PORT"),
		Host:     os.Getenv("POSTGRES_HOST"),
		SSLMode:  os.Getenv("POSTGRES_SSl"),
	})

	if err != nil {
		slog.Error("Can`t connect to db. Error:", err)
	} else {
		slog.Info("Db is connected.")
	}

	app.Storage = storage.New(db)

	app.Services = services.New(app.Storage)

	app.Controller = controllers.New(app.Services)

	handler := app.Controller.InitRoutes()

	app.Server = server.New(handler, port)

	return app
}

func (a *App) Run() error {
	if err := a.Server.Start(); err != nil {
		return err
	}

	err := ShutdownApp(a)

	if err != nil {
		return err
	}

	slog.Info("Application Shutting down.")

	return nil
}

func ShutdownApp(a *App) error {
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGTERM, syscall.SIGINT)
	<-quit

	if err := a.Server.Shutdown(context.Background()); err != nil {
		return err
	}
	
	return nil
}
