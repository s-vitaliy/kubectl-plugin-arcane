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

func (handler *SyncronousCommandHandler) Suspend(id string) error {
	handler.logger.Info("Reading the client configuration")
	clientApiSettings, err := handler.apiSettingsDiscoverer.DiscoveryFromJobs(context.TODO(), id, NAMESPACE)
	if err != nil {
		return fmt.Errorf("failed to discover job %s: %w", id, err)
	}
	handler.logger.Debug("Discovered client API settings", "settings", clientApiSettings)

	err = handler.streamClassOperator.Suspend(context.TODO(), id, NAMESPACE, clientApiSettings)
	if err != nil {
		return fmt.Errorf("failed to suspend stream %s: %w", id, err)
	}

	return nil
}

func (handler *SyncronousCommandHandler) Resume(id string, streamClass string) error {
	handler.logger.Info("Resuming stream", "id", id, "streamClass", streamClass)
	clientApiSettings, err := handler.apiSettingsDiscoverer.DiscoveryFromStreamClass(context.TODO(), streamClass, NAMESPACE)
	if err != nil {
		return fmt.Errorf("failed to discover stream class%s: %w", id, err)
	}
	handler.logger.Debug("Discovered client API settings", "settings", clientApiSettings)

	err = handler.streamClassOperator.Resume(context.TODO(), id, NAMESPACE, clientApiSettings)
	if err != nil {
		handler.logger.Error("Failed to resume stream", "id", id, "error", err)
		return fmt.Errorf("failed to resume stream %s: %w", id, err)
	}

	return nil
}

func (handler *SyncronousCommandHandler) Backfill(id string, watch bool) error {
	// TODO: implement resume logic
	return nil
}

func (handler *SyncronousCommandHandler) Restart(id string, wait bool) error {
	// TODO: implement delete logic
	return nil
}
