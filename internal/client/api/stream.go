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
	"k8s.io/apimachinery/pkg/runtime/schema"
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
	logger       *slog.Logger
	configReader app.ConfigReader
}

// Provideres a new AnnotationStreamCommandHandler with the given configReader.
// This function is used to provide the handler in the dependency injection container.
func ProvideStreamCommandHandler(configReader app.ConfigReader, logger *slog.Logger) (abstractions.StreamCommandHandler, error) {
	return &AnnotationStreamCommandHandler{configReader: configReader, logger: logger}, nil
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

	clientApiSettings, err := handler.discoveryFromJobs(client, id, NAMESPACE)
	if err != nil {
		return fmt.Errorf("failed to discover job %s: %w", id, err)
	}
	handler.logger.Debug("Discovered client API settings", "settings", clientApiSettings)

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

	clientApiSettings, err := handler.discoveryFromStreamClass(client, id, NAMESPACE, streamClass)
	if err != nil {
		return fmt.Errorf("failed to discover stream class%s: %w", id, err)
	}
	handler.logger.Debug("Discovered client API settings", "settings", clientApiSettings)

	dynamicClient := client.Resource(schema.GroupVersionResource{
		Group:    clientApiSettings.apiGroup,
		Version:  clientApiSettings.apiVersion,
		Resource: clientApiSettings.apiPlural,
	}).Namespace(NAMESPACE)

	streamDefinition, err := dynamicClient.Get(context.TODO(), id, v1.GetOptions{})

	if err != nil {
		return fmt.Errorf("failed to get stream definition %s: %w", id, err)
	}

	if streamDefinition.Object["status"] == nil {
		return fmt.Errorf("status field is missing in stream definition %s", id)
	}

	_, ok := streamDefinition.Object["status"].(map[string]interface{})

	if !ok {
		return fmt.Errorf("status field is not a map in stream definition %s", id)
	}

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

func (handler *AnnotationStreamCommandHandler) discoveryFromJobs(dynamicInterface dynamic.Interface, name string, namespace string) (*ClientApiSettings, error) {
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

func (handler *AnnotationStreamCommandHandler) discoveryFromStreamClass(dynamicInterface dynamic.Interface, name string, namespace string, streamClass string) (*ClientApiSettings, error) {
	resourceRef := schema.GroupVersionResource{
		Group:    "streaming.sneaksanddata.com",
		Version:  "v1beta1",
		Resource: "stream-classes",
	}
	dynamicClient := dynamicInterface.Resource(resourceRef).Namespace(namespace)
	streamClassValue, err := dynamicClient.Get(context.TODO(), streamClass, v1.GetOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to get stream class %s: %w", streamClass, err)
	}
	spec, ok := streamClassValue.Object["spec"].(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("failed to get spec from stream class %s", name)
	}
	apiGroup, ok := spec["apiGroupRef"].(string)
	if !ok {
		return nil, fmt.Errorf("failed to get apiGroup from stream class %s", name)
	}
	apiVersion, ok := spec["apiVersion"].(string)
	if !ok {
		return nil, fmt.Errorf("failed to get apiVersion from stream class %s", name)
	}
	apiPlural, ok := spec["pluralName"].(string)
	if !ok {
		return nil, fmt.Errorf("failed to get apiPlural from stream class %s", name)
	}
	settings := &ClientApiSettings{
		apiGroup:   apiGroup,
		apiVersion: apiVersion,
		apiPlural:  apiPlural,
	}
	return settings, nil
}
