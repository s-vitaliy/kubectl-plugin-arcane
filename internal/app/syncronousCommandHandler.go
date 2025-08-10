package app

import (
	"context"
	"fmt"
	"log/slog"
	"strings"

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

func (handler *SyncronousCommandHandler) Backfill(ctx context.Context, id string, streamClass string, watch bool) error {
	handler.logger.Info("Restarting the stream in backfill mode", "id", id, "wait", watch)
	// TODO: handle situation when stream is not running

	clientApiSettings, err := handler.apiSettingsDiscoverer.DiscoveryFromJobs(ctx, id, NAMESPACE)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			handler.logger.Warn("Job not found, probably the stream is suspended, trying to discover from stream class", "id", id)
			if streamClass == "" {
				return fmt.Errorf("stream class is required when the stream is suspended %s", id)
			}
			clientApiSettings, err = handler.apiSettingsDiscoverer.DiscoveryFromStreamClass(ctx, streamClass, NAMESPACE)
			if err != nil {
				return fmt.Errorf("failed to discover stream class %s: %w", streamClass, err)
			}
		} else {
			return fmt.Errorf("failed to discover job %s: %w", id, err)
		}
	}
	handler.logger.Debug("Discovered client API settings", "settings", clientApiSettings)
	done := make(chan error, 1)
	go func() {
		defer close(done)
		if watch {
			done <- handler.streamClassOperator.WaitForStatus(ctx, abstractions.StreamPhaseBackfill, id, NAMESPACE, clientApiSettings)
		} else {
			done <- nil
		}
	}()

	err = handler.streamClassOperator.Backfill(ctx, id, NAMESPACE, clientApiSettings)
	if err != nil {
		handler.logger.Error("Failed to backfill stream", "id", id, "error", err)
		return fmt.Errorf("failed to backfill stream %s: %w", id, err)
	}

	waitErr := <-done
	handler.logger.Info("Backfill operation completed", "id", id, "waitErr", waitErr)
	if waitErr != nil {
		handler.logger.Error("Failed to wait for stream to be backfilled", "id", id, "error", waitErr)
		return fmt.Errorf("failed to wait for stream %s to be backfilled: %w", id, waitErr)
	}

	if watch {
		handler.logger.Info("Waiting for stream to complete backfill", "id", id)
		err = handler.streamClassOperator.WaitForStatus(ctx, abstractions.StreamPhaseRunning, id, NAMESPACE, clientApiSettings)
		if err != nil {
			handler.logger.Error("Failed to wait for stream to be running", "id", id, "error", err)
			return fmt.Errorf("failed to wait for stream %s to be running: %w", id, err)
		}
	}
	return nil
}

func (handler *SyncronousCommandHandler) Restart(ctx context.Context, id string, wait bool) error {
	handler.logger.Info("Restarting stream", "id", id, "wait", wait)
	clientApiSettings, err := handler.apiSettingsDiscoverer.DiscoveryFromJobs(ctx, id, NAMESPACE)
	if err != nil {
		return fmt.Errorf("failed to discover job %s: %w", id, err)
	}
	handler.logger.Debug("Discovered client API settings", "settings", clientApiSettings)

	done := make(chan error, 1)
	go func() {
		defer close(done)
		done <- handler.streamClassOperator.WaitForStatus(ctx, abstractions.StreamPhaseSuspended, id, NAMESPACE, clientApiSettings)
	}()

	err = handler.streamClassOperator.Suspend(ctx, id, NAMESPACE, clientApiSettings)
	if err != nil {
		handler.logger.Error("Failed to suspend stream", "id", id, "error", err)
		return fmt.Errorf("failed to suspend stream %s: %w", id, err)
	}

	waitErr := <-done
	if waitErr != nil {
		handler.logger.Error("Failed to wait for stream to be suspended", "id", id, "error", waitErr)
		return fmt.Errorf("failed to wait for stream %s to be suspended: %w", id, waitErr)
	}

	err = handler.streamClassOperator.Resume(ctx, id, NAMESPACE, clientApiSettings)
	if err != nil {
		handler.logger.Error("Failed to resume stream", "id", id, "error", err)
		return fmt.Errorf("failed to resume stream %s: %w", id, err)
	}

	if wait {
		err = handler.streamClassOperator.WaitForStatus(ctx, abstractions.StreamPhaseRunning, id, NAMESPACE, clientApiSettings)
		if err != nil {
			handler.logger.Error("Failed to wait for stream to be running", "id", id, "error", err)
			return fmt.Errorf("failed to wait for stream %s to be running: %w", id, err)
		}
	}

	return nil
}
