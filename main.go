package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	_ "github.com/kardianos/minwinsvc"
	log "golang.org/x/exp/slog"
)

func main() {
	// Perform the startup and shutdown sequence.
	// create a channel to listen for an interrupt or terminate signal from the OS.
	shutdown := make(chan os.Signal, 1)
	signal.Notify(shutdown, os.Interrupt, syscall.SIGTERM, syscall.SIGINT)

	if err := run(context.Background(), shutdown); err != nil {
		log.Error("startup", "error", err)
		os.Exit(1)
	}
}
