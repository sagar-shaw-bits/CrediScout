package rag

import (
	"strings"
	"testing"

	"github.com/sagar-shaw-bits/CrediScout/internal/domain"
)

func TestChunker_SplitsWithOverlap(t *testing.T) {
	// 10 words, size 4, overlap 1 => step 3 => windows start at 0,3,6; the
	// window at 6 reaches the end, so there are 3 chunks.
	words := strings.Fields("w0 w1 w2 w3 w4 w5 w6 w7 w8 w9")
	doc := domain.Document{ID: "doc1", URL: "u", Title: "t", Text: strings.Join(words, " ")}

	chunks := NewChunker(4, 1).Split(doc)

	if len(chunks) != 3 {
		t.Fatalf("expected 3 chunks, got %d", len(chunks))
	}
	if chunks[0].Text != "w0 w1 w2 w3" {
		t.Errorf("chunk0 = %q", chunks[0].Text)
	}
	// Overlap: chunk1 starts at w3 (shared with chunk0's last word).
	if chunks[1].Text != "w3 w4 w5 w6" {
		t.Errorf("chunk1 = %q", chunks[1].Text)
	}
	if chunks[2].Text != "w6 w7 w8 w9" {
		t.Errorf("last chunk = %q", chunks[2].Text)
	}
	for i, ch := range chunks {
		if ch.Index != i {
			t.Errorf("chunk %d has Index %d", i, ch.Index)
		}
		if ch.DocID != "doc1" {
			t.Errorf("chunk %d lost DocID: %q", i, ch.DocID)
		}
		if ch.ID != "doc1-"+string(rune('0'+i)) {
			t.Errorf("chunk %d ID = %q", i, ch.ID)
		}
	}
}

func TestChunker_EmptyDocument(t *testing.T) {
	if got := NewChunker(10, 2).Split(domain.Document{ID: "x", Text: "   "}); got != nil {
		t.Fatalf("expected nil for empty text, got %v", got)
	}
}

func TestChunker_ShorterThanChunkSize(t *testing.T) {
	doc := domain.Document{ID: "d", Text: "only three words"}
	chunks := NewChunker(50, 10).Split(doc)
	if len(chunks) != 1 {
		t.Fatalf("expected 1 chunk, got %d", len(chunks))
	}
	if chunks[0].Text != "only three words" {
		t.Errorf("chunk = %q", chunks[0].Text)
	}
}
