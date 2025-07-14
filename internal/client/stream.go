package client

type StreamSuspendHandlerer interface {

	/// Suspend suspends the stream with the given ID.
	/// It returns an error if the operation fails.
	Suspend(id string) error
}

type StreamBackfillHandler interface {

	/// Backfill restarts the stream with the given ID in backfill mode.
	/// It returns an error if the operation fails.
	Backfill(id string, watch bool) error
}

type StreamRestartHandler interface {

	/// Backfill restarts the stream with the given ID in backfill mode.
	/// It returns an error if the operation fails.
	Backfill(id string, wait bool) error
}
