package abstractions

import (
	"context"

	"s-vitaliy/kubectl-plugin-arcane/internal/models"

	"k8s.io/client-go/dynamic"
)

type ApiSettingsDiscoverer interface {
	DiscoveryFromJobs(ctx context.Context, dynamicInterface dynamic.Interface, jobName string, namespace string) (*models.ClientApiSettings, error)
	DiscoveryFromStreamClass(ctx context.Context, dynamicInterface dynamic.Interface, streamClass string, namespace string) (*models.ClientApiSettings, error)
}
