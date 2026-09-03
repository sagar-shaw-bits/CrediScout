package config

import (
	"flag"
	"fmt"
	"time"
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
		return fmt.Errorf("config: input file is required")
	}
	if c.IndexFile == "" {
		return fmt.Errorf("config: index file is required")
	}
	if c.OllamaURL == "" {
		return fmt.Errorf("config: ollama URL is required")
	}
	if c.EmbedModel == "" {
		return fmt.Errorf("config: embed model is required")
	}
	if c.ChunkSize <= 0 {
		return fmt.Errorf("config: chunk size must be positive, got %d", c.ChunkSize)
	}
	if c.ChunkOverlap < 0 || c.ChunkOverlap >= c.ChunkSize {
		return fmt.Errorf("config: chunk overlap must be in [0, chunkSize), got %d", c.ChunkOverlap)
	}
	if c.BatchSize <= 0 {
		return fmt.Errorf("config: batch size must be positive, got %d", c.BatchSize)
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
