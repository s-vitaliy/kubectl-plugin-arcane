package commands

import (
	"context"
	"fmt"
	"s-vitaliy/kubectl-plugin-arcane/internal/abstractions"

	"go.uber.org/dig"
)

// Represents the command to suspend a stream.
type SuspendCmd struct {
	Id string `arg:"" help:"The ID of the stream to suspend."`
}

// Represents the command to resume a stream.
type ResumeCmd struct {
	Id    string `arg:"" help:"The ID of the stream to resume."`
	Class string `arg:"" help:"The class of the stream to resume."`
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
	Resume   ResumeCmd   `cmd:"" help:"Resumes the given stream."`
	Backfill BackfillCmd `cmd:"" help:"Restarts the given stream in the backfill mode."`
	Restart  RestartCmd  `cmd:"" help:"Restarts the given stream in the streaming mode."`
}

func (r *SuspendCmd) Run(container *dig.Container) error {
	err := container.Invoke(func(h abstractions.StreamCommandHandler) error {
		if h != nil {
			return h.Suspend(context.Background(), r.Id)
		}
		return fmt.Errorf("no handler provided for suspending stream")
	})
	return err
}

func (r *ResumeCmd) Run(container *dig.Container) error {
	err := container.Invoke(func(h abstractions.StreamCommandHandler) error {
		if h != nil {
			return h.Resume(context.Background(), r.Id, r.Class)
		}
		return fmt.Errorf("no handler provided for resuming stream")
	})
	return err
}
