package common

import (
	"fmt"
	"log/slog"

	"s-vitaliy/kubectl-plugin-arcane/internal/app"

	"k8s.io/client-go/dynamic"
)

func ProvideDynamicClient(configReader app.ConfigReader, logger *slog.Logger) (dynamic.Interface, error) {
	config, err := configReader.ReadConfig()
	if err != nil {
		return nil, fmt.Errorf("failed to read config: %w", err)
	}

	logger.Info("Creating dynamic client", "config", config)
	clientset, err := dynamic.NewForConfig(config)
	if err != nil {
		return nil, err
	}
	logger.Debug("Created dynamic client", "clientset", clientset)
	return clientset, nil
}
