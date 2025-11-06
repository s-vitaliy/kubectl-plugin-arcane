package common

import (
	"fmt"
	"log/slog"

	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/rest"
)

type ConfigReader interface {
	ReadConfig() (*rest.Config, error)
}

func ProvideDynamicClient(configReader ConfigReader, logger *slog.Logger) (dynamic.Interface, error) {
	config, err := configReader.ReadConfig()
	if err != nil {
		return nil, fmt.Errorf("failed to read config: %w", err)
	}

	logger.Info("Creating dynamic client")
	clientset, err := dynamic.NewForConfig(config)
	if err != nil {
		return nil, err
	}
	logger.Debug("Created dynamic client", "clientset", clientset)
	return clientset, nil
}
