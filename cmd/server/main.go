package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/sagar-shaw-bits/CrediScout/internal/config"
	askhandler "github.com/sagar-shaw-bits/CrediScout/internal/handlers/rest/ask"
	"github.com/sagar-shaw-bits/CrediScout/internal/llm"
	"github.com/sagar-shaw-bits/CrediScout/internal/rag"
	"github.com/sagar-shaw-bits/CrediScout/internal/server"
	"github.com/sagar-shaw-bits/CrediScout/internal/server/endpoint"
	asksrv "github.com/sagar-shaw-bits/CrediScout/internal/services/ask"
	"github.com/sagar-shaw-bits/CrediScout/web"
)

func main() {
	cfg := config.ParseServerFlags()

	log := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}))

	if err := cfg.Validate(); err != nil {
		log.Error("invalid config", "err", err)
		os.Exit(1)
	}

	log.Info("loading index", "file", cfg.IndexFile)
	store, err := rag.LoadVectorStore(cfg.IndexFile)
	if err != nil {
		log.Error("load index", "err", err)
		os.Exit(1)
	}
	log.Info("index loaded", "chunks", store.Len())

	embedder := rag.NewOllamaEmbedder(cfg.OllamaURL, cfg.EmbedModel, cfg.RequestTimeout)
	retriever := rag.NewRetriever(embedder, store, cfg.TopK, log)
	model := llm.NewOllamaLLM(cfg.OllamaURL, cfg.ChatModel, cfg.RequestTimeout)

	askService := asksrv.NewService(retriever, model, cfg.TopK)
	askHandler := askhandler.NewAskEndpointHandler(askService, log)
	askBinder := endpoint.NewAskEndpointBinder(askHandler)

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	if err := server.StartHTTPServer(ctx, cfg.Addr, web.Assets, []endpoint.EndpointBinder{askBinder}, log); err != nil {
		log.Error("server error", "err", err)
		os.Exit(1)
	}
}
