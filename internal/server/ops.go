package server

import (
	"time"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

type Health struct {
	Status string `json:"status"`
	Time   string `json:"time"`
}

func NewHealth(status string) Health {
	return Health{
		Status: status,
		Time:   time.Now().Format(time.RFC3339),
	}
}

func (s *proxyServer) addOpsRoutes() {
	// Operational endpoints
	// Health, Metrics, Info
	// /ops/readiness
	s.engine.GET("/ops/readiness", func(c *gin.Context) {
		health := NewHealth("ready")
		c.JSON(200, health)
	})

	// /ops/liveness
	s.engine.GET("/ops/liveness", func(c *gin.Context) {
		health := NewHealth("alive")
		c.JSON(200, health)
	})

	// /ops/metrics
	s.engine.GET("/ops/metrics", gin.WrapH(promhttp.Handler()))

	// /ops/info
	s.engine.GET("/ops/info", func(c *gin.Context) {
		c.JSON(200, s.cfg.Application)
	})
}
