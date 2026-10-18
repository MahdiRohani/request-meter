package server

import (
	"context"
	"time"

	"github.com/MahdiRohani/request-meter/internal/adapters/outbound"
	"github.com/MahdiRohani/request-meter/internal/core/services"
)

func main() {

	slowRepo := outbound.NewSlowDatabaseRepository()
	fastRepo := outbound.NewInMemoryRepository()

	savedCounter := slowRepo.GetAll()
	for _, counter := range savedCounter {
		fastRepo.Save(counter)
	}

	counterService := services.NewCounterService(fastRepo, slowRepo)

	cnx, cancel := context.WithCancel(context.Background())

	counterService.StartSync(cnx, 5*time.Second)

	cancel()
}
