package app

import (
	"os/signal"
	"log/slog"
	"syscall"
	"context"
	"fmt"
	"os"

	"github.com/nikitaSstepanov/p2p-streaming-service/backend/internal/controllers/http/v1"
	"github.com/nikitaSstepanov/p2p-streaming-service/backend/internal/usecases/services"
	"github.com/nikitaSstepanov/p2p-streaming-service/backend/internal/usecases/storage"
	"github.com/nikitaSstepanov/p2p-streaming-service/backend/internal/usecases/state"
	"github.com/nikitaSstepanov/p2p-streaming-service/backend/internal/config"
	"github.com/nikitaSstepanov/p2p-streaming-service/backend/pkg/postgresql"
	"github.com/nikitaSstepanov/p2p-streaming-service/backend/pkg/server"
	"github.com/nikitaSstepanov/p2p-streaming-service/backend/migrations"
	"github.com/nikitaSstepanov/p2p-streaming-service/backend/pkg/logger"
	"github.com/nikitaSstepanov/p2p-streaming-service/backend/pkg/redis"
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

	if err := os.MkdirAll("files", 0777); err != nil {
		slog.Error("Can`t init files directory. Error:", err)
	}

	gin.SetMode(gin.ReleaseMode)

	if err := godotenv.Load(".env"); err != nil {
		slog.Error("Can`t load env. Error:", err)
	}

	if err := config.Init(); err != nil {
		slog.Error("Can't init config Error:", err)
	}

	ctx := context.TODO()
	
	db, err := postgresql.ConnectToDb(ctx, &postgresql.Config{
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

	redis, err := redis.ConnectToRedis(ctx, &redis.Config{
		Host:     viper.GetString("redis.host"),
		Port:     viper.GetString("redis.port"),
		Password: os.Getenv("REDIS_PASSWORD"),
		DBNumber: viper.GetString("redis.db"),
	})
	
	if err != nil {
		slog.Error("Can`t conntect redis. Error:", err)
	} else {
		slog.Info("Redis is connected.")
	}

	state := state.New()

	url := viper.GetString("url")

	app := &App{}

	app.Storage = storage.New(db)
	
	app.Services = services.New(app.Storage, state, redis)

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
