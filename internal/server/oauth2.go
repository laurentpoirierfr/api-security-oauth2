package server

import (
	"log"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/laurentpoirierfr/api-security-oauth2/internal/config"
	ginoauth2 "github.com/zalando/gin-oauth2"
)

// Modification de addOAuth2Middleware pour gérer les routes publiques/privées
func (s *proxyServer) addOAuth2Middleware() {
	s.engine.Use(ginoauth2.RequestLogger([]string{"uid"}, "data"))
	// Middleware d'extraction de token pour toutes les routes
	s.engine.Use(s.tokenExtractionMiddleware())

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

// oauth2Middleware avec extraction et propagation des informations du token
func (s *proxyServer) oauth2Middleware(teams []config.Team) gin.HandlerFunc {
	return func(c *gin.Context) {
		log.Printf("OAuth2 Middleware - Vérification du token")

		// Récupérer le token du header Authorization
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error":   "Unauthorized",
				"message": "Token d'accès manquant",
			})
			c.Abort()
			return
		}

		if !strings.HasPrefix(authHeader, "Bearer ") {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error":   "Unauthorized",
				"message": "Format de token invalide",
			})
			c.Abort()
			return
		}

		tokenString := strings.TrimPrefix(authHeader, "Bearer ")

		// Extraire les informations du token
		tokenInfo, err := s.extractTokenInfo(c.Request.Context(), tokenString)
		if err != nil {
			log.Printf("Erreur d'extraction du token: %v", err)
			c.JSON(http.StatusUnauthorized, gin.H{
				"error":   "Unauthorized",
				"message": "Token invalide",
			})
			c.Abort()
			return
		}

		// Vérifier les teams requises
		if len(teams) > 0 && !s.hasRequiredTeams(tokenInfo, teams) {
			log.Printf("Accès refusé - Teams insuffisantes")
			c.JSON(http.StatusForbidden, gin.H{
				"error":   "Forbidden",
				"message": "Accès non autorisé",
			})
			c.Abort()
			return
		}

		// Ajouter les informations du token dans les headers
		s.setTokenHeaders(c, tokenInfo, tokenString)

		log.Printf("Token validé pour: %s (%s)", tokenInfo.Name, tokenInfo.Email)
		c.Next()
	}
}

// // verifyToken vérifie la validité du token auprès du serveur OAuth2
// func (s *proxyServer) verifyToken(ctx context.Context, tokenString string) (*TokenInfo, error) {
// 	// Utiliser le endpoint tokeninfo pour valider le token
// 	// Vous pouvez utiliser une librairie comme golang.org/x/oauth2
// 	// ou faire une requête HTTP directe vers le endpoint tokeninfo

// 	token := &oauth2.Token{AccessToken: tokenString}

// 	// Vérifier avec le client OAuth2
// 	httpClient := oauth2.NewClient(ctx, oauth2.StaticTokenSource(token))

// 	// Faire une requête vers le endpoint userinfo ou tokeninfo
// 	// pour récupérer les informations du token
// 	userInfoURL := strings.Replace(s.cfg.Server.OAuth2.Endpoints.TokenURL,
// 		"token", "userinfo", 1)

// 	resp, err := httpClient.Get(userInfoURL)
// 	if err != nil {
// 		return nil, err
// 	}
// 	defer resp.Body.Close()

// 	if resp.StatusCode != http.StatusOK {
// 		return nil, fmt.Errorf("token validation failed: %s", resp.Status)
// 	}

// 	// Parser la réponse et retourner les informations du token
// 	var tokenInfo TokenInfo
// 	if err := json.NewDecoder(resp.Body).Decode(&tokenInfo); err != nil {
// 		return nil, err
// 	}

// 	return &tokenInfo, nil
// }

// // hasRequiredScopes vérifie si le token a les scopes requis
// func (s *proxyServer) hasRequiredScopes(tokenScopes string) bool {
// 	// Implémentation basique - à adapter selon vos besoins
// 	requiredScopes := []string{"openid", "profile"}
// 	scopes := strings.Split(tokenScopes, " ")

// 	for _, required := range requiredScopes {
// 		found := false
// 		for _, scope := range scopes {
// 			if scope == required {
// 				found = true
// 				break
// 			}
// 		}
// 		if !found {
// 			return false
// 		}
// 	}
// 	return true
// }

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

// Middleware simplifié pour les routes publiques
func (s *proxyServer) publicMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Pour les routes publiques, on peut juste logger l'accès
		log.Printf("Accès public à: %s", c.Request.URL.Path)
		c.Next()
	}
}
