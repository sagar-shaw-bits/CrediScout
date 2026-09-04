package rag

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/sagar-shaw-bits/CrediScout/internal/domain"
)

// fakeEmbedder is an offline Embedder used to test the pipeline without a live
// Ollama server. It returns deterministic vectors so similarity is predictable.
type fakeEmbedder struct {
	vectors map[string][]float32
}

func (f fakeEmbedder) Embed(_ context.Context, texts []string) ([][]float32, error) {
	out := make([][]float32, len(texts))
	for i, t := range texts {
		if v, ok := f.vectors[t]; ok {
			out[i] = v
			continue
		}
		out[i] = []float32{0, 0, 0}
	}
	return out, nil
}

func TestVectorStore_SearchRanksByCosine(t *testing.T) {
	store := NewVectorStore()
	store.Add(
		domain.Chunk{ID: "a", Text: "apple", Embedding: []float32{1, 0, 0}},
		domain.Chunk{ID: "b", Text: "banana", Embedding: []float32{0, 1, 0}},
		domain.Chunk{ID: "c", Text: "almost-apple", Embedding: []float32{0.9, 0.1, 0}},
	)

	got := store.Search([]float32{1, 0, 0}, 2)
	require.Len(t, got, 2)
	assert.Equal(t, "a", got[0].Chunk.ID, "expected 'a' first")
	assert.Equal(t, "c", got[1].Chunk.ID, "expected 'c' second")
	assert.Greater(t, got[0].Score, got[1].Score, "scores must be descending")
}

func TestVectorStore_SearchSkipsDimensionMismatch(t *testing.T) {
	store := NewVectorStore()
	store.Add(
		domain.Chunk{ID: "ok", Embedding: []float32{1, 0}},
		domain.Chunk{ID: "bad", Embedding: []float32{1, 0, 0}}, // wrong dim
	)
	got := store.Search([]float32{1, 0}, 5)
	require.Len(t, got, 1)
	assert.Equal(t, "ok", got[0].Chunk.ID)
}

func TestVectorStore_SaveAndLoadRoundTrip(t *testing.T) {
	store := NewVectorStore()
	store.Add(
		domain.Chunk{ID: "a", DocID: "d", URL: "u", Title: "t", Index: 0, Text: "hello", Embedding: []float32{1, 2, 3}},
		domain.Chunk{ID: "b", DocID: "d", URL: "u", Title: "t", Index: 1, Text: "world", Embedding: []float32{4, 5, 6}},
	)

	path := filepath.Join(t.TempDir(), "index.jsonl")
	require.NoError(t, store.Save(path))
	_, err := os.Stat(path)
	require.NoError(t, err, "index file must exist")

	loaded, err := LoadVectorStore(path)
	require.NoError(t, err)
	require.Equal(t, 2, loaded.Len())
	got := loaded.Search([]float32{4, 5, 6}, 1)
	require.Len(t, got, 1)
	assert.Equal(t, "b", got[0].Chunk.ID)
	assert.Equal(t, "world", got[0].Chunk.Text)
}

func TestFakeEmbedder_UsableAsEmbedder(t *testing.T) {
	// Compile-time + behavioural check that a fake satisfies Embedder, mirroring
	// how OllamaEmbedder is swapped out in tests.
	var e Embedder = fakeEmbedder{vectors: map[string][]float32{"x": {1, 0}}}
	vecs, err := e.Embed(context.Background(), []string{"x", "unknown"})
	require.NoError(t, err)
	assert.Len(t, vecs, 2)
}
