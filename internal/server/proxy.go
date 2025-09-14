package server

import (
	"io"
	"maps"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

// proxy gère le forwarding des requêtes
func (s *proxyServer) proxy(c *gin.Context) {

	targetURL := s.getTargetURL(c)

	req, err := http.NewRequest(c.Request.Method, targetURL, c.Request.Body)
	if err != nil {
		c.JSON(500, gin.H{"error": "Failed to create request"})
		return
	}

	// Copy headers
	maps.Copy(req.Header, c.Request.Header)

	// Propagate token-related headers to backend
	s.propagateTokenHeadersToBackend(c, req)

	client := &http.Client{}
	client.Timeout = time.Duration(s.cfg.Server.TimeOut) * time.Second

	resp, err := client.Do(req)
	if err != nil {
		c.JSON(502, gin.H{"error": "Failed to reach backend"})
		return
	}
	defer resp.Body.Close()

	// Copy response headers
	for k, v := range resp.Header {
		c.Header(k, v[0])
		c.Status(resp.StatusCode)
		_, err = io.Copy(c.Writer, resp.Body)
		if err != nil {
			c.JSON(500, gin.H{"error": "Failed to copy response"})
		}
	}
}

// getTargetURL construit l'URL cible pour le proxy
func (s *proxyServer) getTargetURL(c *gin.Context) string {
	// Récupérer le chemin de la requête
	requestPath := c.Request.URL.Path

	// Trouver la configuration de route correspondante
	targetBase := s.cfg.Server.DefaultTarget
	for _, route := range s.cfg.Routes {
		if strings.HasPrefix(requestPath, route.Path) {
			targetBase = route.Target // Ajouter Target à votre config
			break
		}
	}

	return targetBase
}

// propagateTokenHeadersToBackend propage les headers vers la requête backend
func (s *proxyServer) propagateTokenHeadersToBackend(c *gin.Context, backendReq *http.Request) {
	// Liste des headers à propager vers le backend
	tokenHeaders := []string{
		"X-User-ID", "X-User-Sub", "X-User-Email", "X-User-Name",
		"X-User-Given-Name", "X-User-Family-Name", "X-User-Preferred-Username",
		"X-User-Groups", "X-User-Teams", "X-User-Realm-Roles",
		"X-Token-Scopes", "X-Token-Type", "X-Token-Issuer", "X-Client-ID",
	}

	for _, header := range tokenHeaders {
		if value := c.Writer.Header().Get(header); value != "" {
			backendReq.Header.Set(header, value)
		}
	}

	// Propager les headers de resource access
	for key, values := range c.Writer.Header() {
		if strings.HasPrefix(key, "X-Resource-") && strings.HasSuffix(key, "-Roles") {
			for _, value := range values {
				backendReq.Header.Add(key, value)
			}
		}
	}
}
