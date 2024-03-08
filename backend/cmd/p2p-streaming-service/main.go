package main

import (
	"log/slog"

	"github.com/nikitaSstepanov/p2p-streaming-service/backend/internal/app"
)

func main() {
	app := app.New()

	if err := app.Run(); err != nil {
		slog.Error("Can`t run application. Error:", err)
	}
}
