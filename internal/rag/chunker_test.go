package rag

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/sagar-shaw-bits/CrediScout/internal/domain"
)

func TestChunker_SplitsWithOverlap(t *testing.T) {
	// 10 words, size 4, overlap 1 => step 3 => windows start at 0,3,6; the
	// window at 6 reaches the end, so there are 3 chunks.
	words := strings.Fields("w0 w1 w2 w3 w4 w5 w6 w7 w8 w9")
	doc := domain.Document{ID: "doc1", URL: "u", Title: "t", Text: strings.Join(words, " ")}

	chunks := NewChunker(4, 1).Split(doc)

	require.Len(t, chunks, 3)
	assert.Equal(t, "w0 w1 w2 w3", chunks[0].Text)
	// Overlap: chunk1 starts at w3 (shared with chunk0's last word).
	assert.Equal(t, "w3 w4 w5 w6", chunks[1].Text)
	assert.Equal(t, "w6 w7 w8 w9", chunks[2].Text)
	for i, ch := range chunks {
		assert.Equal(t, i, ch.Index, "chunk %d index", i)
		assert.Equal(t, "doc1", ch.DocID, "chunk %d DocID", i)
		assert.Equal(t, "doc1-"+string(rune('0'+i)), ch.ID, "chunk %d ID", i)
	}
}

func TestChunker_EmptyDocument(t *testing.T) {
	got := NewChunker(10, 2).Split(domain.Document{ID: "x", Text: "   "})
	assert.Nil(t, got)
}

func TestChunker_ShorterThanChunkSize(t *testing.T) {
	doc := domain.Document{ID: "d", Text: "only three words"}
	chunks := NewChunker(50, 10).Split(doc)
	require.Len(t, chunks, 1)
	assert.Equal(t, "only three words", chunks[0].Text)
}
