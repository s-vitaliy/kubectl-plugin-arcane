package main

import (
	"os"

	"github.com/alecthomas/kong"
)

var CLI struct {
	Stream struct {
		Suspend struct {
			ID string `arg:"" help:"The ID of the stream to suspend."`
		} `cmd:"" help:"Suspends the given stream."`
		Backfill struct {
			ID string `arg:"" help:"The ID of the stream to backfill."`
		} `cmd:"" help:"Restarts the given stream in the backfill mode."`
		Restart struct {
			ID string `arg:"" help:"The ID of the stream to restart."`
		} `cmd:"" help:"Restarts the given stream in the streaming mode."`
	} `cmd:"" help:"Stream operation."`
}

const AppDescription = "A command line tool for managing the Arcane streams."

func main() {
	executableName := getExecutableName()
	ctx := kong.Parse(&CLI, kong.Name(executableName), kong.Description(AppDescription))
	println("Parsed context:", ctx)
}

func getExecutableName() string {
	// Not checking for errors here since argv[0] should always be available
	return os.Args[0]
}
