package config

import (
	"errors"
	"flag"
	"fmt"
	"os"
	"strconv"
	"time"
)

var (
	ErrServerAddr       = errors.New("Server address is required")
	ErrServerIndexFile  = errors.New("Index file is required")
	ErrServerOllamaURL  = errors.New("Ollama URL is required")
	ErrServerEmbedModel = errors.New("Embed model is required")
	ErrServerChatModel  = errors.New("Chat model is required")
	ErrServerTopK       = errors.New("Top-k must be positive")
)

type ServerConfig struct {
	Addr           string
	IndexFile      string
	OllamaURL      string
	EmbedModel     string
	ChatModel      string
	TopK           int
	RequestTimeout time.Duration
}

func DefaultServerConfig() ServerConfig {
	return ServerConfig{
		Addr:           ":8080",
		IndexFile:      "data/index.jsonl",
		OllamaURL:      "http://localhost:11434",
		EmbedModel:     "nomic-embed-text",
		ChatModel:      "llama3.2",
		TopK:           5,
		RequestTimeout: 120 * time.Second,
	}
}

func (c ServerConfig) Validate() error {
	if c.Addr == "" {
		return ErrServerAddr
	}
	if c.IndexFile == "" {
		return ErrServerIndexFile
	}
	if c.OllamaURL == "" {
		return ErrServerOllamaURL
	}
	if c.EmbedModel == "" {
		return ErrServerEmbedModel
	}
	if c.ChatModel == "" {
		return ErrServerChatModel
	}
	if c.TopK <= 0 {
		return fmt.Errorf("%w, got %d", ErrServerTopK, c.TopK)
	}
	return nil
}

func ParseServerFlags() ServerConfig {
	cfg := DefaultServerConfig()
	cfg.Addr = envStr("SERVER_ADDR", cfg.Addr)
	cfg.IndexFile = envStr("INDEX_FILE", cfg.IndexFile)
	cfg.OllamaURL = envStr("OLLAMA_URL", cfg.OllamaURL)
	cfg.EmbedModel = envStr("EMBED_MODEL", cfg.EmbedModel)
	cfg.ChatModel = envStr("CHAT_MODEL", cfg.ChatModel)
	cfg.TopK = envInt("TOP_K", cfg.TopK)

	flag.StringVar(&cfg.Addr, "addr", cfg.Addr, "host:port to listen on")
	flag.StringVar(&cfg.IndexFile, "index", cfg.IndexFile, "JSONL vector index to load")
	flag.StringVar(&cfg.OllamaURL, "ollama-url", cfg.OllamaURL, "base URL of the Ollama server")
	flag.StringVar(&cfg.EmbedModel, "embed-model", cfg.EmbedModel, "Ollama embedding model (must match ingest)")
	flag.StringVar(&cfg.ChatModel, "chat-model", cfg.ChatModel, "Ollama chat model for answering")
	flag.IntVar(&cfg.TopK, "top-k", cfg.TopK, "chunks retrieved as context per question")
	flag.DurationVar(&cfg.RequestTimeout, "timeout", cfg.RequestTimeout, "per-request Ollama timeout")
	flag.Parse()

	return cfg
}

func envStr(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func envInt(key string, fallback int) int {
	if v := os.Getenv(key); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			return n
		}
	}
	return fallback
}
