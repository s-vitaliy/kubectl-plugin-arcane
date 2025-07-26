package api

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"os"

	"s-vitaliy/kubectl-plugin-arcane/internal/app"

	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/rest"
)

type HandlerContext struct {
	Logger *slog.Logger
}

var suspendAnnotation = map[string]interface{}{
	"metadata": map[string]interface{}{
		"annotations": map[string]string{
			"arcane/state": "suspended",
		},
	},
}

var NAMESPACE = "arcane"

type AnnotationStreamCommandHandler struct {
	context      *HandlerContext
	configReader app.ConfigReader
}

func NewAnnotationStreamCommandHandlerV1(configReader app.ConfigReader) *AnnotationStreamCommandHandler {
	return &AnnotationStreamCommandHandler{context: nil, configReader: configReader}
}

func NewAnnotationStreamCommandHandler(context *HandlerContext) *AnnotationStreamCommandHandler { // TODO: remove this duplicate
	configReader := app.FileConfigReader{ConfigOverride: ""}
	return &AnnotationStreamCommandHandler{context: context, configReader: &configReader}
}

// Suspend suspends the stream with the given ID.
// It returns an error if the operation fails.
func (h *AnnotationStreamCommandHandler) Suspend(id string) error {
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))

	logger.Info("Reading the client configuration")
	config, err := h.configReader.ReadConfig()
	if err != nil {
		return fmt.Errorf("failed to read config: %w", err)
	}
	client, err := h.buildDynamicClient(config)
	if err != nil {
		return fmt.Errorf("failed to build dynamic client: %w", err)
	}

	clientApiSettings, err := h.discoveryFromJobs(client, id, NAMESPACE)
	if err != nil {
		return fmt.Errorf("failed to discover job %s: %w", id, err)
	}
	logger.Debug("Discovered client API settings", "settings", clientApiSettings)

	patchBytes, err := json.Marshal(suspendAnnotation)
	if err != nil {
		return fmt.Errorf("failed to marshal suspend annotation: %w", err)
	}
	dynamicClient := client.Resource(schema.GroupVersionResource{
		Group:    clientApiSettings.apiGroup,
		Version:  clientApiSettings.apiVersion,
		Resource: clientApiSettings.apiPlural,
	}).Namespace(NAMESPACE)

	_, err = dynamicClient.Patch(context.TODO(),
		id,
		types.MergePatchType,
		patchBytes,
		v1.PatchOptions{})

	if err != nil {
		logger.Error("Failed to suspend stream", "id", id, "error", err)
		return fmt.Errorf("failed to suspend stream %s: %w", id, err)
	}

	return nil
}

func (h *AnnotationStreamCommandHandler) Backfill(id string, watch bool) error {
	// TODO: implement resume logic
	return nil
}

func (h *AnnotationStreamCommandHandler) Restart(id string, wait bool) error {
	// TODO: implement delete logic
	return nil
}

func (h *AnnotationStreamCommandHandler) buildDynamicClient(config *rest.Config) (dynamic.Interface, error) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	clientset, err := dynamic.NewForConfig(config)
	if err != nil {
		return nil, err
	}
	logger.Debug("Created dynamic client", "clientset", clientset)
	return clientset, nil
}

func (h *AnnotationStreamCommandHandler) discoveryFromJobs(dynamicInterface dynamic.Interface, name string, namespace string) (*ClientApiSettings, error) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	resourceRef := schema.GroupVersionResource{
		Group:    "batch",
		Version:  "v1",
		Resource: "jobs",
	}
	dynamicClient := dynamicInterface.Resource(resourceRef).Namespace(namespace)
	jobValue, err := dynamicClient.Get(context.TODO(), name, v1.GetOptions{})
	if err != nil {
		logger.Error("Failed to get job", "namespace", "name", namespace, name, "error", err)
		return nil, fmt.Errorf("failed to get job %s: %w", name, err)
	}
	metadata, ok := jobValue.Object["metadata"].(map[string]interface{})
	if !ok {
		logger.Error("Failed to get metadata from job", "namespace", "name", namespace, name)
		return nil, fmt.Errorf("failed to get metadata from job %s", name)
	}
	annotations, ok := metadata["annotations"].(map[string]interface{})
	if !ok {
		logger.Error("Failed to get annotations from job metadata", "namespace", "name", namespace, name)
		return nil, fmt.Errorf("failed to get annotations from job %s metadata", name)
	}
	logger.Debug("Annotations from job", "namespace", "name", namespace, name, "annotations", annotations)
	return ReadAnnotations(annotations)
}
