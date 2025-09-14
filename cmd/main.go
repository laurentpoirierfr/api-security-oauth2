package main

import (
	_ "embed"

	"github.com/laurentpoirierfr/api-security-oauth2/internal/config"
	"github.com/laurentpoirierfr/api-security-oauth2/internal/server"
)

//go:embed config.yaml
var embeddedConfig []byte

func main() {
	server := server.NewServer(config.NewConfig(embeddedConfig))
	if err := server.Start(); err != nil {
		panic(err)
	}
}
