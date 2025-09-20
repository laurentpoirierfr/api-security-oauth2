package config

import (
	"bytes"
	"strings"

	"github.com/spf13/viper"
)

func NewConfig(embeddedConfig []byte) *Config {

	v := viper.New()
	// Set the file type of the configuration
	v.SetConfigType("yaml")

	// Environment variables support
	// Convert dots in config keys to underscores in env vars
	// E.g., application.name -> APPLICATION_NAME
	v.SetEnvKeyReplacer(strings.NewReplacer(`.`, `_`))
	// Automatically read environment variables
	v.AutomaticEnv()

	if err := v.ReadConfig(bytes.NewBuffer(embeddedConfig)); err != nil {
		panic(err)
	}
	var cfg Config
	if err := v.Unmarshal(&cfg); err != nil {
		panic(err)
	}
	return &cfg
}
