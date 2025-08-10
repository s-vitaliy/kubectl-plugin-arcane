package v0

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"s-vitaliy/kubectl-plugin-arcane/internal/abstractions"
	"s-vitaliy/kubectl-plugin-arcane/internal/models"
	"strings"

	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
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
	return s.patchObject(ctx, id, namespace, apiSettings, annotation)
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
	return s.patchObject(ctx, id, namespace, apiSettings, annotation)
}

// Backfill implements abstractions.StreamClassOperator.
func (s *streamClassOperationService) Backfill(ctx context.Context, id string, namespace string, apiSettings *models.ClientApiSettings) error {
	s.logger.Info("Restarting the stream in backfill mode", "id", id)
	annotation := map[string]any{
		"metadata": map[string]any{
			"annotations": map[string]string{
				"arcane/state": "reload-requested",
			},
		},
	}
	return s.patchObject(ctx, id, namespace, apiSettings, annotation)
}

// WaitForStatus implements abstractions.StreamClassOperator.
func (s *streamClassOperationService) WaitForStatus(ctx context.Context, targetPhase abstractions.StreamPhase, id string, namespace string, apiSettings *models.ClientApiSettings) error {
	dynamicClient := s.client.Resource(apiSettings.ToGroupVersionResource()).Namespace(namespace)
	watcher, err := dynamicClient.Watch(ctx, v1.ListOptions{
		FieldSelector: fmt.Sprintf("metadata.name=%s", id),
	})
	if err != nil {
		return fmt.Errorf("failed to watch stream %s: %w", id, err)
	}
	defer watcher.Stop()

	s.logger.Info("Waiting for stream status", "id", id, "targetPhase", targetPhase)

	for {
		select {
		case <-ctx.Done():
			return fmt.Errorf("context cancelled while waiting for stream %s status", id)
		case event, ok := <-watcher.ResultChan():
			if !ok {
				return fmt.Errorf("watch channel closed for stream %s", id)
			}

			s.logger.Info("Received stream status update", "id", id)
			stream, ok := event.Object.(*unstructured.Unstructured)

			if !ok {
				return fmt.Errorf("watch channel closed for stream %s", id)
			}

			if !ok {
				return fmt.Errorf("unexpected type %T received from watch channel", event.Object)
			}

			phase, found, err := unstructured.NestedString(stream.Object, "status", "phase")
			if err != nil {
				return fmt.Errorf("failed to get phase from stream %s: %w", id, err)
			}
			if !found {
				return fmt.Errorf("phase not found in stream %s", id)
			}

			s.logger.Info("Stream status update", "id", id, "phase", phase)
			if strings.EqualFold(phase, targetPhase.String()) {
				s.logger.Info("Stream reached desired status", "id", id, "status", phase)
				return nil
			}
		}
	}
}

func (s *streamClassOperationService) patchObject(ctx context.Context, id string, namespace string, apiSettings *models.ClientApiSettings, annotation map[string]any) error {
	s.logger.Debug("Patching stream object", "id", id, "namespace", namespace, "apiSettings", apiSettings)
	if len(annotation) == 0 {
		return fmt.Errorf("no annotations provided for patching stream %s", id)
	}

	dynamicClient := s.client.Resource(apiSettings.ToGroupVersionResource()).Namespace(namespace)
	patchBytes, err := json.Marshal(annotation)
	if err != nil {
		return fmt.Errorf("failed to marshal annotation: %w", err)
	}
	_, err = dynamicClient.Patch(ctx,
		id,
		types.MergePatchType,
		patchBytes,
		v1.PatchOptions{})
	if err != nil {
		s.logger.Error("Failed to patch stream", "id", id, "error", err)
		return fmt.Errorf("failed to patch stream %s: %w", id, err)
	}
	s.logger.Info("Stream patched successfully", "id", id)
	return nil
}
