package server

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// ErrorResponse structure standardisée pour les erreurs
type ErrorResponse struct {
	Success bool   `json:"success"`
	Error   string `json:"error"`
	Message string `json:"message,omitempty"`
	Code    int    `json:"code"`
}

// ErrorDetailResponse pour les erreurs avec détails supplémentaires
type ErrorDetailResponse struct {
	ErrorResponse
	Details any `json:"details,omitempty"`
}

// ErrorMiddleware crée un middleware qui capture et formate les erreurs
func ErrorMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Exécuter les handlers suivants
		c.Next()

		// Vérifier s'il y a des erreurs
		if len(c.Errors) > 0 {
			// Récupérer la dernière erreur
			lastError := c.Errors.Last()

			// Déterminer le code HTTP
			statusCode := c.Writer.Status()
			if statusCode == http.StatusOK {
				// Si le status est 200 mais qu'il y a une erreur, utiliser 500 par défaut
				statusCode = http.StatusInternalServerError
			}

			// Formater la réponse d'erreur
			response := ErrorResponse{
				Success: false,
				Error:   http.StatusText(statusCode),
				Message: lastError.Error(),
				Code:    statusCode,
			}

			// Si l'erreur implémente une interface avec des détails, les inclure
			if detailedError, ok := lastError.Err.(interface{ Details() any }); ok {
				detailedResponse := ErrorDetailResponse{
					ErrorResponse: response,
					Details:       detailedError.Details(),
				}
				c.JSON(statusCode, detailedResponse)
			} else {
				c.JSON(statusCode, response)
			}

			// Nettoyer les erreurs pour éviter les doublons
			c.Errors = c.Errors[:0]
		}
	}
}
