package rag

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

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
	require.NotEmpty(t, hits)
	assert.Equal(t, "ltf", hits[0].Chunk.ID)
}

func TestKeywordIndex_EmptyQueryOrCorpus(t *testing.T) {
	ki := NewKeywordIndex([]domain.Chunk{{ID: "a", Text: "some text"}})
	assert.Nil(t, ki.Search("", 5), "empty query should return nil")

	empty := NewKeywordIndex(nil)
	assert.Nil(t, empty.Search("anything", 5), "empty corpus should return nil")
}

func TestKeywordIndex_RespectsK(t *testing.T) {
	chunks := []domain.Chunk{
		{ID: "a", Text: "reward reward reward"},
		{ID: "b", Text: "reward reward"},
		{ID: "c", Text: "reward"},
	}
	ki := NewKeywordIndex(chunks)
	hits := ki.Search("reward", 2)
	assert.Len(t, hits, 2)
}

func TestTokenize(t *testing.T) {
	got := tokenize("Lifetime-Free, Card! (2024)")
	want := []string{"lifetime", "free", "card", "2024"}
	assert.Equal(t, want, got)
}
