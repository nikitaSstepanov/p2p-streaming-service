package logger

import (
	"log/slog"
	"os"
)

type Logger struct {
	Logger *slog.Logger
}

func New() *Logger {
	var err error

	err = os.RemoveAll("logs")
	if err != nil {
		panic(err)
	}

	err = os.MkdirAll("logs", 0777)
	if err != nil {
		panic(err)
	}

	file, err := os.OpenFile("logs/all.log", os.O_CREATE|os.O_RDWR, 0644)
	if err != nil {
		panic(err)
	}

	return &Logger{
		Logger: slog.New(slog.NewJSONHandler(file, nil)),
	}
}
