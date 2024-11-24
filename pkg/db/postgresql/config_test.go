package postgresql

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestUnit_Config_ExpectCorrectConnectionString(t *testing.T) {
	config := Config{
		Host:           "my-host",
		Port:           8976,
		Database:       "my-database",
		User:           "super-user",
		Password:       "strong-password?",
		ConnectTimeout: 2 * time.Hour,
	}

	actual := config.ToConnectionString()

	expected := "postgresql://super-user:strong-password%3F@my-host:8976/my-database?connect_timeout=7200"
	assert.Equal(t, expected, actual)
}

func TestUnit_Config_ConfigForLocalhost(t *testing.T) {
	config := NewConfigForLocalhost("my-local-database", "my-local-user", "my-local-password")

	actual := config.ToConnectionString()

	expected := "postgresql://my-local-user:my-local-password@localhost:5432/my-local-database?connect_timeout=5"
	assert.Equal(t, expected, actual)
}

func TestUnit_Config_ConfigForDockerContainer(t *testing.T) {
	config := NewConfigForDockerContainer("my-remote-database", "my-remote-user", "my-remote-password")

	actual := config.ToConnectionString()

	expected := "postgresql://my-remote-user:my-remote-password@172.17.0.1:5432/my-remote-database?connect_timeout=5"
	assert.Equal(t, expected, actual)
}
