package applicationError

import (
	"errors"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNew(t *testing.T) {
	t.Run("Should return a new httpError", func(t *testing.T) {
		err := New(500, "Internal Server Error", "Something went wrong.")

		assert.NotNil(t, err)
		assert.Equal(t, err.Code, 500)
		assert.Equal(t, err.Message, "Internal Server Error")
		assert.Equal(t, err.ClientMessage, "Something went wrong.")
	})

	t.Run("Empty clientMsg should fall back to message", func(t *testing.T) {
		err := New(500, "Internal Server Error", "")

		assert.NotNil(t, err)
		assert.Equal(t, err.ClientMessage, "Internal Server Error")
	})

	t.Run("Should return a new httpError with stack trace", func(t *testing.T) {
		err := New(500, "Internal Server Error", "Something went wrong.")

		assert.NotNil(t, err)
		assert.Contains(t, err.Stack, "Error: Internal Server Error")
	})
}

func httpErrorFuncOne() error {
	return New(500, "Internal Server Error", "Something went wrong.")
}

func errorFuncOne() error {
	return errors.New("Internal Server Error")
}
func unwrapError(errorType string) error {
	if errorType == "httpError" {
		err := httpErrorFuncOne()
		return UnWrap(err)
	}
	if errorType == "error" {
		err := errorFuncOne()
		return UnWrap(err)
	}

	return nil
}
func wrapError(errorType string) error {
	if errorType == "httpError" {
		err := httpErrorFuncOne()
		return Wrap(err)
	}
	if errorType == "error" {
		err := errorFuncOne()
		return Wrap(err)
	}

	return nil
}

func TestWrap(t *testing.T) {
	t.Run("Should wrap httpError and add stack trace", func(t *testing.T) {
		e := wrapError("httpError")
		err, ok := e.(*httpError)

		assert.NotNil(t, err)
		assert.True(t, ok)

		stackTraces := strings.Split(err.Stack, "\n")
		assert.Equal(t, len(stackTraces), 3)

		assert.Contains(t, stackTraces[0], "Error: Internal Server Error")
		assert.Contains(t, stackTraces[1], "httpErrorFuncOne")
		assert.Contains(t, stackTraces[2], "wrapError")
	})

	t.Run("Should Wrap error and add stack trace", func(t *testing.T) {
		e := unwrapError("error")
		err, ok := e.(*httpError)

		assert.NotNil(t, err)
		assert.True(t, ok)

		stackTraces := strings.Split(err.Stack, "\n")
		assert.Equal(t, len(stackTraces), 2)

		assert.Contains(t, stackTraces[0], "Error: Internal Server Error")
		assert.Contains(t, stackTraces[1], "wrapError")
	})
}

func TestUnWrap(t *testing.T) {
	t.Run("Should return httpError and add stack trace", func(t *testing.T) {
		e := unwrapError("httpError")
		err, ok := e.(*httpError)

		assert.NotNil(t, err)
		assert.True(t, ok)

		stackTraces := strings.Split(err.Stack, "\n")
		assert.Equal(t, len(stackTraces), 2)

		assert.Contains(t, stackTraces[0], "Error: Internal Server Error")
		assert.Contains(t, stackTraces[1], "httpErrorFuncOne")
	})

	t.Run("Should wrap error and add stack trace", func(t *testing.T) {
		e := unwrapError("error")
		err, ok := e.(*httpError)

		assert.NotNil(t, err)
		assert.True(t, ok)

		stackTraces := strings.Split(err.Stack, "\n")
		assert.Equal(t, len(stackTraces), 2)

		assert.Contains(t, stackTraces[0], "Error: Internal Server Error")
		assert.Contains(t, stackTraces[1], "unwrapError")
	})
}
