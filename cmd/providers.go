package main

import (
	"s-vitaliy/kubectl-plugin-arcane/internal/app"
	"s-vitaliy/kubectl-plugin-arcane/internal/client/api"
	"s-vitaliy/kubectl-plugin-arcane/internal/handlers"
)

func provideStreamCommandHandler(configReader app.ConfigReader) (handlers.StreamCommandHandler, error) {
	return api.NewAnnotationStreamCommandHandlerV1(configReader), nil
}

func provideConfigReader() (app.ConfigReader, error) {
	return &app.FileConfigReader{ConfigOverride: ""}, nil
}
