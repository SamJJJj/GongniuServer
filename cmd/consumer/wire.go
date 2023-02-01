// +build wireinject

package main

import (
	"demo/internal/server"
	eagle "github.com/go-eagle/eagle/pkg/app"
	"github.com/google/wire"
)

func InitApp(cfg *eagle.Config, config *eagle.ServerConfig) (*eagle.App, error) {
	wire.Build(server.ProviderSet, newApp)
	return &eagle.App{}, nil
}
