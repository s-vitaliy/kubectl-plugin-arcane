package app

import (
	"fmt"
	"os/exec"

	"s-vitaliy/kubectl-plugin-arcane/internal/client/api/common"

	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

var _ common.ConfigReader = (*execConfigReader)(nil)

func NewExecConfigReader(execPath *string, arguments []string) (common.ConfigReader, error) { // coverage-ignore, constructor
	return &execConfigReader{command: execPath, arguments: arguments}, nil
}

type execConfigReader struct {
	command   *string
	arguments []string
}

func (r *execConfigReader) ReadConfig() (*rest.Config, error) {
	// Run the command and capture output
	cmd := exec.Command(*r.command, r.arguments...)
	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("failed to execute command %s: %w", r.command, err)
	}

	// Parse the output as kubeconfig
	return clientcmd.RESTConfigFromKubeConfig(output)
}
