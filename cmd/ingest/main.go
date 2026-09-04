package main

import (
	"bufio"
	"context"
	"encoding/json"
	"io"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/sagar-shaw-bits/CrediScout/internal/config"
	"github.com/sagar-shaw-bits/CrediScout/internal/domain"
	"github.com/sagar-shaw-bits/CrediScout/internal/rag"
)

func main() {
	cfg := config.ParseRAGFlags()

	log := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}))

	if err := cfg.Validate(); err != nil {
		log.Error("invalid RAG config", "err", err)
		os.Exit(1)
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	if err := run(ctx, cfg, log); err != nil {
		log.Error("RAG ingest failed", "err", err)
		os.Exit(1)
	}
}

func run(ctx context.Context, cfg config.RAGConfig, log *slog.Logger) error {
	docs, err := readDocuments(cfg.InputFile)
	if err != nil {
		return err
	}
	log.Info("loaded corpus", "documents", len(docs), "file", cfg.InputFile)

	chunker := rag.NewChunker(cfg.ChunkSize, cfg.ChunkOverlap)
	var chunks []domain.Chunk
	for _, d := range docs {
		chunks = append(chunks, chunker.Split(d)...)
	}
	log.Info("chunked corpus", "chunks", len(chunks))
	if len(chunks) == 0 {
		return nil
	}

	embedder := rag.NewOllamaEmbedder(cfg.OllamaURL, cfg.EmbedModel, cfg.RequestTimeout)
	log.Info("embedding chunks", "model", cfg.EmbedModel, "batch_size", cfg.BatchSize)
	start := time.Now()

	if _, err := rag.EmbedChunks(ctx, embedder, chunks, cfg.BatchSize); err != nil {
		return err
	}
	log.Info("embedding done", "elapsed", time.Since(start).Round(time.Millisecond).String())

	store := rag.NewVectorStore()
	store.Add(chunks...)
	if err := store.Save(cfg.IndexFile); err != nil {
		return err
	}
	log.Info("index written", "file", cfg.IndexFile, "chunks", store.Len())
	return nil
}

func readDocuments(path string) ([]domain.Document, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer func() { _ = f.Close() }()

	var docs []domain.Document
	r := bufio.NewReader(f)
	for {
		line, err := r.ReadBytes('\n')
		if len(line) > 0 {
			var d domain.Document
			if jErr := json.Unmarshal(line, &d); jErr != nil {
				return nil, jErr
			}
			docs = append(docs, d)
		}
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, err
		}
	}
	return docs, nil
}
