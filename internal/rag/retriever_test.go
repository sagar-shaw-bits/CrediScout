package rag

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/sagar-shaw-bits/CrediScout/internal/domain"
)

// staticEmbedder returns a fixed query vector, letting us test retrieval
// ranking without a live Ollama server.
type staticEmbedder struct{ vec []float32 }

func (s staticEmbedder) Embed(_ context.Context, texts []string) ([][]float32, error) {
	out := make([][]float32, len(texts))
	for i := range texts {
		out[i] = s.vec
	}
	return out, nil
}

func TestRetriever_ReturnsRankedChunks(t *testing.T) {
	store := NewVectorStore()
	store.Add(
		domain.Chunk{ID: "match", URL: "u1", Text: "match", Embedding: []float32{1, 0, 0}},
		domain.Chunk{ID: "off", URL: "u2", Text: "off", Embedding: []float32{0, 1, 0}},
	)
	r := NewRetriever(staticEmbedder{vec: []float32{1, 0, 0}}, store, 5, nil)

	hits, err := r.Retrieve(context.Background(), "anything", 1)
	require.NoError(t, err)
	require.Len(t, hits, 1)
	assert.Equal(t, "match", hits[0].Chunk.ID)
}

func TestRetriever_EmptyQuery(t *testing.T) {
	r := NewRetriever(staticEmbedder{vec: []float32{1}}, NewVectorStore(), 5, nil)
	_, err := r.Retrieve(context.Background(), "", 1)
	assert.Error(t, err)
}

func TestRetriever_DefaultsToTopKWhenKNonPositive(t *testing.T) {
	store := NewVectorStore()
	for i := 0; i < 10; i++ {
		store.Add(domain.Chunk{ID: string(rune('a' + i)), Embedding: []float32{1, 0}})
	}
	r := NewRetriever(staticEmbedder{vec: []float32{1, 0}}, store, 3, nil)

	hits, err := r.Retrieve(context.Background(), "q", 0) // k<=0 => default 3
	require.NoError(t, err)
	assert.Len(t, hits, 3, "expected default topK results")
}
