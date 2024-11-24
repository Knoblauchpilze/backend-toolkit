package config

import (
	"fmt"
	"os"
	"testing"

	"github.com/google/uuid"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type sampleServerConfig struct {
	Port uint16
}

type sampleConfig struct {
	Server sampleServerConfig
}

// https://stackoverflow.com/questions/23729790/how-can-i-do-test-setup-using-the-testing-package-in-go
func TestMain(m *testing.M) {
	err := os.MkdirAll("configs", 0777)
	if err != nil {
		os.Exit(1)
	}

	defer func() {
		os.RemoveAll("configs")
	}()

	m.Run()
}

func TestUnit_Load(t *testing.T) {
	configName := writeSampleConfigFile(t)

	in := sampleConfig{
		Server: sampleServerConfig{
			Port: 22,
		},
	}

	actual, err := Load(configName, in)
	assert.Nil(t, err)
	assert.Equal(t, uint16(20), actual.Server.Port)
}

func TestUnit_Load_WhenFileDoesNotExist_ExpectError(t *testing.T) {
	configName := writeSampleConfigFile(t)

	in := sampleConfig{
		Server: sampleServerConfig{
			Port: 22,
		},
	}

	otherConfigName := configName + "-suffix"

	_, err := Load(otherConfigName, in)
	_, ok := err.(viper.ConfigFileNotFoundError)
	assert.True(t, ok)
}

func TestUnit_Load_WhenFileDoesNotExist_ExpectDefaultConfigReturned(t *testing.T) {
	configName := writeSampleConfigFile(t)

	in := sampleConfig{
		Server: sampleServerConfig{
			Port: 22,
		},
	}

	otherConfigName := configName + "-suffix"

	actual, err := Load(otherConfigName, in)
	require.NotNil(t, err)
	assert.Equal(t, in.Server.Port, actual.Server.Port)
}

func TestUnit_Load_WhenEnvironmentVariableExists_ExpectTakesPrecedenceOverConfig(t *testing.T) {
	configName := writeSampleConfigFile(t)

	in := sampleConfig{
		Server: sampleServerConfig{
			Port: 22,
		},
	}

	// https://stackoverflow.com/questions/68686006/set-particular-environment-variables-during-execution-of-a-test-suite
	t.Setenv("ENV_SERVER_PORT", "26")

	actual, err := Load(configName, in)
	assert.Nil(t, err)
	assert.Equal(t, uint16(26), actual.Server.Port)
}

func TestUnit_Load_WhenConfigDoesNotExistInFileButEnvironmentVariableDoes_ExpectValueNotOverridden(t *testing.T) {
	// https://github.com/spf13/viper/issues/1797
	// This might change with the next release
	configName := writeConfigFile(t, nil)

	in := sampleConfig{
		Server: sampleServerConfig{
			Port: 22,
		},
	}

	t.Setenv("ENV_SERVER_PORT", "26")

	actual, err := Load(configName, in)
	assert.Nil(t, err)
	assert.Equal(t, uint16(22), actual.Server.Port)
}

func writeSampleConfigFile(t *testing.T) string {
	// https://stackoverflow.com/questions/19975954/a-yaml-file-cannot-contain-tabs-as-indentation
	sampleYaml := "Server:\n  Port: 20\n"
	return writeConfigFile(t, []byte(sampleYaml))
}

func writeConfigFile(t *testing.T, content []byte) string {
	configName := fmt.Sprintf("config-%s", uuid.New())
	configFileName := fmt.Sprintf("configs/%s.yml", configName)
	err := os.WriteFile(configFileName, content, 0666)
	require.Nil(t, err)

	return configName
}
