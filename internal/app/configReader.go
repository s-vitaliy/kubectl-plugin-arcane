package app

import (
	"fmt"
	"os"

	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

type ConfigReader interface {
	ReadConfig() (*rest.Config, error)
}

func ProvideConfigReader() (ConfigReader, error) {
	return &fileConfigReader{ConfigOverride: ""}, nil
}

type fileConfigReader struct {
	ConfigOverride string
}

func (r *fileConfigReader) ReadConfig() (*rest.Config, error) {
	if r.ConfigOverride != "" {
		return r.readFromFile(r.ConfigOverride)
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return nil, fmt.Errorf("failed to get user home directory: %w", err)
	}
	path := home + "/.kube/config"
	return r.readFromFile(path)

}

func (r *fileConfigReader) readFromFile(path string) (*rest.Config, error) {

	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read kube config file %s: %w", path, err)
	}
	return clientcmd.RESTConfigFromKubeConfig(data)
}
