package rag

import (
	"fmt"
	"strings"

	"github.com/sagar-shaw-bits/CrediScout/internal/domain"
)

type Chunker struct {
	size    int
	overlap int
}

func NewChunker(size, overlap int) *Chunker {
	if size <= 0 {
		size = 200
	}
	if overlap < 0 || overlap >= size {
		overlap = size / 5
	}
	return &Chunker{size: size, overlap: overlap}
}

func (c *Chunker) Split(doc domain.Document) []domain.Chunk {
	words := strings.Fields(doc.Text)
	if len(words) == 0 {
		return nil
	}

	step := c.size - c.overlap
	if step <= 0 {
		step = c.size
	}

	var chunks []domain.Chunk
	idx := 0
	for start := 0; start < len(words); start += step {
		end := start + c.size
		if end > len(words) {
			end = len(words)
		}
		chunks = append(chunks, domain.Chunk{
			ID:    fmt.Sprintf("%s-%d", doc.ID, idx),
			DocID: doc.ID,
			URL:   doc.URL,
			Title: doc.Title,
			Index: idx,
			Text:  strings.Join(words[start:end], " "),
		})
		idx++
		if end == len(words) {
			break
		}
	}
	return chunks
}
