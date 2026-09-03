package server

import (
	"context"
	"io/fs"
	"log/slog"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/sagar-shaw-bits/CrediScout/internal/server/endpoint"
)

func StartHTTPServer(ctx context.Context, addr string, ui fs.FS, binders []endpoint.EndpointBinder, log *slog.Logger) error {
	gin.SetMode(gin.ReleaseMode)

	router := gin.New()
	router.Use(gin.Recovery(), requestLogger(log))

	router.GET("/healthz", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	fileServer := http.FileServer(http.FS(ui))
	router.GET("/", gin.WrapH(fileServer))
	router.GET("/app.js", gin.WrapH(fileServer))
	router.GET("/style.css", gin.WrapH(fileServer))

	api := router.Group("/api")
	for _, b := range binders {
		b.AddEndpoints(api)
	}

	srv := &http.Server{
		Addr:              addr,
		Handler:           router,
		ReadHeaderTimeout: 10 * time.Second,
	}

	errCh := make(chan error, 1)
	go func() {
		log.Info("server listening", "addr", addr)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			errCh <- err
		}
	}()

	select {
	case err := <-errCh:
		return err
	case <-ctx.Done():
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		log.Info("shutting down server")
		return srv.Shutdown(shutdownCtx)
	}
}

func requestLogger(log *slog.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		c.Next()
		log.Info("request",
			"method", c.Request.Method,
			"path", c.Request.URL.Path,
			"status", c.Writer.Status(),
			"elapsed", time.Since(start).Round(time.Millisecond).String(),
		)
	}
}
