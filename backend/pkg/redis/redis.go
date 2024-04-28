package redis

import (
	"context"
	"strconv"
	"fmt"

	"github.com/redis/go-redis/v9"
)

type Config struct {
	Host      string
	Port      string
	Password  string
	DBNumber  string
}

func GetConfig(cfg *Config) (*redis.Options, error) {
	db, err := strconv.Atoi(cfg.DBNumber)

	if err != nil {
		return nil, err
	}

	return &redis.Options{
		Addr: fmt.Sprintf("%s:%s", cfg.Host, cfg.Port),
		Password: cfg.Password,
		DB: db,
	}, nil
}

func ConnectToRedis(ctx context.Context, cfg *Config) (*redis.Client, error) {
	config, err := GetConfig(cfg)

	if err != nil {
		return nil, err
	}

	client := redis.NewClient(config)

	status, err := client.Ping(ctx).Result()

	if err != nil && status != "PONG" {
		return nil, err
	}
	
	return client, nil
}