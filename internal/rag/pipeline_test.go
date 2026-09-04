package rag

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/sagar-shaw-bits/CrediScout/internal/domain"
)

func TestEmbedChunks_BatchesAndPreservesOrder(t *testing.T) {
	var seenBatches int
	emb := embedFunc(func(_ context.Context, texts []string) ([][]float32, error) {
		seenBatches++
		out := make([][]float32, len(texts))
		for i, tx := range texts {
			out[i] = []float32{float32(len(tx))}
		}
		return out, nil
	})

	chunks := []domain.Chunk{
		{ID: "0", Text: "a"},
		{ID: "1", Text: "bb"},
		{ID: "2", Text: "ccc"},
		{ID: "3", Text: "dddd"},
		{ID: "4", Text: "eeeee"},
	}

	got, err := EmbedChunks(context.Background(), emb, chunks, 2)
	require.NoError(t, err)
	assert.Equal(t, 3, seenBatches, "expected 3 batches (2+2+1)")
	for i, ch := range got {
		want := float32(len(chunks[i].Text))
		assert.Equal(t, []float32{want}, ch.Embedding, "chunk %d embedding", i)
	}
}

func TestEmbedChunks_PropagatesError(t *testing.T) {
	emb := embedFunc(func(_ context.Context, _ []string) ([][]float32, error) {
		return nil, errors.New("boom")
	})
	_, err := EmbedChunks(context.Background(), emb, []domain.Chunk{{Text: "x"}}, 1)
	assert.Error(t, err)
}

// embedFunc adapts a function to the Embedder interface for testing.
type embedFunc func(ctx context.Context, texts []string) ([][]float32, error)

func (f embedFunc) Embed(ctx context.Context, texts []string) ([][]float32, error) {
	return f(ctx, texts)
}
