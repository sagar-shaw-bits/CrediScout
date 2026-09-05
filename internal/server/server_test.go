package server

import (
	"context"
	"io"
	"log/slog"
	"testing"
	"testing/fstest"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"

	"github.com/sagar-shaw-bits/CrediScout/internal/server/endpoint"
)

type mockEndpoint struct {
	called bool
}

func (m *mockEndpoint) AddEndpoints(rg *gin.RouterGroup) {
	m.called = true
	rg.GET("/mock", func(c *gin.Context) {})
}

func TestStartHTTPServer(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())

	ui := fstest.MapFS{
		"index.html": {Data: []byte("<html></html>")},
	}
	log := slog.New(slog.NewTextHandler(io.Discard, nil))
	endpoints := &mockEndpoint{}
	binders := []endpoint.EndpointBinder{endpoints}

	errCh := make(chan error, 1)
	go func() {
		errCh <- StartHTTPServer(ctx, "127.0.0.1:0", ui, binders, log)
	}()

	time.Sleep(100 * time.Millisecond)
	cancel()

	select {
	case err := <-errCh:
		assert.NoError(t, err)
	case <-time.After(2 * time.Second):
		t.Fatal("server did not shut down within timeout")
	}

	assert.True(t, endpoints.called, "expected binder AddEndpoints to be invoked")
}
