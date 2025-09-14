package server

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/laurentpoirierfr/api-security-oauth2/internal/config"
	ginoauth2 "github.com/zalando/gin-oauth2"
	"github.com/zalando/gin-oauth2/zalando"
	"golang.org/x/oauth2"
)

// Modification de addOAuth2Middleware pour gérer les routes publiques/privées
func (s *proxyServer) addOAuth2Middleware() {
	s.engine.Use(ginoauth2.RequestLogger([]string{"uid"}, "data"))

	for _, route := range s.cfg.Routes {
		routeGroup := s.engine.Group(route.Path)

		// Appliquer OAuth2 seulement si des teams sont configurées
		if len(route.Teams) > 0 {
			log.Printf("Protecting route %s with teams: %+v", route.Path, route.Teams)
			routeGroup.Use(s.oauth2Middleware(route.Teams))
		} else {
			log.Printf("Public route: %s", route.Path)
			routeGroup.Use(s.publicMiddleware())
		}

		routeGroup.Any("", s.proxy)
	}
}

// oauth2Middleware crée un middleware Gin pour l'authentification OAuth2
func (s *proxyServer) oauth2Middleware(teams []config.Team) gin.HandlerFunc {
	// Convertir les teams en AccessTuple pour zalando
	var accessTuples []zalando.AccessTuple
	for _, team := range teams {
		accessTuples = append(accessTuples, zalando.AccessTuple{
			Realm: "teams", // ou le realm approprié
			Uid:   team.Name,
			Cn:    team.Description,
		})
	}

	return func(c *gin.Context) {
		log.Printf("OAuth2 Middleware - Vérification du token")
		log.Printf("Teams requises: %+v", teams)

		// Récupérer le token du header Authorization
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			log.Printf("Authorization header manquant")
			c.JSON(http.StatusUnauthorized, gin.H{
				"error":   "Unauthorized",
				"message": "Token d'accès manquant",
			})
			c.Abort()
			return
		}

		// Vérifier le format du header
		if !strings.HasPrefix(authHeader, "Bearer ") {
			log.Printf("Format de token invalide")
			c.JSON(http.StatusUnauthorized, gin.H{
				"error":   "Unauthorized",
				"message": "Format de token invalide. Utilisez 'Bearer <token>'",
			})
			c.Abort()
			return
		}

		// Extraire le token
		tokenString := strings.TrimPrefix(authHeader, "Bearer ")

		// Vérifier le token avec le endpoint tokeninfo
		tokenInfo, err := s.verifyToken(c.Request.Context(), tokenString)
		if err != nil {
			log.Printf("Erreur de vérification du token: %v", err)
			c.JSON(http.StatusUnauthorized, gin.H{
				"error":   "Unauthorized",
				"message": "Token invalide ou expiré",
			})
			c.Abort()
			return
		}

		// Vérifier les scopes si nécessaire
		if !s.hasRequiredScopes(tokenInfo.Scope) {
			log.Printf("Scopes insuffisants: %s", tokenInfo.Scope)
			c.JSON(http.StatusForbidden, gin.H{
				"error":   "Forbidden",
				"message": "Permissions insuffisantes",
			})
			c.Abort()
			return
		}

		// Vérifier l'appartenance aux teams
		if len(teams) > 0 {
			if !s.hasRequiredTeams(tokenInfo, teams) {
				log.Printf("Teams insuffisantes. Token teams: %+v, Required: %+v",
					tokenInfo.Teams, teams)
				c.JSON(http.StatusForbidden, gin.H{
					"error":   "Forbidden",
					"message": "Accès non autorisé aux ressources",
				})
				c.Abort()
				return
			}
		}

		// Stocker les informations du token dans le contexte
		c.Set("tokenInfo", tokenInfo)
		c.Set("accessToken", tokenString)

		log.Printf("Token validé avec succès pour l'utilisateur: %s", tokenInfo.UID)
		c.Next()
	}
}

// verifyToken vérifie la validité du token auprès du serveur OAuth2
func (s *proxyServer) verifyToken(ctx context.Context, tokenString string) (*TokenInfo, error) {
	// Utiliser le endpoint tokeninfo pour valider le token
	// Vous pouvez utiliser une librairie comme golang.org/x/oauth2
	// ou faire une requête HTTP directe vers le endpoint tokeninfo

	token := &oauth2.Token{AccessToken: tokenString}

	// Vérifier avec le client OAuth2
	httpClient := oauth2.NewClient(ctx, oauth2.StaticTokenSource(token))

	// Faire une requête vers le endpoint userinfo ou tokeninfo
	// pour récupérer les informations du token
	userInfoURL := strings.Replace(s.cfg.Server.OAuth2.Endpoints.TokenURL,
		"token", "userinfo", 1)

	resp, err := httpClient.Get(userInfoURL)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("token validation failed: %s", resp.Status)
	}

	// Parser la réponse et retourner les informations du token
	var tokenInfo TokenInfo
	if err := json.NewDecoder(resp.Body).Decode(&tokenInfo); err != nil {
		return nil, err
	}

	return &tokenInfo, nil
}

// hasRequiredScopes vérifie si le token a les scopes requis
func (s *proxyServer) hasRequiredScopes(tokenScopes string) bool {
	// Implémentation basique - à adapter selon vos besoins
	requiredScopes := []string{"openid", "profile"}
	scopes := strings.Split(tokenScopes, " ")

	for _, required := range requiredScopes {
		found := false
		for _, scope := range scopes {
			if scope == required {
				found = true
				break
			}
		}
		if !found {
			return false
		}
	}
	return true
}

// hasRequiredTeams vérifie si l'utilisateur appartient aux teams requises
func (s *proxyServer) hasRequiredTeams(tokenInfo *TokenInfo, requiredTeams []config.Team) bool {
	// Cette implémentation dépend de la structure de votre token
	// Adaptez-la selon comment les teams sont stockées dans le token

	// Exemple: si les teams sont dans tokenInfo.Groups
	tokenTeams := tokenInfo.Groups // ou tokenInfo.Teams selon votre structure

	for _, requiredTeam := range requiredTeams {
		found := false
		for _, tokenTeam := range tokenTeams {
			if tokenTeam == requiredTeam.Name {
				found = true
				break
			}
		}
		if !found {
			return false
		}
	}
	return true
}

// Structure pour les informations du token
type TokenInfo struct {
	Sub        string   `json:"sub"`
	UID        string   `json:"uid"`
	Email      string   `json:"email"`
	Name       string   `json:"name"`
	Groups     []string `json:"groups"`
	Teams      []string `json:"teams"`
	Scope      string   `json:"scope"`
	Expiration int64    `json:"exp"`
}

// Alternative: Utilisation de la librairie zalando/gin-oauth2
func (s *proxyServer) oauth2MiddlewareZalando(teams []config.Team) gin.HandlerFunc {
	// Convertir les teams en AccessTuple
	var accessTuples []zalando.AccessTuple
	for _, team := range teams {
		accessTuples = append(accessTuples, zalando.AccessTuple{
			Realm: "teams",
			Uid:   team.Name,
			Cn:    team.Description,
		})
	}

	// Utiliser le middleware zalando
	return ginoauth2.Auth(zalando.GroupCheck(accessTuples), oauth2.Endpoint{
		AuthURL:  s.cfg.Server.OAuth2.Endpoints.AuthURL,
		TokenURL: s.cfg.Server.OAuth2.Endpoints.TokenURL,
	})
}

// Middleware simplifié pour les routes publiques
func (s *proxyServer) publicMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Pour les routes publiques, on peut juste logger l'accès
		log.Printf("Accès public à: %s", c.Request.URL.Path)
		c.Next()
	}
}
