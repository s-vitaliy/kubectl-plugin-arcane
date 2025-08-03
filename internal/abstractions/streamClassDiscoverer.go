package abstractions

import (
	"context"

	"s-vitaliy/kubectl-plugin-arcane/internal/models"
)

type ApiSettingsDiscoverer interface {
	DiscoveryFromJobs(ctx context.Context, jobName string, namespace string) (*models.ClientApiSettings, error)
	DiscoveryFromStreamClass(ctx context.Context, streamClass string, namespace string) (*models.ClientApiSettings, error)
}
