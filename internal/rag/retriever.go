package rag

import (
	"context"
	"fmt"
)

type Retriever struct {
	embedder Embedder
	store    *VectorStore
	topK     int
}

func NewRetriever(embedder Embedder, store *VectorStore, topK int) *Retriever {
	if topK <= 0 {
		topK = 5
	}
	return &Retriever{embedder: embedder, store: store, topK: topK}
}

func (r *Retriever) Retrieve(ctx context.Context, query string, k int) ([]Result, error) {
	if query == "" {
		return nil, fmt.Errorf("retrieve: empty query")
	}
	if k <= 0 {
		k = r.topK
	}

	vecs, err := r.embedder.Embed(ctx, []string{query})
	if err != nil {
		return nil, fmt.Errorf("retrieve: embed query: %w", err)
	}
	if len(vecs) != 1 {
		return nil, fmt.Errorf("retrieve: expected 1 query vector, got %d", len(vecs))
	}

	return r.store.Search(vecs[0], k), nil
}
