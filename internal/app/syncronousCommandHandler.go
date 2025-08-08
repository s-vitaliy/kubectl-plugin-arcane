package app

import (
	"context"
	"fmt"
	"log/slog"

	"s-vitaliy/kubectl-plugin-arcane/internal/abstractions"
)

var NAMESPACE = "arcane"

type SyncronousCommandHandler struct {
	logger                *slog.Logger
	apiSettingsDiscoverer abstractions.ApiSettingsDiscoverer
	streamClassOperator   abstractions.StreamClassOperator
}

var _ abstractions.StreamCommandHandler = (*SyncronousCommandHandler)(nil)

// Provideres a new AnnotationStreamCommandHandler with the given configReader.
// This function is used to provide the handler in the dependency injection container.
func ProvideStreamCommandHandler(logger *slog.Logger,
	apiSettingsDiscoverer abstractions.ApiSettingsDiscoverer,
	streamClassOperator abstractions.StreamClassOperator) (abstractions.StreamCommandHandler, error) {

	handler := &SyncronousCommandHandler{
		logger:                logger,
		apiSettingsDiscoverer: apiSettingsDiscoverer,
		streamClassOperator:   streamClassOperator,
	}
	return handler, nil
}

func (handler *SyncronousCommandHandler) Suspend(ctx context.Context, id string) error {
	handler.logger.Info("Reading the client configuration")
	clientApiSettings, err := handler.apiSettingsDiscoverer.DiscoveryFromJobs(ctx, id, NAMESPACE)
	if err != nil {
		return fmt.Errorf("failed to discover job %s: %w", id, err)
	}
	handler.logger.Debug("Discovered client API settings", "settings", clientApiSettings)

	err = handler.streamClassOperator.Suspend(ctx, id, NAMESPACE, clientApiSettings)
	if err != nil {
		return fmt.Errorf("failed to suspend stream %s: %w", id, err)
	}

	return nil
}

func (handler *SyncronousCommandHandler) Resume(ctx context.Context, id string, streamClass string) error {
	handler.logger.Info("Resuming stream", "id", id, "streamClass", streamClass)
	clientApiSettings, err := handler.apiSettingsDiscoverer.DiscoveryFromStreamClass(ctx, streamClass, NAMESPACE)
	if err != nil {
		return fmt.Errorf("failed to discover stream class%s: %w", id, err)
	}
	handler.logger.Debug("Discovered client API settings", "settings", clientApiSettings)

	err = handler.streamClassOperator.Resume(ctx, id, NAMESPACE, clientApiSettings)
	if err != nil {
		handler.logger.Error("Failed to resume stream", "id", id, "error", err)
		return fmt.Errorf("failed to resume stream %s: %w", id, err)
	}

	return nil
}

func (handler *SyncronousCommandHandler) Backfill(ctx context.Context, id string, watch bool) error {
	// TODO: implement backfill logic
	return nil
}

func (handler *SyncronousCommandHandler) Restart(ctx context.Context, id string, wait bool) error {
	// TODO: implement delete logic
	return nil
}
