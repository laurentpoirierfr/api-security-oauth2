package config

import (
	"bytes"

	"github.com/spf13/viper"
)

type Config struct {
	Application Application `mapstructure:"application"`
	Server      Server      `mapstructure:"server"`
	Routes      []Route     `mapstructure:"routes"`
}

func NewConfig(embeddedConfig []byte) *Config {
	v := viper.New()
	v.SetConfigType("yaml")

	if err := v.ReadConfig(bytes.NewBuffer(embeddedConfig)); err != nil {
		panic(err)
	}
	var cfg Config
	if err := v.Unmarshal(&cfg); err != nil {
		panic(err)
	}
	return &cfg
}

type Application struct {
	Name        string `mapstructure:"name"`
	Description string `mapstructure:"description"`
	Version     string `mapstructure:"version"`
}

type Server struct {
	Port       string `mapstructure:"port"`
	BackendURL string `mapstructure:"backend_url"`
	TimeOut    int    `mapstructure:"timeout"`
	OAuth2     OAuth2 `mapstructure:"oauth2"`
}

type Route struct {
	Path  string `mapstructure:"path"`
	Teams []Team `mapstructure:"teams"`
}

type Team struct {
	Name        string `mapstructure:"name"`
	Description string `mapstructure:"description"`
}
type OAuth2 struct {
	// ClientID     string          `mapstructure:"client_id"`
	// ClientSecret string          `mapstructure:"client_secret"`
	// RedirectURL  string          `mapstructure:"redirect_url"`
	Endpoints OAuth2Endpoints `mapstructure:"endpoints"`
}

type OAuth2Endpoints struct {
	AuthURL      string `mapstructure:"auth_url"`
	TokenURL     string `mapstructure:"token_url"`
	TokenInfoURL string `mapstructure:"tokeninfo_url"`
}
