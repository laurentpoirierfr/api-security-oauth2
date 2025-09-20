package token

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/golang-jwt/jwt"
)

type TokenInfo struct {
	Sub         string `json:"sub"`
	UID         string `json:"uid"`
	Email       string `json:"email"`
	Name        string `json:"name"`
	GivenName   string `json:"given_name"`
	FamilyName  string `json:"family_name"`
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

// HasRealmRole vérifie si le rôle spécifié est présent dans RealmAccess
func (t *TokenInfo) HasRealmRole(role string) bool {
	for _, r := range t.RealmAccess.Roles {
		if r == role {
			return true
		}
	}
	return false
}

// Méthode String pour afficher les informations du token de manière lisible
func (t *TokenInfo) String() string {
	data, err := json.MarshalIndent(t, "", "  ")
	if err != nil {
		return fmt.Sprintf("erreur lors du formatage du token info: %v", err)
	}
	return string(data)
}

// Fonction utilitaire pour vérifier si le token est expiré
func (ti *TokenInfo) IsExpired() bool {
	return time.Now().Unix() > ti.Expiration
}

// Fonction utilitaire pour obtenir le temps restant avant expiration
func (ti *TokenInfo) TimeToExpiration() time.Duration {
	return time.Unix(ti.Expiration, 0).Sub(time.Now())
}

// extractTokenInfo extrait et parse les informations du token JWT
func ExtractTokenInfo(ctx context.Context, tokenString string) (*TokenInfo, error) {
	// Nettoyer le token (enlever "Bearer " si présent)
	tokenString = strings.TrimPrefix(tokenString, "Bearer ")

	// Parser le token sans vérification de signature (pour extraire les claims)
	// Note: En production, vous devriez vérifier la signature avec la clé publique appropriée
	token, _, err := new(jwt.Parser).ParseUnverified(tokenString, jwt.MapClaims{})
	if err != nil {
		return nil, fmt.Errorf("erreur lors du parsing du token: %w", err)
	}

	// Extraire les claims
	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return nil, fmt.Errorf("impossible d'extraire les claims du token")
	}

	// Vérifier l'expiration
	if exp, ok := claims["exp"].(float64); ok {
		if time.Now().Unix() > int64(exp) {
			return nil, fmt.Errorf("le token a expiré")
		}
	}

	// Convertir les claims en JSON puis les unmarshaler dans TokenInfo
	claimsJSON, err := json.Marshal(claims)
	if err != nil {
		return nil, fmt.Errorf("erreur lors de la conversion des claims en JSON: %w", err)
	}

	var tokenInfo TokenInfo
	if err := json.Unmarshal(claimsJSON, &tokenInfo); err != nil {
		return nil, fmt.Errorf("erreur lors de l'unmarshaling des claims: %w", err)
	}

	return &tokenInfo, nil
}
