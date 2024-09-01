package app

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/gin-gonic/gin"
	controller "github.com/nikitaSstepanov/p2p-streaming-service/backend/internal/controller/http/v1"
	"github.com/nikitaSstepanov/p2p-streaming-service/backend/internal/usecase"
	"github.com/nikitaSstepanov/p2p-streaming-service/backend/internal/usecase/pkg/auth"
	"github.com/nikitaSstepanov/p2p-streaming-service/backend/internal/usecase/state"
	"github.com/nikitaSstepanov/p2p-streaming-service/backend/internal/usecase/storage"
	"github.com/nikitaSstepanov/p2p-streaming-service/backend/migrations"
	"github.com/nikitaSstepanov/p2p-streaming-service/backend/pkg/client/postgresql"
	"github.com/nikitaSstepanov/p2p-streaming-service/backend/pkg/client/redis"
	"github.com/nikitaSstepanov/p2p-streaming-service/backend/pkg/logging"
	"github.com/nikitaSstepanov/p2p-streaming-service/backend/pkg/server"
)

type App struct {
	controller *controller.Controller
	usecase    *usecase.UseCase
	storage    *storage.Storage
	server     *server.Server
	logger     *logging.Logger
}

func New(configPath string) *App {
	cfg, err := GetAppConfig(configPath)
	if err != nil {
		panic("Can`t get app config. Error: " + err.Error())
	}

	logger := logging.NewLogger(cfg.Logger)

	if err := os.MkdirAll("files", 0777); err != nil {
		logger.Error("Can`t init files directory. Error:", err)
	}

	gin.SetMode(gin.ReleaseMode)

	ctx := context.TODO()
	
	postgres, err := postgresql.ConnectToDB(ctx, cfg.Postgres)
	if err != nil {
		logger.Error("Can`t connect to db. Error:", err)
	} else {
		logger.Info("Db is connected.")
	}

	if err := migrations.Migrate(postgres); err != nil {
		logger.Error("Can`t migrate db scheme. Error:", err)
	}

	redis, err := redis.ConnectToRedis(ctx, cfg.Redis)
	if err != nil {
		slog.Error("Can`t conntect redis. Error:", err)
	} else {
		slog.Info("Redis is connected.")
	}

	jwt := auth.NewJwt(cfg.Jwt)

	state := state.New()

	app := &App{}

	app.storage = storage.New(postgres, redis)
	
	app.usecase = usecase.New(app.storage, state, jwt)

	app.controller = controller.New(app.usecase)

	handler := app.controller.InitRoutes()

	app.server = server.New(handler, cfg.Server)
	
	return app
}

func (a *App) Run() error {
	if err := a.server.Start(a.logger); err != nil {
		slog.Error("Can`t run application. Error:", err)
	}

	a.logger.Info("Application is running.")

	err := ShutdownApp(a)

	if err != nil {
		a.logger.Error("Application shutdown error. Error:", err)
	}

	a.logger.Info("Application Shutting down.")

	return nil
}

func ShutdownApp(a *App) error {
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGTERM, syscall.SIGINT)
	<-quit

	if err := a.server.Shutdown(context.Background()); err != nil {
		return err
	}
	
	return nil
}
