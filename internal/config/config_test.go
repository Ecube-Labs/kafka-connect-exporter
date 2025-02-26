package config

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetEnvWithDefault(t *testing.T) {
	t.Run("Should return the value of an environment variable", func(t *testing.T) {
		os.Setenv("TEST_ENV_VAR", "test")
		t.Cleanup(func() {
			os.Unsetenv("TEST_ENV_VAR")
		})

		value := getEnvWithDefault("TEST_ENV_VAR", "default")
		assert.Equal(t, "test", value)
	})

	t.Run("Should return the default value when the environment variable is not set", func(t *testing.T) {
		value := getEnvWithDefault("TEST_ENV_VAR", "default")
		assert.Equal(t, "default", value)
	})
}
