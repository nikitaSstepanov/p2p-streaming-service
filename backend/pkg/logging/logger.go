package logging

import (
	"log/slog"
	"context"
	"os"
)

type Config struct {
	Level      Level `yaml:"level"`
	AddSource  bool  `yaml:"add_source"`
	IsJSON     bool  `yaml:"is_json"`
	SetDefault bool  `yaml:"set_default"`
}

func NewLogger(cfg *Config) *Logger {
	options := &HandlerOptions{
		AddSource: cfg.AddSource,
		Level:     cfg.Level,
	}
	
	var h Handler 

	if cfg.IsJSON {
		h = NewJSONHandler(os.Stdout, options)
	} else {
		h = NewTextHandler(os.Stdout, options)
	}

	logger := New(h)

	if cfg.SetDefault {
		SetDefault(logger)
	}

	return logger
}

func L(ctx context.Context) *Logger {
	return loggerFromContext(ctx)
}

func Default() *Logger {
	return slog.Default()
}