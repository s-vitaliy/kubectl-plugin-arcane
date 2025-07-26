package commands

import (
	"log/slog"

	"go.uber.org/dig"
)

type Context struct {
	Logger    *slog.Logger
	Container *dig.Container
}
