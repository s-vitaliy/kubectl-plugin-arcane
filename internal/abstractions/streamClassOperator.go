package abstractions

import (
	"context"
	"s-vitaliy/kubectl-plugin-arcane/internal/models"
)

type StreamClassOperator interface {
	// Suspend suspends a running stream by its ID.
	Suspend(ctx context.Context, id string, namespace string, apiSettings *models.ClientApiSettings) error

	// Resume resumes a suspended stream by its ID.
	Resume(ctx context.Context, id string, namespace string, apiSettings *models.ClientApiSettings) error
}
