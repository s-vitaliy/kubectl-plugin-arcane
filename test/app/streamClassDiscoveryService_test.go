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

var configReader common.ConfigReader

func TestDiscoveryFromStreamClass(t *testing.T) {
	container := dig.New()

	err := container.Provide(common.ProvideDynamicClient)
	assert.NoError(t, err)

	err = container.Provide(common.ProvideStreamClassDiscoveryService)
	assert.NoError(t, err)

	err = container.Provide(func() common.ConfigReader {
		return configReader
	})

	_, err = configReader.ReadConfig()
	if err != nil {
		panic(err)
	}

	assert.NoError(t, err)

	err = container.Provide(func() *slog.Logger {
		handler := slog.NewTextHandler(io.Discard, nil)
		return slog.New(handler)
	})
	assert.NoError(t, err)

	err = container.Invoke(func(service abstractions.ApiSettingsDiscoverer) {
		assert.NotNil(t, service)
		api, err := service.DiscoveryFromStreamClass(t.Context(), "arcane-stream-microsoft-sql-server", "arcane")
		assert.NoError(t, err)
		assert.NotNil(t, api)
	})

	assert.NoError(t, err)

	assert.True(t, true)
}

var cmd = flag.String("cmd", "/opt/homebrew/bin/kind get kubeconfig", "Command to get kubeconfig")

func TestMain(m *testing.M) {
	flag.Parse()
	command := strings.Split(*cmd, " ")

	var err error
	configReader, err = app.NewExecConfigReader(&command[0], command[1:])
	if err != nil {
		panic(err)
	}

	// _, err = configReader.ReadConfig()
	// if err != nil {
	// 	panic(err)
	// }

	code := m.Run()
	os.Exit(code)
}
