package app

import (
	"os/signal"
	"log/slog"
	"context"
	"syscall"
	"fmt"
	"os"

	"github.com/nikitaSstepanov/p2p-streaming-service/backend/internal/config"
	"github.com/nikitaSstepanov/p2p-streaming-service/backend/internal/pkg/controllers"
	"github.com/nikitaSstepanov/p2p-streaming-service/backend/internal/pkg/services"
	"github.com/nikitaSstepanov/p2p-streaming-service/backend/internal/pkg/storage"
	"github.com/nikitaSstepanov/p2p-streaming-service/backend/libs/logger"
	"github.com/nikitaSstepanov/p2p-streaming-service/backend/libs/postgresql"
	"github.com/nikitaSstepanov/p2p-streaming-service/backend/libs/server"
	"github.com/nikitaSstepanov/p2p-streaming-service/backend/migrations"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"github.com/spf13/viper"
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

	if err := config.Init(); err != nil {
		slog.Error("Can't init config Error:", err)
	}

	ctx := context.TODO()

	db, err := postgresql.ConnectToDb(ctx, postgresql.Config{
		Username: viper.GetString("db.username"),
		Password: os.Getenv("POSTGRES_PASSWORD"),
		DBName:   viper.GetString("db.dbname"),
		Port:     viper.GetString("db.port"),
		Host:     viper.GetString("db.host"),
		SSLMode:  viper.GetString("db.sslmode"),
	})

	if err != nil {
		slog.Error("Can`t connect to db. Error:", err)
	} else {
		slog.Info("Db is connected.")
	}

	if err := migrations.Migrate(db); err != nil {
		slog.Error("Can`t migrate db scheme. Error:", err)
	}

	url := viper.GetString("url")

	app := &App{}

	app.Storage = storage.New(db)

	app.Services = services.New(app.Storage)

	app.Controller = controllers.New(app.Services)

	handler := app.Controller.InitRoutes()

	app.Server = server.New(handler, url)

	return app
}

func (a *App) Run() error {
	if err := a.Server.Start(); err != nil {
		slog.Error("Can`t run application. Error:", err)
	}

	fmt.Println("Application is running.")

	err := ShutdownApp(a)

	if err != nil {
		slog.Error("Application shutdown error. Error:", err)
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
