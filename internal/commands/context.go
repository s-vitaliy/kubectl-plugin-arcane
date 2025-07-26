package commands

import (
	"log/slog"
	"s-vitaliy/kubectl-plugin-arcane/internal/handlers"

	"go.uber.org/dig"
)

type Context struct {
	Logger    *slog.Logger
	ApiClient handlers.StreamCommandHandler
	Container *dig.Container
}
