package api

import "fmt"

type ClientApiSettings struct {
	apiGroup   string
	apiVersion string
	apiPlural  string
}

func ReadAnnotations(annotations map[string]interface{}) (*ClientApiSettings, error) {
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

	setttings := &ClientApiSettings{
		apiGroup:   apiGroup,
		apiVersion: apiVersion,
		apiPlural:  apiPlural,
	}

	return setttings, nil
}

func getAnnotation(annotations map[string]interface{}, key string) (string, error) {
	value, ok := annotations[key]
	if !ok {
		return "", fmt.Errorf("missing required annotation: %s, found annotations: %v", key, annotations)
	}
	if strValue, ok := value.(string); ok {
		return strValue, nil
	}
	return "", fmt.Errorf("failed to read annotation %s: expected string, got %T", key, value)
}
