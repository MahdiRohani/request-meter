package app

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/MahdiRohani/request-meter/configs"
	"github.com/MahdiRohani/request-meter/internal/adapter/inbound"
	"github.com/MahdiRohani/request-meter/internal/adapter/outbound"
	services "github.com/MahdiRohani/request-meter/internal/service"
	"github.com/gin-gonic/gin"
)

func Run() {

	postgres := outbound.NewPostgresRepository(configs.ApplicationConfigs.DatabaseURL)

	redisLock := outbound.NewRedisLockAdapter(configs.ApplicationConfigs.RedisAddr)

	memoryRepo := outbound.NewInMemoryRepository()

	counterService := services.NewCounterService(memoryRepo, postgres, redisLock)

	ctx, cancelWorker := context.WithCancel(context.Background())
	var wg sync.WaitGroup

	counterService.StartSync(ctx, &wg, configs.ApplicationConfigs.SyncInterval)

	counterHandler := inbound.NewCounterHandler(counterService)

	log.SetFlags(log.LstdFlags | log.Lmicroseconds)
	gin.SetMode(gin.ReleaseMode)

	router := gin.Default()
	router.POST("/counter/:name", counterHandler.IncrementEndpoint)
	router.GET("/stats/:name", counterHandler.GetStateEndpoint)

	srv := &http.Server{
		Addr:    ":" + configs.ApplicationConfigs.ServerPort,
		Handler: router,
	}

	go func() {
		log.Printf("Server is running on port %s...", configs.ApplicationConfigs.ServerPort)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("listen: %s\n", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Shutting down server gracefully...")

	cancelWorker()

	wg.Wait()
	ctxShutdown, cancelShutdown := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancelShutdown()

	if err := srv.Shutdown(ctxShutdown); err != nil {
		log.Fatal("Server forced to shutdown:", err)
	}

	log.Println("Server exiting successfully. Data saved.")
}
