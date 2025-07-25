package api

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"

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
	configReader ConfigReader
}

func NewAnnotationStreamCommandHandler(context *HandlerContext) *AnnotationStreamCommandHandler {
	configReader := FileConfigReader{configOverride: ""}
	return &AnnotationStreamCommandHandler{context: context, configReader: &configReader}
}

// Suspend suspends the stream with the given ID.
// It returns an error if the operation fails.
func (h *AnnotationStreamCommandHandler) Suspend(id string) error {
	h.context.Logger.Info("Suspending stream", "id", id)
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
	h.context.Logger.Debug("Discovered client API settings", "settings", clientApiSettings)

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
	clientset, err := dynamic.NewForConfig(config)
	if err != nil {
		return nil, err
	}
	h.context.Logger.Debug("Created dynamic client", "clientset", clientset)
	return clientset, nil
}

func (h *AnnotationStreamCommandHandler) discoveryFromJobs(dynamicInterface dynamic.Interface, name string, namespace string) (*ClientApiSettings, error) {
	resourceRef := schema.GroupVersionResource{
		Group:    "batch",
		Version:  "v1",
		Resource: "jobs",
	}
	dynamicClient := dynamicInterface.Resource(resourceRef).Namespace(namespace)
	jobValue, err := dynamicClient.Get(context.TODO(), name, v1.GetOptions{})
	if err != nil {
		h.context.Logger.Error("Failed to get job", "namespace", "name", namespace, name, "error", err)
		return nil, fmt.Errorf("failed to get job %s: %w", name, err)
	}
	metadata, ok := jobValue.Object["metadata"].(map[string]interface{})
	if !ok {
		h.context.Logger.Error("Failed to get metadata from job", "namespace", "name", namespace, name)
		return nil, fmt.Errorf("failed to get metadata from job %s", name)
	}
	annotations, ok := metadata["annotations"].(map[string]interface{})
	if !ok {
		h.context.Logger.Error("Failed to get annotations from job metadata", "namespace", "name", namespace, name)
		return nil, fmt.Errorf("failed to get annotations from job %s metadata", name)
	}
	h.context.Logger.Debug("Annotations from job", "namespace", "name", namespace, name, "annotations", annotations)
	return ReadAnnotations(annotations)
}
