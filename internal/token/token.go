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
