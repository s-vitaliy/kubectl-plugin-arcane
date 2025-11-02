package app

import (
	"fmt"
	"os"

	"s-vitaliy/kubectl-plugin-arcane/internal/client/api/common"

	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

func ProvideConfigReader() (common.ConfigReader, error) { // coverage-ignore, the code is trivial
	return &fileConfigReader{ConfigOverride: ""}, nil
}

type fileConfigReader struct {
	ConfigOverride string
}

func (r *fileConfigReader) ReadConfig() (*rest.Config, error) { // coverage-ignore, the code is trivial
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

func (r *fileConfigReader) readFromFile(path string) (*rest.Config, error) { // coverage-ignore, the code is trivial

	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read kube config file %s: %w", path, err)
	}
	return clientcmd.RESTConfigFromKubeConfig(data)
}
