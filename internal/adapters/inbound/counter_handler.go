package inbound

import (
	"net/http"

	"github.com/MahdiRohani/request-meter/internal/core/ports/inbound"
	"github.com/gin-gonic/gin"
)

type CounterHandler struct {
	service inbound.CounterService
}

func NewCounterHandler(service inbound.CounterService) *CounterHandler {
	return &CounterHandler{
		service: service,
	}
}

func (h *CounterHandler) IncrementEndpoint(c *gin.Context) {
	name := c.Param("name")

	if name == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "counter name is required"})
		return
	}

	err := h.service.Increment(name)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "counter incremented successfully", "counter_name": name})
}

func (h *CounterHandler) GetStateEndpoint(c *gin.Context) {
	name := c.Param("name")

	if name == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "counter name is required"})
		return
	}
	stats, err := h.service.GetStats(name)

	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, stats)
}
