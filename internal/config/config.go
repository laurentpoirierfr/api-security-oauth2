package config

// Config represents the application configuration structure
type Config struct {
	Application Application `mapstructure:"application"`
	Server      Server      `mapstructure:"server"`
	Routes      []Route     `mapstructure:"routes"`
}

// Application holds metadata about the application
type Application struct {
	Name        string `mapstructure:"name"`
	Description string `mapstructure:"description"`
	Version     string `mapstructure:"version"`
}

// Server holds server-related configuration
type Server struct {
	Port          string `mapstructure:"port"`
	DefaultTarget string `mapstructure:"default_target"`
	TimeOut       int    `mapstructure:"timeout"`
	OAuth2        OAuth2 `mapstructure:"oauth2"`
}

// Route defines a routing rule
type Route struct {
	Path   string `mapstructure:"path"`
	Target string `mapstructure:"target"`
	Teams  []Team `mapstructure:"teams"`
}

// Team defines a team with name and description
type Team struct {
	Name        string `mapstructure:"name"`
	Description string `mapstructure:"description"`
}

// OAuth2 holds OAuth2-related configuration
type OAuth2 struct {
	// ClientID     string          `mapstructure:"client_id"`
	// ClientSecret string          `mapstructure:"client_secret"`
	// RedirectURL  string          `mapstructure:"redirect_url"`
	Endpoints OAuth2Endpoints `mapstructure:"endpoints"`
}

// OAuth2Endpoints holds the URLs for various OAuth2 endpoints
type OAuth2Endpoints struct {
	AuthURL      string `mapstructure:"auth_url"`
	TokenURL     string `mapstructure:"token_url"`
	TokenInfoURL string `mapstructure:"tokeninfo_url"`
	UserInfoURL  string `mapstructure:"userinfo_url"`
}
