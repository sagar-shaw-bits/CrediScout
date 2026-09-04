package rag

import (
	"context"
	"fmt"
	"log/slog"
	"sort"

	"github.com/sagar-shaw-bits/CrediScout/internal/domain"
)

const (
	candidateMultiplier = 10
	minCandidates       = 50
	rrfK                = 60.0
)

type Retriever struct {
	embedder   Embedder
	store      *VectorStore
	keyword    *KeywordIndex
	topK       int
	candidateK int
	log        *slog.Logger
}

func NewRetriever(embedder Embedder, store *VectorStore, topK int, log *slog.Logger) *Retriever {
	if topK <= 0 {
		topK = 5
	}
	if log == nil {
		log = slog.Default()
	}
	candidateK := topK * candidateMultiplier
	if candidateK < minCandidates {
		candidateK = minCandidates
	}
	return &Retriever{
		embedder:   embedder,
		store:      store,
		keyword:    NewKeywordIndex(store.snapshotChunks()),
		topK:       topK,
		candidateK: candidateK,
		log:        log,
	}
}

func (h *Retriever) Retrieve(ctx context.Context, query string, k int) ([]Result, error) {
	if query == "" {
		return nil, fmt.Errorf("retrieve: empty query")
	}
	if k <= 0 {
		k = h.topK
	}

	lexical := h.keyword.Search(query, h.candidateK)

	var (
		dense []Result
		qvec  []float32
	)
	vecs, err := h.embedder.Embed(ctx, []string{query})
	switch {
	case err != nil:
		h.log.Warn("retrieval: embed failed, falling back to lexical-only", "err", err.Error())
	case len(vecs) != 1:
		h.log.Warn("retrieval: unexpected query vector count, falling back to lexical-only", "got", len(vecs))
	default:
		qvec = vecs[0]
		dense = h.store.Search(qvec, h.candidateK)
	}

	fused := fuseRRF([][]Result{dense, lexical}, rrfK)
	if len(fused) > k {
		fused = fused[:k]
	}

	if qvec != nil {
		for i := range fused {
			fused[i].Score = cosine(qvec, fused[i].Chunk.Embedding)
		}
	}

	var topScore float32
	if len(fused) > 0 {
		topScore = fused[0].Score
	}
	h.log.Info("retrieval",
		"dense", len(dense),
		"lexical", len(lexical),
		"fused", len(fused),
		"returned", len(fused),
		"top_score", topScore,
	)

	return fused, nil
}

func fuseRRF(lists [][]Result, k float64) []Result {
	type agg struct {
		chunk domain.Chunk
		score float64
	}
	scores := make(map[string]*agg)
	for _, list := range lists {
		for rank, r := range list {
			a, ok := scores[r.Chunk.ID]
			if !ok {
				a = &agg{chunk: r.Chunk}
				scores[r.Chunk.ID] = a
			}
			a.score += 1.0 / (k + float64(rank+1))
		}
	}

	out := make([]Result, 0, len(scores))
	for _, a := range scores {
		out = append(out, Result{Chunk: a.chunk, Score: float32(a.score)})
	}
	sort.Slice(out, func(i, j int) bool {
		if out[i].Score != out[j].Score {
			return out[i].Score > out[j].Score
		}
		return out[i].Chunk.ID < out[j].Chunk.ID
	})
	return out
}
