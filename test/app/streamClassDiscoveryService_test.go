package test_app

import (
	"flag"
	"io"
	"log/slog"
	"os"
	"s-vitaliy/kubectl-plugin-arcane/internal/app"
	"s-vitaliy/kubectl-plugin-arcane/internal/app/abstractions"
	"s-vitaliy/kubectl-plugin-arcane/internal/client/api/common"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"go.uber.org/dig"
)

var streamConfig *common.ConfigReader

func TestDiscoveryFromStreamClass(t *testing.T) {
	container := dig.New()
	err := container.Provide(common.ProvideDynamicClient)
	assert.NoError(t, err)

	err = container.Provide(common.ProvideStreamClassDiscoveryService)
	assert.NoError(t, err)

	err = container.Provide(func() common.ConfigReader {
		return *streamConfig
	})
	assert.NoError(t, err)

	err = container.Provide(func() *slog.Logger {
		handler := slog.NewTextHandler(io.Discard, nil)
		return slog.New(handler)
	})
	assert.NoError(t, err)

	err = container.Invoke(func(service abstractions.ApiSettingsDiscoverer) {
		assert.NotNil(t, service)
	})

	assert.NoError(t, err)
}

var cmd = flag.String("cmd", "/opt/homebrew/bin/kind get kubeconfig", "Command to get kubeconfig")

func TestMain(m *testing.M) {
	flag.Parse()
	command := strings.Split(*cmd, " ")

	streamConfig, err := app.NewExecConfigReader(&command[0], command[1:])
	if err != nil {
		panic(err)
	}

	_, err = streamConfig.ReadConfig()
	if err != nil {
		panic(err)
	}

	code := m.Run()
	os.Exit(code)
}
