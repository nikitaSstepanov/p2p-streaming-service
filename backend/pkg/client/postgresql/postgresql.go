package postgresql

import (
	"context"
	"time"
	"fmt"

	"github.com/nikitaSstepanov/p2p-streaming-service/backend/pkg/utils/repeatable"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5"
)

type Client interface {
	Exec(ctx context.Context, sql string, arguments ...any) (pgconn.CommandTag, error)
	QueryRow(ctx context.Context, sql string, args ...any) pgx.Row
	Begin(ctx context.Context) (pgx.Tx, error)
	Query(ctx context.Context, sql string, args ...any) (pgx.Rows, error)
	Ping(ctx context.Context) error
	Close()
}

type Config struct {
	Host     string `yaml:"host"`
	Port     string `yaml:"port"`
	Username string `yaml:"username"`
	Password string `env:"POSTGRES_PASSWORD"`
	DBName   string `yaml:"dbname"`
	SSLMode  string `yaml:"sslmode"`
}

const (
	maxAttempts = 5
)

func GetConnConfig(cfg *Config) (*pgxpool.Config, error) {
	config, err := pgxpool.ParseConfig(fmt.Sprintf(
		"user=%s password=%s host=%s port=%s dbname=%s sslmode=%s",
		cfg.Username, cfg.Password, cfg.Host, cfg.Port, cfg.DBName, cfg.SSLMode))
	if err != nil {
		return nil, err
	}

	return config, nil
}

func ConnectToDB(ctx context.Context, cfg *Config) (*pgxpool.Pool, error) {
	var db *pgxpool.Pool

	var err error

	connCfg, err := GetConnConfig(cfg)
	if err != nil {
		return nil, err
	}

	err = repeatable.DoWithTries(
		func() error {
			ctx, cancel := context.WithTimeout(ctx, 5 * time.Second)
			defer cancel()

			db, err = pgxpool.NewWithConfig(ctx, connCfg)
			if err != nil {
				return err
			}

			if err := db.Ping(ctx); err != nil {
				return err
			}
			
			return nil
		}, maxAttempts, 5 * time.Second)

	if err != nil {
		return nil, err
	}

	if err := db.Ping(ctx); err != nil {
		return nil, err
	}

	return db, nil
}