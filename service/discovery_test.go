package main

import (
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/jamesread/starapp/service/internal/config"
	"github.com/jamesread/starapp/service/internal/server"
	"github.com/jamesread/starapp/service/internal/store"
)

func TestDiscoveryEndpoints(t *testing.T) {
	st := store.OpenMemory()
	svc := server.New(&config.Config{}, st, nil, nil)

	mux := http.NewServeMux()
	path, handler := svc.Handler(nil)
	mux.Handle(path, handler)
	mux.HandleFunc("/openapi", serveOpenAPI)
	mux.HandleFunc("/llms.txt", serveLLMsTxt)

	ts := httptest.NewServer(mux)
	t.Cleanup(ts.Close)

	t.Run("llms.txt", func(t *testing.T) {
		req, err := http.NewRequestWithContext(t.Context(), http.MethodGet, ts.URL+"/llms.txt", nil)
		require.NoError(t, err)
		res, err := http.DefaultClient.Do(req)
		require.NoError(t, err)
		defer func() { require.NoError(t, res.Body.Close()) }()
		require.Equal(t, http.StatusOK, res.StatusCode)
		require.Contains(t, res.Header.Get("Content-Type"), "text/plain")
		_, _ = io.Copy(io.Discard, res.Body)
	})

	t.Run("openapi", func(t *testing.T) {
		req, err := http.NewRequestWithContext(t.Context(), http.MethodGet, ts.URL+"/openapi", nil)
		require.NoError(t, err)
		res, err := http.DefaultClient.Do(req)
		require.NoError(t, err)
		defer func() { require.NoError(t, res.Body.Close()) }()
		require.Equal(t, http.StatusOK, res.StatusCode)
		require.Contains(t, res.Header.Get("Content-Type"), "application/json")
		require.NotEmpty(t, openAPISpec)
		_, _ = io.Copy(io.Discard, res.Body)
	})
}
