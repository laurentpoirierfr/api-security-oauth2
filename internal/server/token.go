package server

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

// TokenInfo structure pour les informations du token
type TokenInfo struct {
	Sub         string   `json:"sub"`
	UID         string   `json:"uid"`
	Email       string   `json:"email"`
	Name        string   `json:"name"`
	GivenName   string   `json:"given_name"`
	FamilyName  string   `json:"family_name"`
	Groups      []string `json:"groups"`
	Teams       []string `json:"teams"`
	RealmAccess struct {
		Roles []string `json:"roles"`
	} `json:"realm_access"`
	ResourceAccess map[string]struct {
		Roles []string `json:"roles"`
	} `json:"resource_access"`
	Scope             string `json:"scope"`
	Expiration        int64  `json:"exp"`
	IssuedAt          int64  `json:"iat"`
	Issuer            string `json:"iss"`
	ClientID          string `json:"client_id"`
	TokenType         string `json:"token_type"`
	PreferredUsername string `json:"preferred_username"`
}

// extractTokenInfo extrait et parse les informations du token via l'endpoint userinfo
func (s *proxyServer) extractTokenInfo(ctx context.Context, tokenString string) (*TokenInfo, error) {
	// Créer la requête vers l'endpoint userinfo
	req, err := http.NewRequestWithContext(ctx, "GET", s.cfg.Server.OAuth2.Endpoints.UserInfoURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Ajouter le token Bearer dans les headers
	req.Header.Set("Authorization", "Bearer "+tokenString)
	req.Header.Set("Content-Type", "application/json")

	// Créer un client HTTP avec timeout
	client := &http.Client{
		Timeout: 10 * time.Second,
	}

	// Exécuter la requête
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to get user info: %w", err)
	}
	defer resp.Body.Close()

	// Vérifier le statut de la réponse
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("userinfo endpoint returned status: %d", resp.StatusCode)
	}

	// Décoder la réponse JSON
	var tokenInfo TokenInfo
	if err := json.NewDecoder(resp.Body).Decode(&tokenInfo); err != nil {
		return nil, fmt.Errorf("failed to decode user info: %w", err)
	}

	// Validation des champs obligatoires
	if tokenInfo.Sub == "" {
		return nil, fmt.Errorf("missing required field 'sub' in token info")
	}

	// Initialiser les champs qui peuvent être vides
	if tokenInfo.TokenType == "" {
		tokenInfo.TokenType = "Bearer"
	}

	// Initialiser les structures imbriquées si elles sont nil
	if tokenInfo.RealmAccess.Roles == nil {
		tokenInfo.RealmAccess.Roles = []string{}
	}
	if tokenInfo.ResourceAccess == nil {
		tokenInfo.ResourceAccess = make(map[string]struct {
			Roles []string `json:"roles"`
		})
	}
	if tokenInfo.Groups == nil {
		tokenInfo.Groups = []string{}
	}
	if tokenInfo.Teams == nil {
		tokenInfo.Teams = []string{}
	}

	return &tokenInfo, nil
}

// setTokenHeaders ajoute les informations du token dans les headers de sortie
func (s *proxyServer) setTokenHeaders(c *gin.Context, tokenInfo *TokenInfo, tokenString string) {
	// Headers standards pour l'identification
	c.Header("X-User-ID", tokenInfo.UID)
	c.Header("X-User-Email", tokenInfo.Email)
	c.Header("X-User-Name", tokenInfo.Name)
	c.Header("X-Token-Subject", tokenInfo.Sub)

	// Headers pour les teams et groupes
	if len(tokenInfo.Groups) > 0 {
		c.Header("X-User-Groups", strings.Join(tokenInfo.Groups, ","))
	}
	if len(tokenInfo.Teams) > 0 {
		c.Header("X-User-Teams", strings.Join(tokenInfo.Teams, ","))
	}

	// Headers pour les rôles
	if len(tokenInfo.RealmAccess.Roles) > 0 {
		c.Header("X-User-Realm-Roles", strings.Join(tokenInfo.RealmAccess.Roles, ","))
	}

	// Headers pour les scopes
	if tokenInfo.Scope != "" {
		c.Header("X-Token-Scopes", tokenInfo.Scope)
	}

	// Headers techniques
	c.Header("X-Token-Type", tokenInfo.TokenType)
	c.Header("X-Token-Issuer", tokenInfo.Issuer)
	c.Header("X-Client-ID", tokenInfo.ClientID)

	// Stocker les informations dans le contexte pour un usage ultérieur
	c.Set("tokenInfo", tokenInfo)
	c.Set("accessToken", tokenString)
	c.Set("userID", tokenInfo.UID)
	c.Set("userEmail", tokenInfo.Email)
}

// Middleware qui extrait et valide le token
func (s *proxyServer) tokenExtractionMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		token, err := getTokenFromHeader(c)
		if err != nil {
			// Token non trouvé, mais on continue pour les routes publiques
			c.Set("hasToken", false)
			c.Next()
			return
		}

		// Token trouvé, stocker dans le contexte
		c.Set("hasToken", true)
		c.Set("accessToken", token)

		tokenInfo, err := s.extractTokenInfo(c.Request.Context(), token)
		if err != nil {
			c.JSON(401, gin.H{
				"error":   "Invalid token",
				"message": "Token invalide ou expiré",
			})
			c.Abort()
			return
		}
		s.setTokenHeaders(c, tokenInfo, token)

		c.Next()
	}
}

// getTokenFromHeader extrait le token Bearer du header Authorization
func getTokenFromHeader(c *gin.Context) (string, error) {
	authHeader := c.GetHeader("Authorization")
	if authHeader == "" {
		return "", errors.New("authorization header manquant")
	}

	// Vérifier le format Bearer
	if !strings.HasPrefix(authHeader, "Bearer ") {
		return "", errors.New("format d'autorisation invalide, attendu: Bearer <token>")
	}

	// Extraire le token (enlever "Bearer " prefix)
	token := strings.TrimSpace(authHeader[7:])
	if token == "" {
		return "", errors.New("token manquant après Bearer")
	}

	return token, nil
}
