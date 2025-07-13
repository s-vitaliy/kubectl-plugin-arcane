package main

import (
	"fmt"
	"os"

	"github.com/alecthomas/kong"
)

type SuspendCmd struct {
	Id string `arg:"" help:"The ID of the stream to suspend."`
}

type BackfillCmd struct {
	Id string `arg:"" help:"The ID of the stream to backfill."`
}

type RestartCmd struct {
	Id string `arg:"" help:"The ID of the stream to backfill."`
}

type StreamCmd struct {
	Suspend  SuspendCmd  `cmd:"" help:"Suspends the given stream."`
	Backfill BackfillCmd `cmd:"" help:"Restarts the given stream in the backfill mode."`
	Restart  RestartCmd  `cmd:"" help:"Restarts the given stream in the streaming mode."`
}

var CLI struct {
	Stream StreamCmd `cmd:"" help:"Manage Arcane streams."`
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

type Context struct {
	// Add any context-specific fields here if needed
}

func (r *SuspendCmd) Run(ctx *Context) error {
	fmt.Println("rm", r.Id)
	return nil
}
