package abstractions

import (
	"context"
	"s-vitaliy/kubectl-plugin-arcane/internal/models"
)

type StreamPhase int

const (
	StreamPhaseRunning StreamPhase = iota
	StreamPhaseSuspended
	StreamPhaseBackfill
	StreamPhaseFailed
)

var stateName = map[StreamPhase]string{
	StreamPhaseRunning:   "Running",
	StreamPhaseSuspended: "Suspended",
	StreamPhaseBackfill:  "Reloading",
	StreamPhaseFailed:    "Failed",
}

func (ss StreamPhase) String() string {
	return stateName[ss]
}

type StreamClassOperator interface {
	// Suspend suspends a running stream by its ID.
	Suspend(ctx context.Context, id string, namespace string, apiSettings *models.ClientApiSettings) error

	// Resume resumes a suspended stream by its ID.
	Resume(ctx context.Context, id string, namespace string, apiSettings *models.ClientApiSettings) error

	// WaitForStatus waits for the stream to reach the desired status.
	WaitForStatus(ctx context.Context, status StreamPhase, id string, namespace string, apiSettings *models.ClientApiSettings) error

	// Backfill restarts the stream in backfill mode.
	Backfill(ctx context.Context, id string, namespace string, clientApiSettings *models.ClientApiSettings) error
}
