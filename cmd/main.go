package main

import (
	"os"

	"log/slog"

	"internal/client/api"

	"github.com/alecthomas/kong"
)

var CLI struct {
	Stream StreamCmd `cmd:"" help:"Manage Arcane streams."`
}

const AppDescription = "A command line tool for managing the Arcane streams."

func main() {
	handler := slog.NewTextHandler(os.Stdout, nil)
	logger := slog.New(handler)
	apiClient := api.NewAnnotationStreamCommandHandler(&api.HandlerContext{
		Logger: logger,
	})

	executableName := getExecutableName()
	ctx := kong.Parse(&CLI, kong.Name(executableName), kong.Description(AppDescription))
	err := ctx.Run(&Context{logger: logger, apiClient: apiClient})
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

type Context struct {
	logger    *slog.Logger
	apiClient StreamCommandHandler
}
