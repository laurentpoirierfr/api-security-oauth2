package server

import (
	"io"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

func (s *proxyServer) proxy(c *gin.Context) {
	backend_url := s.cfg.Server.BackendURL
	if backend_url == "" {
		c.JSON(500, gin.H{"error": "Backend URL not configured"})
		return
	}

	targetURL := backend_url + c.Request.URL.Path

	req, err := http.NewRequest(c.Request.Method, targetURL, c.Request.Body)
	if err != nil {
		c.JSON(500, gin.H{"error": "Failed to create request"})
		return
	}

	// Copy headers
	for k, v := range c.Request.Header {
		req.Header[k] = v
	}

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
