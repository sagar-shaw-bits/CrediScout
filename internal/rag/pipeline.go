package rag

import (
	"context"
	"fmt"

	"github.com/sagar-shaw-bits/CrediScout/internal/domain"
)

func EmbedChunks(ctx context.Context, emb Embedder, chunks []domain.Chunk, batchSize int) ([]domain.Chunk, error) {
	if len(chunks) == 0 {
		return chunks, nil
	}
	if batchSize <= 0 {
		batchSize = len(chunks)
	}

	for start := 0; start < len(chunks); start += batchSize {
		if err := ctx.Err(); err != nil {
			return nil, err
		}
		end := start + batchSize
		if end > len(chunks) {
			end = len(chunks)
		}

		texts := make([]string, end-start)
		for i := start; i < end; i++ {
			texts[i-start] = chunks[i].Text
		}

		vecs, err := emb.Embed(ctx, texts)
		if err != nil {
			return nil, fmt.Errorf("embed batch [%d:%d]: %w", start, end, err)
		}
		if len(vecs) != end-start {
			return nil, fmt.Errorf("embed batch [%d:%d]: expected %d vectors, got %d", start, end, end-start, len(vecs))
		}
		for i := start; i < end; i++ {
			chunks[i].Embedding = vecs[i-start]
		}
	}
	return chunks, nil
}
