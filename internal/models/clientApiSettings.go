package models

import (
	"fmt"

	"k8s.io/apimachinery/pkg/runtime/schema"
)

type ClientApiSettings struct {
	apiGroup   string
	apiVersion string
	apiPlural  string
}

func NewClientApiSettings(apiGroup, apiVersion, apiPlural string) *ClientApiSettings {
	return &ClientApiSettings{
		apiGroup:   apiGroup,
		apiVersion: apiVersion,
		apiPlural:  apiPlural,
	}
}

func (settings *ClientApiSettings) ToGroupVersionResource() schema.GroupVersionResource {
	return schema.GroupVersionResource{
		Group:    settings.apiGroup,
		Version:  settings.apiVersion,
		Resource: settings.apiPlural,
	}
}

func FromJobAnnotations(annotations map[string]interface{}) (*ClientApiSettings, error) {
	apiGroup, err := getAnnotation(annotations, "stream.arcane.sneaksanddata.com/api-group")
	if err != nil {
		return nil, err
	}
	apiVersion, err := getAnnotation(annotations, "stream.arcane.sneaksanddata.com/api-version")
	if err != nil {
		return nil, err
	}
	apiPlural, err := getAnnotation(annotations, "stream.arcane.sneaksanddata.com/api-plural-name")
	if err != nil {
		return nil, err
	}

	settings := NewClientApiSettings(
		apiGroup,
		apiVersion,
		apiPlural,
	)

	return settings, nil
}

func (s *ClientApiSettings) String() string {
	return fmt.Sprintf("ClientApiSettings(Group: %s, Version: %s, Plural: %s)", s.apiGroup, s.apiVersion, s.apiPlural)
}

func getAnnotation(annotations map[string]any, key string) (string, error) {
	value, ok := annotations[key]
	if !ok {
		return "", fmt.Errorf("missing required annotation: %s, found annotations: %v", key, annotations)
	}
	if strValue, ok := value.(string); ok {
		return strValue, nil
	}
	return "", fmt.Errorf("failed to read annotation %s: expected string, got %T", key, value)
}
