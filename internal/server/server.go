package server

import (
	"github.com/gin-gonic/gin"
	"github.com/laurentpoirierfr/api-security-oauth2/internal/config"

	"github.com/charmbracelet/log"
)

type Server interface {
	Start() error
}

func NewServer(cfg *config.Config) Server {
	return &proxyServer{
		cfg: cfg,
	}
}

type proxyServer struct {
	engine *gin.Engine
	cfg    *config.Config
}

func (s *proxyServer) Start() error {
	// Initialize Gin engine
	s.engine = gin.Default()

	// Add operational routes (health, metrics, etc.)
	s.addOpsRoutes()

	// Add your routes and handlers here
	// s.engine.Any("/api/*proxyPath", s.proxy)

	// Add middleware
	s.addMiddlewares()

	log.Info("Starting server on port " + s.cfg.Server.Port)
	return s.engine.Run(":" + s.cfg.Server.Port)
}
