package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/MahdiRohani/request-meter/internal/app"
)

func main() {
	ctx := context.Background()

	meterApp := app.NewApp(ctx)
	err := meterApp.Start(ctx)
	if err != nil {
		log.Fatalf("failed to start the application: %v", err)
	}

	ListenForGracefulShutdown(meterApp)
}

func ListenForGracefulShutdown(application app.App) {
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	log.Println("Application started. Press Ctrl+C to shut down.")
	<-sigChan

	done := make(chan bool)

	log.Println("Shutting down...")
	go func() {
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
		defer cancel()

		err := application.Stop(shutdownCtx)
		if err != nil {
			log.Println(err)
		}
		close(done)
	}()

	select {
	case <-done:
		log.Println("Shutdown complete")
	case <-time.After(20 * time.Second):
		log.Println("Shutdown timeout. Forced exit.")
	}
}
