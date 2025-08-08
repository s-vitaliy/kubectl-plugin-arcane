package abstractions

import "context"

type StreamSuspendHandlerer interface {

	/// Suspends the stream with the given ID.
	/// It returns an error if the operation fails.
	Suspend(ctx context.Context, id string) error
}

type StreamResumeHandlerer interface {

	/// Resumes the stream with the given ID.
	/// It returns an error if the operation fails.
	Resume(ctx context.Context, id string, streamClass string) error
}

type StreamBackfillHandler interface {

	/// Backfill restarts the stream with the given ID in backfill mode.
	/// It returns an error if the operation fails.
	Backfill(ctx context.Context, id string, watch bool) error
}

type StreamRestartHandler interface {

	/// Backfill restarts the stream with the given ID in backfill mode.
	/// It returns an error if the operation fails.
	Restart(ctx context.Context, id string, wait bool) error
}

type StreamCommandHandler interface {
	StreamSuspendHandlerer
	StreamResumeHandlerer
	StreamBackfillHandler
	StreamRestartHandler
}
