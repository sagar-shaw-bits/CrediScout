package config

import (
	"flag"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDefaultRAGConfigShouldSetDefaults(t *testing.T) {
	cfg := DefaultRAGConfig()

	assert.Equal(t, "data/documents.jsonl", cfg.InputFile)
	assert.Equal(t, "data/index.jsonl", cfg.IndexFile)
	assert.Equal(t, "http://localhost:11434", cfg.OllamaURL)
	assert.Equal(t, "nomic-embed-text", cfg.EmbedModel)
	assert.Equal(t, 200, cfg.ChunkSize)
	assert.Equal(t, 40, cfg.ChunkOverlap)
	assert.Equal(t, 32, cfg.BatchSize)
	assert.Equal(t, 60*time.Second, cfg.RequestTimeout)
}

func TestRAGConfigValidateWhenValidShouldReturnNil(t *testing.T) {
	assert.Nil(t, DefaultRAGConfig().Validate())
}

func TestRAGConfigValidateWhenInvalidShouldReturnError(t *testing.T) {
	tests := []struct {
		name    string
		mutate  func(*RAGConfig)
		wantErr error
	}{
		{"empty input file", func(c *RAGConfig) { c.InputFile = "" }, ErrRAGInputFile},
		{"empty index file", func(c *RAGConfig) { c.IndexFile = "" }, ErrRAGIndexFile},
		{"empty ollama url", func(c *RAGConfig) { c.OllamaURL = "" }, ErrRAGOllamaURL},
		{"empty embed model", func(c *RAGConfig) { c.EmbedModel = "" }, ErrRAGEmbedModel},
		{"non positive chunk size", func(c *RAGConfig) { c.ChunkSize = 0 }, ErrRAGChunkSize},
		{"negative chunk overlap", func(c *RAGConfig) { c.ChunkOverlap = -1 }, ErrRAGChunkOverlap},
		{"chunk overlap >= chunk size", func(c *RAGConfig) { c.ChunkOverlap = c.ChunkSize }, ErrRAGChunkOverlap},
		{"non positive batch size", func(c *RAGConfig) { c.BatchSize = 0 }, ErrRAGBatchSize},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := DefaultRAGConfig()
			tt.mutate(&cfg)

			err := cfg.Validate()
			require.Error(t, err)
			assert.ErrorIs(t, err, tt.wantErr)
		})
	}
}

func TestParseRAGFlagsWhenPresentShouldOverwriteDefaults(t *testing.T) {
	oldArgs := os.Args
	defer func() { os.Args = oldArgs }()
	flag.CommandLine = flag.NewFlagSet(os.Args[0], flag.ExitOnError)

	os.Args = []string{
		"cmd",
		"--in", "in.jsonl",
		"--out", "out.jsonl",
		"--ollama-url", "http://ollama:1234",
		"--embed-model", "test-embed",
		"--chunk-size", "111",
		"--chunk-overlap", "22",
		"--batch-size", "64",
		"--embed-timeout", "90s",
	}

	cfg := ParseRAGFlags()

	assert.Equal(t, "in.jsonl", cfg.InputFile)
	assert.Equal(t, "out.jsonl", cfg.IndexFile)
	assert.Equal(t, "http://ollama:1234", cfg.OllamaURL)
	assert.Equal(t, "test-embed", cfg.EmbedModel)
	assert.Equal(t, 111, cfg.ChunkSize)
	assert.Equal(t, 22, cfg.ChunkOverlap)
	assert.Equal(t, 64, cfg.BatchSize)
	assert.Equal(t, 90*time.Second, cfg.RequestTimeout)
}
