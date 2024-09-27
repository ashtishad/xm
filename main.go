package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	_ "github.com/ashtishad/xm/docs"
	"github.com/ashtishad/xm/internal/server"
)

// @title XM API
// @host localhost:8080
// @BasePath /api
func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	srv, err := server.NewServer(ctx)
	if err != nil {
		log.Fatalf("failed to create server: %v", err)
	}

	go func() {
		if err := srv.Start(ctx); err != nil {
			srv.Logger.Error("server error", "err", err)
			cancel()
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	<-quit

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer shutdownCancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		srv.Logger.Error("server forced to shutdown", "err", err)
		os.Exit(1)
	}
}
