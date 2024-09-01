package app

import (
	"github.com/ilyakaznacheev/cleanenv"
	"github.com/joho/godotenv"
	"github.com/nikitaSstepanov/p2p-streaming-service/backend/internal/usecase/pkg/auth"
	"github.com/nikitaSstepanov/p2p-streaming-service/backend/pkg/client/postgresql"
	"github.com/nikitaSstepanov/p2p-streaming-service/backend/pkg/client/redis"
	"github.com/nikitaSstepanov/p2p-streaming-service/backend/pkg/logging"
	"github.com/nikitaSstepanov/p2p-streaming-service/backend/pkg/server"
)

type AppConfig struct {
	Logger   *logging.Config    `yaml:"logger"`
	Postgres *postgresql.Config `yaml:"postgres"`
	Redis    *redis.Config		`yaml:"redis"`
	Server   *server.Config		`yaml:"server"`
	Jwt      *auth.JwtOptions   `yaml:"jwt"`
}

func GetAppConfig(path string) (*AppConfig, error) {
	var cfg AppConfig

	if err := godotenv.Load(".env"); err != nil {
		return nil, err
	}

	err := cleanenv.ReadConfig(path, &cfg)
	if err != nil {
		return nil, err
	}

	return &cfg, nil
}
