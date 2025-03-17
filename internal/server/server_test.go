package server

import (
	"context"
	"io"
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestNew(t *testing.T) {
	t.Run("should instantiate a new server instance", func(t *testing.T) {
		mux := http.NewServeMux()
		port := "8080"
		srv := New(port, mux)
		assert.NotNil(t, srv, "expected new server instance to be non-nil")
	})
}

func TestRun(t *testing.T) {
	t.Run("should respond with expected content", func(t *testing.T) {
		mux := http.NewServeMux()
		testPath := "/test"
		expectedResponse := "Hello, Test!"
		mux.HandleFunc(testPath, func(w http.ResponseWriter, r *http.Request) {
			w.Write([]byte(expectedResponse))
		})

		port := "8090"
		addr := "localhost:" + port
		srv := New(port, mux)

		err := srv.Run()
		assert.NoError(t, err, "server should start successfully")

		resp, err := http.Get("http://" + addr + testPath)
		assert.NoError(t, err, "GET request should succeed")
		body, err := io.ReadAll(resp.Body)
		resp.Body.Close()
		assert.NoError(t, err, "should be able to read response body")
		assert.Equal(t, expectedResponse, string(body), "response should match")
	})
}

func TestShutdown(t *testing.T) {
	mux := http.NewServeMux()
	testPath := "/shutdown"
	mux.HandleFunc(testPath, func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("Shutdown Test"))
	})

	port := "8091"
	addr := "localhost:" + port
	srv := New(port, mux)

	err := srv.Run()
	assert.NoError(t, err, "server should start successfully")

	t.Run("should shutdown the server gracefully", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()
		err := srv.Shutdown(ctx)
		assert.NoError(t, err, "server shutdown should succeed")
	})

	t.Run("should refuse connection after shutdown", func(t *testing.T) {
		_, err := http.Get("http://" + addr + testPath)
		assert.Error(t, err, "expected error when connecting to shutdown server")
	})
}
