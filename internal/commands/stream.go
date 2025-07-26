package commands

import (
	"fmt"
	"s-vitaliy/kubectl-plugin-arcane/internal/handlers"
)

// Represents the command to suspend a stream.
type SuspendCmd struct {
	Id string `arg:"" help:"The ID of the stream to suspend."`
}

// Represents the command to backfill a stream.
type BackfillCmd struct {
	Id string `arg:"" help:"The ID of the stream to backfill."`
}

// Represents the command to restart a stream.
type RestartCmd struct {
	Id string `arg:"" help:"The ID of the stream to backfill."`
}

// The Stream interaction commmands.
type StreamCmd struct {
	Suspend  SuspendCmd  `cmd:"" help:"Suspends the given stream."`
	Backfill BackfillCmd `cmd:"" help:"Restarts the given stream in the backfill mode."`
	Restart  RestartCmd  `cmd:"" help:"Restarts the given stream in the streaming mode."`
}

func (r *SuspendCmd) Run(ctx *Context) error {
	err := ctx.Container.Invoke(func(h handlers.StreamCommandHandler) error {
		if h != nil {
			return h.Suspend(r.Id)
		}
		return fmt.Errorf("no handler provided for suspending stream")
	})
	return err
}
