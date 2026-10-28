package app

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	_ "github.com/lib/pq"

	"github.com/MahdiRohani/request-meter/internal/adapters/inbound"
	"github.com/MahdiRohani/request-meter/internal/adapters/outbound"
	"github.com/MahdiRohani/request-meter/internal/core/services"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"github.com/redis/go-redis/v9"
)

func Run() {
	if err := godotenv.Load(); err != nil {
		log.Println("Warning: No .env file found. Using system environment variables.")
	}

	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		log.Fatal("DATABASE_URL is required in .env")
	}

	redisAddr := os.Getenv("REDIS_ADDR")
	if redisAddr == "" {
		log.Fatal("REDIS_ADDR is required in .env")
	}

	serverPort := os.Getenv("SERVER_PORT")
	if serverPort == "" {
		serverPort = "8080"
	}

	syncIntervalStr := os.Getenv("SYNC_INTERVAL")
	syncInterval, err := time.ParseDuration(syncIntervalStr)
	if err != nil || syncInterval <= 0 {
		syncInterval = 500 * time.Millisecond
	}

	postgres := outbound.NewPostgresRepository(dbURL)
	rdb := redis.NewClient(&redis.Options{
		Addr: redisAddr,
	})
	redisLock := outbound.NewRedisLockAdapter(rdb)

	memoryRepo := outbound.NewInMemoryRepository()

	counterService := services.NewCounterService(memoryRepo, postgres, redisLock)

	ctx, cancelWorker := context.WithCancel(context.Background())

	counterService.StartSync(ctx, syncInterval)

	counterHandler := inbound.NewCounterHandler(counterService)

	log.SetFlags(log.LstdFlags | log.Lmicroseconds)
	gin.SetMode(gin.ReleaseMode)
	router := gin.Default()
	router.POST("/counter/:name", counterHandler.IncrementEndpoint)
	router.GET("/stats/:name", counterHandler.GetStateEndpoint)

	srv := &http.Server{
		Addr:    ":" + serverPort,
		Handler: router,
	}

	go func() {
		log.Printf("Server is running on port %s...", serverPort)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("listen: %s\n", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Shutting down server gracefully...")

	cancelWorker()
	time.Sleep(2 * time.Second)

	ctxShutdown, cancelShutdown := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancelShutdown()

	if err := srv.Shutdown(ctxShutdown); err != nil {
		log.Fatal("Server forced to shutdown:", err)
	}

	log.Println("Server exiting successfully. Data saved.")
}
