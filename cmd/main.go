package main

import (
	"os"

	"log/slog"

	"s-vitaliy/kubectl-plugin-arcane/internal/client/api"
	"s-vitaliy/kubectl-plugin-arcane/internal/commands"

	"github.com/alecthomas/kong"
	"go.uber.org/dig"
)

var CLI struct {
	Stream commands.StreamCmd `cmd:"" help:"Manage Arcane streams."`
}

const AppDescription = "A command line tool for managing the Arcane streams."

func main() {
	handler := slog.NewTextHandler(os.Stdout, nil)
	logger := slog.New(handler)
	apiClient := api.NewAnnotationStreamCommandHandler(&api.HandlerContext{
		Logger: logger,
	})

	container := dig.New()
	container.Provide(provideStreamCommandHandler)
	container.Provide(provideConfigReader)

	executableName := getExecutableName()
	ctx := kong.Parse(&CLI, kong.Name(executableName), kong.Description(AppDescription))
	err := ctx.Run(&commands.Context{Logger: logger, ApiClient: apiClient, Container: container})

	if err != nil {
		logger.Error("Command execution failed", slog.String("command", ctx.Command()), slog.String("error", err.Error()))
		os.Exit(1)
	}
	logger.Info("Command executed successfully", slog.String("command", ctx.Command()))
}

func getExecutableName() string {
	// Not checking for errors here since argv[0] should always be available
	return os.Args[0]
}
