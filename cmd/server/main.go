package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/MahdiRohani/request-meter/internal/adapters/inbound/http"
	"github.com/MahdiRohani/request-meter/internal/adapters/outbound"
	"github.com/MahdiRohani/request-meter/internal/core/services"
	"github.com/gin-gonic/gin"
)

func main() {

	slowRepo := outbound.NewSlowDatabaseRepository()
	fastRepo := outbound.NewInMemoryRepository()
	lock := outbound.NewRedisRepository()

	savedCounter := slowRepo.GetAll()
	for _, counter := range savedCounter {
		fastRepo.Save(counter)
	}

	counterService := services.NewCounterService(fastRepo, slowRepo, lock)

	cnx, cancel := context.WithCancel(context.Background())

	counterHandler := http.NewCounterHandler(counterService)

	gin.SetMode(gin.ReleaseMode)
	router := gin.Default()

	api := router.Group("/api/v1")
	{
		api.POST("/counter/:name", counterHandler.IncrementEndpoint)
		api.GET("/counter/:name", counterHandler.GetStateEndpoint)
	}

	go func() {
		if err := router.Run(":8080"); err != nil {
			log.Fatalf("Failed to start server: %v", err)
		}
	}()

	counterService.StartSync(cnx, 5*time.Second)
	stopChan := make(chan os.Signal, 1)
	signal.Notify(stopChan, os.Interrupt, syscall.SIGTERM)
	<-stopChan

	cancel()
	time.Sleep(2 * time.Second)
}
