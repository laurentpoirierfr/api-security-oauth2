package server

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/laurentpoirierfr/api-security-oauth2/internal/helper"
	"golang.org/x/time/rate"
)

// RateLimiterMiddleware limits the number of requests to prevent DDoS attacks
// RateLimiterMiddleware returns a Gin middleware that limits the number of incoming requests
// to prevent DDoS attacks. The rate and burst limits can be configured via the environment
// variables "RATE_LIMIT" and "BURST_LIMIT". If these variables are not set, default values
// of 1 request per second and a burst of 4 are used. When the request rate exceeds the limit,
// the middleware responds with HTTP 429 (Too Many Requests) and a JSON error message.
func RateLimiterMiddleware() gin.HandlerFunc {
	rateLimit := 1
	burstLimit := 4

	if rlEnv := helper.GetEnvInt("RATE_LIMIT", rateLimit); rlEnv > 0 {
		rateLimit = rlEnv
	}
	if blEnv := helper.GetEnvInt("BURST_LIMIT", burstLimit); blEnv > 0 {
		burstLimit = blEnv
	}

	limiter := rate.NewLimiter(rate.Limit(rateLimit), burstLimit)
	return func(c *gin.Context) {
		if limiter.Allow() {
			c.Next()
		} else {
			c.JSON(http.StatusTooManyRequests, gin.H{
				"message": "Limite exceed",
			})
		}

	}
}
