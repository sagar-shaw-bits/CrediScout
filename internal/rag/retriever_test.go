package rag

import (
	"context"
	"testing"

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
	r := NewRetriever(staticEmbedder{vec: []float32{1, 0, 0}}, store, 5)

	hits, err := r.Retrieve(context.Background(), "anything", 1)
	if err != nil {
		t.Fatalf("Retrieve: %v", err)
	}
	if len(hits) != 1 {
		t.Fatalf("expected 1 hit, got %d", len(hits))
	}
	if hits[0].Chunk.ID != "match" {
		t.Errorf("expected 'match' first, got %q", hits[0].Chunk.ID)
	}
}

func TestRetriever_EmptyQuery(t *testing.T) {
	r := NewRetriever(staticEmbedder{vec: []float32{1}}, NewVectorStore(), 5)
	if _, err := r.Retrieve(context.Background(), "", 1); err == nil {
		t.Fatal("expected error for empty query")
	}
}

func TestRetriever_DefaultsToTopKWhenKNonPositive(t *testing.T) {
	store := NewVectorStore()
	for i := 0; i < 10; i++ {
		store.Add(domain.Chunk{ID: string(rune('a' + i)), Embedding: []float32{1, 0}})
	}
	r := NewRetriever(staticEmbedder{vec: []float32{1, 0}}, store, 3)

	hits, err := r.Retrieve(context.Background(), "q", 0) // k<=0 => default 3
	if err != nil {
		t.Fatalf("Retrieve: %v", err)
	}
	if len(hits) != 3 {
		t.Fatalf("expected 3 hits (default topK), got %d", len(hits))
	}
}
