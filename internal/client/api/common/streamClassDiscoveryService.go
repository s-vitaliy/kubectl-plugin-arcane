package common

import (
	"context"
	"fmt"
	"log/slog"
	"s-vitaliy/kubectl-plugin-arcane/internal/abstractions"
	"s-vitaliy/kubectl-plugin-arcane/internal/models"

	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"
)

type streamClassDiscoveryService struct {
	logger           *slog.Logger
	dynamicInterface dynamic.Interface
}

var _ abstractions.ApiSettingsDiscoverer = &streamClassDiscoveryService{}

// ProvideStreamClassDiscoveryService provides a new instance of streamClassDiscoveryService.
func ProvideStreamClassDiscoveryService(logger *slog.Logger, dynamicInterface dynamic.Interface) abstractions.ApiSettingsDiscoverer {
	if logger == nil {
		logger = slog.Default()
	}
	return &streamClassDiscoveryService{logger: logger, dynamicInterface: dynamicInterface}
}

// DiscoveryFromJobs discovers API settings from a job.
func (s *streamClassDiscoveryService) DiscoveryFromJobs(ctx context.Context, jobstreamClass string, namespace string) (*models.ClientApiSettings, error) {
	jobResourceRef := schema.GroupVersionResource{Group: "batch", Version: "v1", Resource: "jobs"}
	dynamicClient := s.dynamicInterface.Resource(jobResourceRef).Namespace(namespace)

	jobValue, err := dynamicClient.Get(ctx, jobstreamClass, v1.GetOptions{})
	if err != nil {
		s.logger.Error("Failed to get job", "namespace", "streamClass", namespace, jobstreamClass, "error", err)
		return nil, fmt.Errorf("failed to get job %s: %w", jobstreamClass, err)
	}

	metadata, ok := jobValue.Object["metadata"].(map[string]any)
	if !ok {
		s.logger.Error("Failed to get metadata from job", "namespace", "streamClass", namespace, jobstreamClass)
		return nil, fmt.Errorf("failed to get metadata from job %s", jobstreamClass)
	}

	annotations, ok := metadata["annotations"].(map[string]any)
	if !ok {
		s.logger.Error("Failed to get annotations from job metadata", "namespace", "streamClass", namespace, jobstreamClass)
		return nil, fmt.Errorf("failed to get annotations from job %s metadata", jobstreamClass)
	}

	s.logger.Debug("Annotations from job", "namespace", "streamClass", namespace, jobstreamClass, "annotations", annotations)
	return models.FromJobAnnotations(annotations)
}

// DiscoveryFromStreamClass discovers API settings from a stream class.
func (s *streamClassDiscoveryService) DiscoveryFromStreamClass(ctx context.Context, streamClass string, namespace string) (*models.ClientApiSettings, error) {
	streamClassResourceRef := schema.GroupVersionResource{
		Group:    "streaming.sneaksanddata.com",
		Version:  "v1beta1",
		Resource: "stream-classes",
	}
	dynamicClient := s.dynamicInterface.Resource(streamClassResourceRef).Namespace(namespace)
	streamClassValue, err := dynamicClient.Get(ctx, streamClass, v1.GetOptions{})

	if err != nil {
		return nil, fmt.Errorf("failed to get stream class %s: %w", streamClass, err)
	}

	spec, ok := streamClassValue.Object["spec"].(map[string]any)
	if !ok {
		return nil, fmt.Errorf("failed to get spec from stream class %s", streamClass)
	}

	apiGroup, ok := spec["apiGroupRef"].(string)
	if !ok {
		return nil, fmt.Errorf("failed to get apiGroup from stream class %s", streamClass)
	}

	apiVersion, ok := spec["apiVersion"].(string)
	if !ok {
		return nil, fmt.Errorf("failed to get apiVersion from stream class %s", streamClass)
	}

	apiPlural, ok := spec["pluralName"].(string)
	if !ok {
		return nil, fmt.Errorf("failed to get apiPlural from stream class %s", streamClass)
	}

	settings := models.NewClientApiSettings(apiGroup, apiVersion, apiPlural)
	return settings, nil
}
