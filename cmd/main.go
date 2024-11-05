package main

import (
	"log/slog"
	"os"
	"snake/internal/pkg/engine"
)

func main() {
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	slog.SetDefault(logger)

	game, err := engine.New("GoSnake", 1000, 750)
	if err != nil {
		slog.Error("engine", "error", err)
		return
	}
	defer game.Cleanup()

	game.Run()
}
