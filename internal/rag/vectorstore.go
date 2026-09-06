package rag

import (
	"bufio"
	"encoding/json"
	"fmt"
	"math"
	"os"
	"path/filepath"
	"sort"
	"sync"

	"github.com/sagar-shaw-bits/CrediScout/internal/domain"
)

type Result struct {
	Chunk domain.Chunk
	Score float32
}

type VectorStore struct {
	mu     sync.RWMutex
	chunks []domain.Chunk
}

func NewVectorStore() *VectorStore {
	return &VectorStore{}
}

func (s *VectorStore) Add(chunks ...domain.Chunk) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.chunks = append(s.chunks, chunks...)
}

func (s *VectorStore) Len() int {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return len(s.chunks)
}

func (s *VectorStore) snapshotChunks() []domain.Chunk {
	s.mu.RLock()
	defer s.mu.RUnlock()
	out := make([]domain.Chunk, len(s.chunks))
	copy(out, s.chunks)
	return out
}

func (s *VectorStore) Search(query []float32, k int) []Result {
	if len(query) == 0 || k <= 0 {
		return nil
	}

	s.mu.RLock()
	defer s.mu.RUnlock()

	results := make([]Result, 0, len(s.chunks))
	for _, ch := range s.chunks {
		if len(ch.Embedding) != len(query) {
			continue
		}
		results = append(results, Result{Chunk: ch, Score: cosine(query, ch.Embedding)})
	}

	sort.Slice(results, func(i, j int) bool {
		return results[i].Score > results[j].Score
	})

	if len(results) > k {
		results = results[:k]
	}
	return results
}

func (s *VectorStore) Save(path string) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return fmt.Errorf("vectorstore: create dir: %w", err)
	}

	f, err := os.Create(path)
	if err != nil {
		return fmt.Errorf("vectorstore: create file: %w", err)
	}
	defer f.Close()

	w := bufio.NewWriter(f)
	enc := json.NewEncoder(w)

	s.mu.RLock()
	defer s.mu.RUnlock()
	for _, ch := range s.chunks {
		if err := enc.Encode(ch); err != nil {
			return fmt.Errorf("vectorstore: encode chunk %s: %w", ch.ID, err)
		}
	}
	if err := w.Flush(); err != nil {
		return fmt.Errorf("vectorstore: flush: %w", err)
	}
	return nil
}

func LoadVectorStore(path string) (*VectorStore, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("vectorstore: open index: %w", err)
	}
	defer f.Close()

	store := NewVectorStore()
	r := bufio.NewReader(f)
	for {
		line, err := r.ReadBytes('\n')
		if len(line) > 0 {
			var ch domain.Chunk
			if jErr := json.Unmarshal(line, &ch); jErr != nil {
				return nil, fmt.Errorf("vectorstore: decode chunk: %w", jErr)
			}
			store.chunks = append(store.chunks, ch)
		}
		if err != nil {
			break
		}
	}
	return store, nil
}

func cosine(a, b []float32) float32 {
	var dot, na, nb float64
	for i := range a {
		x, y := float64(a[i]), float64(b[i])
		dot += x * y
		na += x * x
		nb += y * y
	}
	if na == 0 || nb == 0 {
		return 0
	}
	return float32(dot / (math.Sqrt(na) * math.Sqrt(nb)))
}
