package commands

import (
	"log/slog"
	"s-vitaliy/kubectl-plugin-arcane/internal/handlers"
)

type Context struct {
	Logger    *slog.Logger
	ApiClient handlers.StreamCommandHandler
}
