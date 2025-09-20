package token

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/golang-jwt/jwt"
)

// TokenResponse représente la réponse du serveur d'autorisation
type TokenResponse struct {
	AccessToken  string `json:"access_token"`
	TokenType    string `json:"token_type"`
	ExpiresIn    int    `json:"expires_in"`
	RefreshToken string `json:"refresh_token"`
	Scope        string `json:"scope"`
	Error        string `json:"error,omitempty"`
	ErrorDesc    string `json:"error_description,omitempty"`
}

// OAuthConfig contient les paramètres de configuration OAuth2
type OAuthConfig struct {
	TokenURL     string
	ClientID     string
	ClientSecret string
	Username     string
	Password     string
	Scope        string // optionnel
}

// GetAccessToken obtient un token d'accès via le grant type "password"
func GetAccessToken(ctx context.Context, config OAuthConfig) (string, error) {
	// Créer un client HTTP qui ignore les certificats SSL non valides (comme curl -k)
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	client := &http.Client{
		Transport: tr,
		Timeout:   30 * time.Second,
	}

	// Préparer les données du formulaire
	data := url.Values{}
	data.Set("grant_type", "password")
	data.Set("client_id", config.ClientID)
	data.Set("client_secret", config.ClientSecret)
	data.Set("username", config.Username)
	data.Set("password", config.Password)

	if config.Scope != "" {
		data.Set("scope", config.Scope)
	}

	// Créer la requête POST
	req, err := http.NewRequestWithContext(ctx, "POST", config.TokenURL, strings.NewReader(data.Encode()))
	if err != nil {
		return "", fmt.Errorf("erreur lors de la création de la requête: %w", err)
	}

	// Définir les headers
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	// Effectuer la requête
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("erreur lors de l'exécution de la requête: %w", err)
	}
	defer resp.Body.Close()

	// Lire la réponse
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("erreur lors de la lecture de la réponse: %w", err)
	}

	// Parser la réponse JSON
	var tokenResp TokenResponse
	if err := json.Unmarshal(body, &tokenResp); err != nil {
		return "", fmt.Errorf("erreur lors du parsing JSON: %w", err)
	}

	// Vérifier les erreurs dans la réponse
	if tokenResp.Error != "" {
		return "", fmt.Errorf("erreur OAuth2: %s - %s", tokenResp.Error, tokenResp.ErrorDesc)
	}

	// Vérifier que nous avons bien reçu un token
	if tokenResp.AccessToken == "" {
		return "", fmt.Errorf("aucun access_token reçu dans la réponse")
	}

	return tokenResp.AccessToken, nil
}

// GetTokenResponse retourne la réponse complète au lieu de juste le token
func GetTokenResponse(ctx context.Context, config OAuthConfig) (*TokenResponse, error) {
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	client := &http.Client{
		Transport: tr,
		Timeout:   30 * time.Second,
	}

	data := url.Values{}
	data.Set("grant_type", "password")
	data.Set("client_id", config.ClientID)
	data.Set("client_secret", config.ClientSecret)
	data.Set("username", config.Username)
	data.Set("password", config.Password)

	if config.Scope != "" {
		data.Set("scope", config.Scope)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", config.TokenURL, strings.NewReader(data.Encode()))
	if err != nil {
		return nil, fmt.Errorf("erreur lors de la création de la requête: %w", err)
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("erreur lors de l'exécution de la requête: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("erreur lors de la lecture de la réponse: %w", err)
	}

	var tokenResp TokenResponse
	if err := json.Unmarshal(body, &tokenResp); err != nil {
		return nil, fmt.Errorf("erreur lors du parsing JSON: %w", err)
	}

	if tokenResp.Error != "" {
		return &tokenResp, fmt.Errorf("erreur OAuth2: %s - %s", tokenResp.Error, tokenResp.ErrorDesc)
	}

	return &tokenResp, nil
}

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
