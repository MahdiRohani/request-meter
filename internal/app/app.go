package app

import (
	"context"
	"log"
	"sync"

	"github.com/MahdiRohani/request-meter/configs"
	"github.com/MahdiRohani/request-meter/internal/adapter/inbound"
	"github.com/MahdiRohani/request-meter/internal/adapter/outbound"
	services "github.com/MahdiRohani/request-meter/internal/service"
)

type App struct {
	start func() error
	stop  func() error
}

func NewApp(ctx context.Context) App {

	postgres := outbound.NewPostgresRepository(configs.ApplicationConfigs.DatabaseURL)

	redisLock := outbound.NewRedisLockAdapter(configs.ApplicationConfigs.RedisAddr)

	memoryRepo := outbound.NewInMemoryRepository()

	counterService := services.NewCounterService(memoryRepo, postgres, redisLock)

	counterHandler := inbound.NewCounterHandler(counterService)

	httpServer := inbound.NewHttpServer(counterHandler)

	workerCtx, cancelWorker := context.WithCancel(context.Background())
	var wg sync.WaitGroup

	return App{
		start: func() error {
			counterService.StartSync(workerCtx, &wg, configs.ApplicationConfigs.SyncInterval)
			return httpServer.Serve(ctx)
		},
		stop: func() error {
			cancelWorker()
			wg.Wait()
			if err := httpServer.Close(ctx); err != nil {
				log.Println("Error closing HTTP server:", err)
			}
			if err := postgres.Close(ctx); err != nil {
				log.Println("Error closing Postgres:", err)
			}
			if err := redisLock.Close(ctx); err != nil {
				log.Println("Error closing Redis:", err)
			}
			return nil
		},
	}
}

func (a App) Start(ctx context.Context) error {
	return a.start()
}

func (a App) Stop(ctx context.Context) error {
	return a.stop()
}
