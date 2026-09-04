package rag

import (
	"testing"

	"github.com/sagar-shaw-bits/CrediScout/internal/domain"
)

func TestKeywordIndex_SurfacesExactTerm(t *testing.T) {
	// The "lifetime free" chunk is buried among generic ones; a semantically
	// weak but lexically exact query must still rank it first.
	chunks := []domain.Chunk{
		{ID: "generic1", Text: "HDFC credit cards offer rewards and cashback on spends"},
		{ID: "ltf", Text: "The Millennia card is a Lifetime Free credit card with no annual fee"},
		{ID: "generic2", Text: "Apply online for a credit card with instant approval"},
	}
	ki := NewKeywordIndex(chunks)

	hits := ki.Search("lifetime free credit card", 3)
	if len(hits) == 0 {
		t.Fatal("expected at least one hit")
	}
	if hits[0].Chunk.ID != "ltf" {
		t.Errorf("expected 'ltf' ranked first, got %q", hits[0].Chunk.ID)
	}
}

func TestKeywordIndex_EmptyQueryOrCorpus(t *testing.T) {
	ki := NewKeywordIndex([]domain.Chunk{{ID: "a", Text: "some text"}})
	if hits := ki.Search("", 5); hits != nil {
		t.Errorf("empty query should return nil, got %d hits", len(hits))
	}

	empty := NewKeywordIndex(nil)
	if hits := empty.Search("anything", 5); hits != nil {
		t.Errorf("empty corpus should return nil, got %d hits", len(hits))
	}
}

func TestKeywordIndex_RespectsK(t *testing.T) {
	chunks := []domain.Chunk{
		{ID: "a", Text: "reward reward reward"},
		{ID: "b", Text: "reward reward"},
		{ID: "c", Text: "reward"},
	}
	ki := NewKeywordIndex(chunks)
	hits := ki.Search("reward", 2)
	if len(hits) != 2 {
		t.Fatalf("expected 2 hits, got %d", len(hits))
	}
}

func TestTokenize(t *testing.T) {
	got := tokenize("Lifetime-Free, Card! (2024)")
	want := []string{"lifetime", "free", "card", "2024"}
	if len(got) != len(want) {
		t.Fatalf("got %v, want %v", got, want)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Errorf("token %d: got %q, want %q", i, got[i], want[i])
		}
	}
}
