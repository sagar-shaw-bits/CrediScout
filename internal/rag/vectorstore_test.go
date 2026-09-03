package rag

import (
	"context"
	"os"
	"path/filepath"
	"testing"

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
	if len(got) != 2 {
		t.Fatalf("expected 2 results, got %d", len(got))
	}
	if got[0].Chunk.ID != "a" {
		t.Errorf("expected 'a' first, got %q", got[0].Chunk.ID)
	}
	if got[1].Chunk.ID != "c" {
		t.Errorf("expected 'c' second, got %q", got[1].Chunk.ID)
	}
	if got[0].Score <= got[1].Score {
		t.Errorf("scores not descending: %v then %v", got[0].Score, got[1].Score)
	}
}

func TestVectorStore_SearchSkipsDimensionMismatch(t *testing.T) {
	store := NewVectorStore()
	store.Add(
		domain.Chunk{ID: "ok", Embedding: []float32{1, 0}},
		domain.Chunk{ID: "bad", Embedding: []float32{1, 0, 0}}, // wrong dim
	)
	got := store.Search([]float32{1, 0}, 5)
	if len(got) != 1 || got[0].Chunk.ID != "ok" {
		t.Fatalf("expected only 'ok', got %+v", got)
	}
}

func TestVectorStore_SaveAndLoadRoundTrip(t *testing.T) {
	store := NewVectorStore()
	store.Add(
		domain.Chunk{ID: "a", DocID: "d", URL: "u", Title: "t", Index: 0, Text: "hello", Embedding: []float32{1, 2, 3}},
		domain.Chunk{ID: "b", DocID: "d", URL: "u", Title: "t", Index: 1, Text: "world", Embedding: []float32{4, 5, 6}},
	)

	path := filepath.Join(t.TempDir(), "index.jsonl")
	if err := store.Save(path); err != nil {
		t.Fatalf("Save: %v", err)
	}
	if _, err := os.Stat(path); err != nil {
		t.Fatalf("index file missing: %v", err)
	}

	loaded, err := LoadVectorStore(path)
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if loaded.Len() != 2 {
		t.Fatalf("expected 2 chunks, got %d", loaded.Len())
	}
	got := loaded.Search([]float32{4, 5, 6}, 1)
	if len(got) != 1 || got[0].Chunk.ID != "b" || got[0].Chunk.Text != "world" {
		t.Fatalf("round-trip lost data: %+v", got)
	}
}

func TestFakeEmbedder_UsableAsEmbedder(t *testing.T) {
	// Compile-time + behavioural check that a fake satisfies Embedder, mirroring
	// how OllamaEmbedder is swapped out in tests.
	var e Embedder = fakeEmbedder{vectors: map[string][]float32{"x": {1, 0}}}
	vecs, err := e.Embed(context.Background(), []string{"x", "unknown"})
	if err != nil {
		t.Fatalf("Embed: %v", err)
	}
	if len(vecs) != 2 {
		t.Fatalf("expected 2 vectors, got %d", len(vecs))
	}
}
