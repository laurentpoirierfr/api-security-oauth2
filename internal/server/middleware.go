package server

import (
	xss "github.com/dvwright/xss-mw"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"

	"github.com/danielkov/gin-helmet/ginhelmet"
)

func (s *proxyServer) addMiddlewares() {
	// Add your middleware here
	// https://github.com/gin-gonic/contrib

	// Custom error handling middleware
	s.engine.Use(ErrorMiddleware())

	// Logging middleware
	s.engine.Use(gin.Logger())

	// XSS protection middleware
	var xssMdlwr xss.XssMw
	s.engine.Use(xssMdlwr.RemoveXss())

	// CORS middleware
	s.engine.Use(cors.Default()) // All origins allowed by default

	// Use default security headers
	s.engine.Use(ginhelmet.Default())

	// Oauth2 Middleware
	s.addOAuth2Middleware()

	// Recovery middleware to recover from panics
	s.engine.Use(gin.Recovery())

}
