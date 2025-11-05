package main

import (
	"os"

	"log/slog"

	"s-vitaliy/kubectl-plugin-arcane/internal/app"
	"s-vitaliy/kubectl-plugin-arcane/internal/client/api/common"
	v0 "s-vitaliy/kubectl-plugin-arcane/internal/client/api/v0"
	"s-vitaliy/kubectl-plugin-arcane/internal/commands"

	"github.com/alecthomas/kong"
	"go.uber.org/dig"
)

var CLI struct {
	Stream commands.StreamCmd `cmd:"" help:"Manage Arcane streams."`
}

const AppDescription = "A command line tool for managing the Arcane streams."

func main() { // coverage-ignore
	handler := slog.NewTextHandler(os.Stdout, nil)
	logger := slog.New(handler)
	container := dig.New()

	err := container.Provide(func() *slog.Logger {
		return logger
	})

	if err != nil {
		logger.Error("Failed to provide logger", slog.String("error", err.Error()))
		os.Exit(1)
	}

	err = container.Provide(app.ProvideStreamCommandHandler)
	if err != nil {
		logger.Error("Failed to provide stream command handler", slog.String("error", err.Error()))
		os.Exit(1)
	}

	err = container.Provide(app.ProvideConfigReader)
	if err != nil {
		logger.Error("Failed to provide config reader", slog.String("error", err.Error()))
		os.Exit(1)
	}

	err = container.Provide(common.ProvideStreamClassDiscoveryService)
	if err != nil {
		logger.Error("Failed to provide stream class discovery service", slog.String("error", err.Error()))
		os.Exit(1)
	}

	err = container.Provide(common.ProvideDynamicClient)
	if err != nil {
		logger.Error("Failed to provide dynamic client", slog.String("error", err.Error()))
		os.Exit(1)
	}

	err = container.Provide(v0.ProvideStreamClassOperationService)
	if err != nil {
		logger.Error("Failed to provide stream class operation service", slog.String("error", err.Error()))
		os.Exit(1)
	}

	executableName := getExecutableName()
	command := kong.Parse(&CLI, kong.Name(executableName), kong.Description(AppDescription))
	err = command.Run(container)

	if err != nil {
		logger.Error("Command execution failed", slog.String("command", command.Command()), slog.String("error", err.Error()))
		os.Exit(1)
	}
	logger.Info("Command executed successfully", slog.String("command", command.Command()))
}

func getExecutableName() string { // coverage-ignore
	// Not checking for errors here since argv[0] should always be available
	return os.Args[0]
}
