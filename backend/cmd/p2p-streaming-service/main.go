package main

import (
	"github.com/nikitaSstepanov/p2p-streaming-service/backend/internal/app"
)

func main() {
	app := app.New()

	app.Run()
}