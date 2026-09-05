package config

import (
	"errors"
	"flag"
	"fmt"
	"time"
)

var (
	ErrRAGInputFile    = errors.New("Input file is required")
	ErrRAGIndexFile    = errors.New("Index file is required")
	ErrRAGOllamaURL    = errors.New("Ollama URL is required")
	ErrRAGEmbedModel   = errors.New("Embed model is required")
	ErrRAGChunkSize    = errors.New("Chunk size must be positive")
	ErrRAGChunkOverlap = errors.New("Chunk overlap must be in [0, chunkSize)")
	ErrRAGBatchSize    = errors.New("Batch size must be positive")
)

type RAGConfig struct {
	InputFile      string
	IndexFile      string
	OllamaURL      string
	EmbedModel     string
	ChunkSize      int
	ChunkOverlap   int
	BatchSize      int
	RequestTimeout time.Duration
}

func DefaultRAGConfig() RAGConfig {
	return RAGConfig{
		InputFile:      "data/documents.jsonl",
		IndexFile:      "data/index.jsonl",
		OllamaURL:      "http://localhost:11434",
		EmbedModel:     "nomic-embed-text",
		ChunkSize:      200,
		ChunkOverlap:   40,
		BatchSize:      32,
		RequestTimeout: 60 * time.Second,
	}
}

func (c RAGConfig) Validate() error {
	if c.InputFile == "" {
		return ErrRAGInputFile
	}
	if c.IndexFile == "" {
		return ErrRAGIndexFile
	}
	if c.OllamaURL == "" {
		return ErrRAGOllamaURL
	}
	if c.EmbedModel == "" {
		return ErrRAGEmbedModel
	}
	if c.ChunkSize <= 0 {
		return fmt.Errorf("%w, got %d", ErrRAGChunkSize, c.ChunkSize)
	}
	if c.ChunkOverlap < 0 || c.ChunkOverlap >= c.ChunkSize {
		return fmt.Errorf("%w, got %d", ErrRAGChunkOverlap, c.ChunkOverlap)
	}
	if c.BatchSize <= 0 {
		return fmt.Errorf("%w, got %d", ErrRAGBatchSize, c.BatchSize)
	}
	return nil
}

func ParseRAGFlags() RAGConfig {
	cfg := DefaultRAGConfig()

	flag.StringVar(&cfg.InputFile, "in", cfg.InputFile, "JSONL corpus produced by the crawler")
	flag.StringVar(&cfg.IndexFile, "out", cfg.IndexFile, "JSONL vector index to write")
	flag.StringVar(&cfg.OllamaURL, "ollama-url", cfg.OllamaURL, "base URL of the Ollama server")
	flag.StringVar(&cfg.EmbedModel, "embed-model", cfg.EmbedModel, "Ollama embedding model")
	flag.IntVar(&cfg.ChunkSize, "chunk-size", cfg.ChunkSize, "target words per chunk")
	flag.IntVar(&cfg.ChunkOverlap, "chunk-overlap", cfg.ChunkOverlap, "words shared between consecutive chunks")
	flag.IntVar(&cfg.BatchSize, "batch-size", cfg.BatchSize, "chunks embedded per Ollama request")
	flag.DurationVar(&cfg.RequestTimeout, "embed-timeout", cfg.RequestTimeout, "per-request embedding timeout")
	flag.Parse()

	return cfg
}
