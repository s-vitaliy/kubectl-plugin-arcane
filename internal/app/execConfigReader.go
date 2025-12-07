package app

import (
	"fmt"
	"os/exec"

	"s-vitaliy/kubectl-plugin-arcane/internal/client/api/common"

	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

var _ common.ConfigReader = (*execConfigReader)(nil)

// NewValidatedExecConfigReaderProvider returns a provider function that creates an execConfigReader
// and validates it by invoking ReadConfig before returning.
func NewValidatedExecConfigReaderProvider(command *string, arguments []string) func() (common.ConfigReader, error) {
	return func() (common.ConfigReader, error) {
		reader := ProvideExecConfigReader(command, arguments)
		_, err := reader.ReadConfig()
		if err != nil { // coverage-ignore
			return nil, fmt.Errorf("failed to validate config reader: %w", err)
		}
		return reader, nil
	}
}

// ProvideExecConfigReader provides a new instance of execConfigReader.
func ProvideExecConfigReader(command *string, arguments []string) common.ConfigReader {
	return &execConfigReader{command, arguments}
}

func NewExecConfigReader(command *string, arguments []string) (common.ConfigReader, error) { // coverage-ignore, the code is trivial
	return &execConfigReader{command, arguments}, nil
}

type execConfigReader struct {
	command   *string
	arguments []string
}

func (r *execConfigReader) ReadConfig() (*rest.Config, error) {
	// Run the command and capture output
	cmd := exec.Command(*r.command, r.arguments...)
	output, err := cmd.Output()
	if err != nil { // coverage-ignore
		return nil, fmt.Errorf("failed to execute command %s: %w", *r.command, err)
	}

	// Parse the output as kubeconfig
	return clientcmd.RESTConfigFromKubeConfig(output)
}
