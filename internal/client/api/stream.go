package api

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"os"

	"s-vitaliy/kubectl-plugin-arcane/internal/app"

	"s-vitaliy/kubectl-plugin-arcane/internal/abstractions"

	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/rest"
)

var suspendAnnotation = map[string]interface{}{
	"metadata": map[string]interface{}{
		"annotations": map[string]string{
			"arcane/state": "suspended",
		},
	},
}

var resumeAnnotation = map[string]interface{}{
	"metadata": map[string]interface{}{
		"annotations": map[string]interface{}{
			"arcane/state": nil,
		},
	},
}

var NAMESPACE = "arcane"

type AnnotationStreamCommandHandler struct {
	logger                *slog.Logger
	configReader          app.ConfigReader
	apiSettingsDiscoverer abstractions.ApiSettingsDiscoverer
}

// Provideres a new AnnotationStreamCommandHandler with the given configReader.
// This function is used to provide the handler in the dependency injection container.
func ProvideStreamCommandHandler(configReader app.ConfigReader, logger *slog.Logger, apiSettingsDiscoverer abstractions.ApiSettingsDiscoverer) (abstractions.StreamCommandHandler, error) {
	handler := &AnnotationStreamCommandHandler{
		configReader:          configReader,
		logger:                logger,
		apiSettingsDiscoverer: apiSettingsDiscoverer,
	}
	return handler, nil
}

func (handler *AnnotationStreamCommandHandler) Suspend(id string) error {
	handler.logger.Info("Reading the client configuration")
	config, err := handler.configReader.ReadConfig()
	if err != nil {
		return fmt.Errorf("failed to read config: %w", err)
	}
	client, err := handler.buildDynamicClient(config)
	if err != nil {
		return fmt.Errorf("failed to build dynamic client: %w", err)
	}

	clientApiSettings, err := handler.apiSettingsDiscoverer.DiscoveryFromJobs(context.TODO(), client, id, NAMESPACE)
	if err != nil {
		return fmt.Errorf("failed to discover job %s: %w", id, err)
	}
	handler.logger.Debug("Discovered client API settings", "settings", clientApiSettings)

	patchBytes, err := json.Marshal(suspendAnnotation)
	if err != nil {
		return fmt.Errorf("failed to marshal suspend annotation: %w", err)
	}
	dynamicClient := client.Resource(clientApiSettings.ToGroupVersionResource()).Namespace(NAMESPACE)

	_, err = dynamicClient.Patch(context.TODO(),
		id,
		types.MergePatchType,
		patchBytes,
		v1.PatchOptions{})

	if err != nil {
		handler.logger.Error("Failed to suspend stream", "id", id, "error", err)
		return fmt.Errorf("failed to suspend stream %s: %w", id, err)
	}

	return nil
}

func (handler *AnnotationStreamCommandHandler) Resume(id string, streamClass string) error {
	handler.logger.Info("Resuming stream", "id", id, "streamClass", streamClass)
	config, err := handler.configReader.ReadConfig()
	if err != nil {
		return fmt.Errorf("failed to read config: %w", err)
	}
	client, err := handler.buildDynamicClient(config)
	if err != nil {
		return fmt.Errorf("failed to build dynamic client: %w", err)
	}

	clientApiSettings, err := handler.apiSettingsDiscoverer.DiscoveryFromStreamClass(context.TODO(), client, streamClass, NAMESPACE)
	if err != nil {
		return fmt.Errorf("failed to discover stream class%s: %w", id, err)
	}
	handler.logger.Debug("Discovered client API settings", "settings", clientApiSettings)

	dynamicClient := client.Resource(clientApiSettings.ToGroupVersionResource()).Namespace(NAMESPACE)

	patchBytes, err := json.Marshal(resumeAnnotation)
	if err != nil {
		return fmt.Errorf("failed to marshal resume annotation: %w", err)
	}
	_, err = dynamicClient.Patch(context.TODO(),
		id,
		types.MergePatchType,
		patchBytes,
		v1.PatchOptions{})

	if err != nil {
		handler.logger.Error("Failed to resume stream", "id", id, "error", err)
		return fmt.Errorf("failed to resume stream %s: %w", id, err)
	}
	return nil
}

func (handler *AnnotationStreamCommandHandler) Backfill(id string, watch bool) error {
	// TODO: implement resume logic
	return nil
}

func (handler *AnnotationStreamCommandHandler) Restart(id string, wait bool) error {
	// TODO: implement delete logic
	return nil
}

func (handler *AnnotationStreamCommandHandler) buildDynamicClient(config *rest.Config) (dynamic.Interface, error) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	clientset, err := dynamic.NewForConfig(config)
	if err != nil {
		return nil, err
	}
	logger.Debug("Created dynamic client", "clientset", clientset)
	return clientset, nil
}
