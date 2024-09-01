package redis

import (
	"context"
	"strconv"
	"fmt"

	"github.com/redis/go-redis/v9"
)

type Config struct {
	Host     string `yaml:"host"`
	Port     string `yaml:"port"`
	Password string `env:"REDIS_PASSWORD"`
	DBNumber string `yaml:"db"`
}

func GetConfig(cfg *Config) (*redis.Options, error) {
	db, err := strconv.Atoi(cfg.DBNumber)
	if err != nil {
		return nil, err
	}

	return &redis.Options{
		Addr:     fmt.Sprintf("%s:%s", cfg.Host, cfg.Port),
		Password: cfg.Password,
		DB:       db,
	}, nil
}

func ConnectToRedis(ctx context.Context, cfg *Config) (*redis.Client, error) {
	config, err := GetConfig(cfg)
	if err != nil {
		return nil, err
	}

	client := redis.NewClient(config)

	if err := client.Ping(ctx).Err(); err != nil {
		return nil, err
	}

	return client, nil
}