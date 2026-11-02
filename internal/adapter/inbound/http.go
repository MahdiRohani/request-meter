package inbound

import (
	"context"
	"log"
	"net/http"

	"github.com/MahdiRohani/request-meter/configs"
	"github.com/gin-gonic/gin"
)

type HttpServer struct {
	server *http.Server
}

func NewHttpServer(counterHandler *CounterHandler) HttpServer {
	log.SetFlags(log.LstdFlags | log.Lmicroseconds)
	gin.SetMode(gin.ReleaseMode)

	router := gin.Default()
	router.POST("/counter/:name", counterHandler.IncrementEndpoint)
	router.GET("/stats/:name", counterHandler.GetStateEndpoint)

	srv := &http.Server{
		Addr:    ":" + configs.ApplicationConfigs.ServerPort,
		Handler: router,
	}

	return HttpServer{
		server: srv,
	}
}

func (h HttpServer) Serve(ctx context.Context) error {
	go func() {
		log.Printf("HTTP Server is running on port %s...", configs.ApplicationConfigs.ServerPort)
		if err := h.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("listen: %v", err)
		}
	}()
	return nil
}

func (h HttpServer) Close(ctx context.Context) error {
	log.Println("Shutting down HTTP server...")
	return h.server.Shutdown(ctx)
}
