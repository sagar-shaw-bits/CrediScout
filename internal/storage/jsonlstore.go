package storage

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"

	"github.com/sagar-shaw-bits/CrediScout/internal/domain"
)

type JSONLStore struct {
	mu     sync.Mutex
	file   *os.File
	writer *bufio.Writer
}

func NewJSONLStore(path string) (*JSONLStore, error) {
	if dir := filepath.Dir(path); dir != "" {
		if err := os.MkdirAll(dir, 0o755); err != nil {
			return nil, fmt.Errorf("storage: create dir %q: %w", dir, err)
		}
	}
	f, err := os.Create(path)
	if err != nil {
		return nil, fmt.Errorf("storage: create %q: %w", path, err)
	}
	return &JSONLStore{file: f, writer: bufio.NewWriter(f)}, nil
}

func (s *JSONLStore) Save(ctx context.Context, doc domain.Document) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if doc.ID == "" {
		return fmt.Errorf("storage: document has empty ID")
	}

	data, err := json.Marshal(doc)
	if err != nil {
		return fmt.Errorf("storage: marshal doc %s: %w", doc.ID, err)
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	if _, err := s.writer.Write(data); err != nil {
		return fmt.Errorf("storage: write doc %s: %w", doc.ID, err)
	}
	if err := s.writer.WriteByte('\n'); err != nil {
		return fmt.Errorf("storage: write newline: %w", err)
	}
	return nil
}

func (s *JSONLStore) Close() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if err := s.writer.Flush(); err != nil {
		s.file.Close()
		return fmt.Errorf("storage: flush: %w", err)
	}
	return s.file.Close()
}
