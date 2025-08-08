package v0

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"s-vitaliy/kubectl-plugin-arcane/internal/abstractions"
	"s-vitaliy/kubectl-plugin-arcane/internal/models"

	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/dynamic"
)

type streamClassOperationService struct {
	client dynamic.Interface
	logger *slog.Logger
}

var _ abstractions.StreamClassOperator = &streamClassOperationService{}

// ProvideStreamClassOperationService provides a new StreamClassOperator implementation.
func ProvideStreamClassOperationService(client dynamic.Interface, logger *slog.Logger) abstractions.StreamClassOperator {
	return &streamClassOperationService{
		client: client,
		logger: logger,
	}
}

// Suspend implements abstractions.StreamClassOperator.
func (s *streamClassOperationService) Suspend(ctx context.Context, id string, namespace string, apiSettings *models.ClientApiSettings) error {
	annotation := map[string]any{
		"metadata": map[string]any{
			"annotations": map[string]string{
				"arcane/state": "suspended",
			},
		},
	}
	patchBytes, err := json.Marshal(annotation)
	if err != nil {
		return fmt.Errorf("failed to marshal suspend annotation: %w", err)
	}
	dynamicClient := s.client.Resource(apiSettings.ToGroupVersionResource()).Namespace(namespace)

	_, err = dynamicClient.Patch(ctx, id, types.MergePatchType, patchBytes, v1.PatchOptions{})

	if err != nil {
		s.logger.Error("Failed to suspend stream", "id", id, "error", err)
		return fmt.Errorf("failed to suspend stream %s: %w", id, err)
	}
	return nil
}

// Resume implements abstractions.StreamClassOperator.
func (s *streamClassOperationService) Resume(ctx context.Context, id string, namespace string, apiSettings *models.ClientApiSettings) error {
	annotation := map[string]any{
		"metadata": map[string]any{
			"annotations": map[string]any{
				"arcane/state": nil,
			},
		},
	}
	dynamicClient := s.client.Resource(apiSettings.ToGroupVersionResource()).Namespace(namespace)

	patchBytes, err := json.Marshal(annotation)
	if err != nil {
		return fmt.Errorf("failed to marshal resume annotation: %w", err)
	}
	_, err = dynamicClient.Patch(ctx,
		id,
		types.MergePatchType,
		patchBytes,
		v1.PatchOptions{})
	if err != nil {
		s.logger.Error("Failed to resume stream", "id", id, "error", err)
		return fmt.Errorf("failed to resume stream %s: %w", id, err)
	}
	s.logger.Info("Stream resumed successfully", "id", id)
	return nil
}
