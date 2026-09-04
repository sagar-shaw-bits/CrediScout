package crawler

import (
	"context"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/sagar-shaw-bits/CrediScout/internal/config"
	"github.com/sagar-shaw-bits/CrediScout/internal/domain"
)

type memStore struct {
	mu   sync.Mutex
	docs map[string]domain.Document
}

type nopWriter struct{}

func (nopWriter) Write(p []byte) (int, error) { return len(p), nil }

func newMemStore() *memStore { return &memStore{docs: make(map[string]domain.Document)} }

func (m *memStore) Save(_ context.Context, doc domain.Document) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.docs[doc.URL] = doc
	return nil
}

func (m *memStore) len() int {
	m.mu.Lock()
	defer m.mu.Unlock()
	return len(m.docs)
}

func testSite() *httptest.Server {
	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		_, _ = w.Write([]byte(`<html><head><title>Home</title></head>
			<body><h1>Home</h1>
			<a href="/a">A</a> <a href="/b">B</a>
			<a href="https://external.example.com/x">External</a>
			</body></html>`))
	})
	mux.HandleFunc("/a", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		_, _ = w.Write([]byte(`<html><head><title>Page A</title></head>
			<body><p>Alpha content</p><a href="/c">C</a></body></html>`))
	})
	mux.HandleFunc("/b", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		_, _ = w.Write([]byte(`<html><head><title>Page B</title></head>
			<body><p>Beta content</p><a href="/a">back to A</a></body></html>`))
	})
	mux.HandleFunc("/c", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		_, _ = w.Write([]byte(`<html><head><title>Page C</title></head>
			<body><p>Gamma content</p></body></html>`))
	})
	return httptest.NewServer(mux)
}

func testConfig(seed string) config.CrawlerConfig {
	cfg := config.DefaultCrawlerConfig()
	cfg.Seeds = []string{seed}
	cfg.URLKeywords = nil
	cfg.MinContentLen = 0
	cfg.MaxPages = 100
	cfg.MaxDepth = 5
	cfg.Concurrency = 4
	cfg.PolitenessDelay = 0
	cfg.RequestTimeout = 5 * time.Second
	return cfg
}

func TestCrawl_VisitsSameDomainOnly(t *testing.T) {
	srv := testSite()
	defer srv.Close()

	store := newMemStore()
	c, err := New(testConfig(srv.URL), store, slog.New(slog.NewTextHandler(nopWriter{}, nil)))
	require.NoError(t, err)

	stats, err := c.Run(context.Background())
	require.NoError(t, err)

	require.Equal(t, 4, store.len(), "expected 4 stored pages")
	assert.Equal(t, 4, stats.Stored)
	for u := range store.docs {
		assert.True(t, strings.HasPrefix(u, srv.URL), "crawled off-domain URL: %s", u)
	}
}

func TestCrawl_RespectsMaxPages(t *testing.T) {
	srv := testSite()
	defer srv.Close()

	cfg := testConfig(srv.URL)
	cfg.MaxPages = 2

	store := newMemStore()
	c, err := New(cfg, store, slog.New(slog.NewTextHandler(nopWriter{}, nil)))
	require.NoError(t, err)

	stats, err := c.Run(context.Background())
	require.NoError(t, err)
	assert.Equal(t, 2, stats.Stored, "expected exactly 2 stored pages")
}

func TestCrawl_RespectsMaxDepth(t *testing.T) {
	srv := testSite()
	defer srv.Close()

	cfg := testConfig(srv.URL)
	cfg.MaxDepth = 0 // only the seed page

	store := newMemStore()
	c, err := New(cfg, store, slog.New(slog.NewTextHandler(nopWriter{}, nil)))
	require.NoError(t, err)

	_, err = c.Run(context.Background())
	require.NoError(t, err)
	assert.Equal(t, 1, store.len(), "expected only the seed page at depth 0")
}

func TestCrawl_OnlyFollowsKeywordURLs(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/cards/home", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		_, _ = w.Write([]byte(`<html><head><title>Cards</title></head><body><p>` +
			strings.Repeat("card content ", 30) +
			`</p><a href="/cards/gold">gold</a><a href="/loans/home">loan</a></body></html>`))
	})
	mux.HandleFunc("/cards/gold", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		_, _ = w.Write([]byte(`<html><head><title>Gold</title></head><body><p>` +
			strings.Repeat("gold card ", 30) + `</p></body></html>`))
	})
	mux.HandleFunc("/loans/home", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		_, _ = w.Write([]byte(`<html><head><title>Loan</title></head><body><p>` +
			strings.Repeat("loan content ", 30) + `</p></body></html>`))
	})
	srv := httptest.NewServer(mux)
	defer srv.Close()

	cfg := testConfig(srv.URL + "/cards/home")
	cfg.URLKeywords = []string{"cards"}

	store := newMemStore()
	c, err := New(cfg, store, slog.New(slog.NewTextHandler(nopWriter{}, nil)))
	require.NoError(t, err)
	_, err = c.Run(context.Background())
	require.NoError(t, err)
	require.Equal(t, 2, store.len(), "expected 2 cards pages")
	for u := range store.docs {
		assert.NotContains(t, u, "/loans/", "crawled off-topic loan URL: %s", u)
	}
}

func TestConfigValidate(t *testing.T) {
	cfg := config.DefaultCrawlerConfig()
	assert.Error(t, cfg.Validate(), "expected error for missing seeds")
	cfg.Seeds = []string{"https://example.com"}
	assert.NoError(t, cfg.Validate())
}

func TestCrawl_DedupsIdenticalContent(t *testing.T) {
	body := `<html><head><title>Dup</title></head><body><p>` +
		strings.Repeat("identical content ", 30) +
		`</p><a href="/twin">twin</a></body></html>`
	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		_, _ = w.Write([]byte(body))
	})
	mux.HandleFunc("/twin", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		_, _ = w.Write([]byte(body))
	})
	srv := httptest.NewServer(mux)
	defer srv.Close()

	store := newMemStore()
	c, err := New(testConfig(srv.URL), store, slog.New(slog.NewTextHandler(nopWriter{}, nil)))
	require.NoError(t, err)
	_, err = c.Run(context.Background())
	require.NoError(t, err)
	assert.Equal(t, 1, store.len(), "expected 1 stored page after content dedup")
}

func TestCrawl_RetriesOn5xx(t *testing.T) {
	var hits int32
	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if atomic.AddInt32(&hits, 1) <= 2 {
			w.WriteHeader(http.StatusServiceUnavailable)
			return
		}
		w.Header().Set("Content-Type", "text/html")
		_, _ = w.Write([]byte(`<html><head><title>OK</title></head><body><p>` +
			strings.Repeat("recovered content ", 30) + `</p></body></html>`))
	})
	srv := httptest.NewServer(mux)
	defer srv.Close()

	cfg := testConfig(srv.URL)
	cfg.MaxDepth = 0
	cfg.MaxRetries = 3
	cfg.RetryBackoff = time.Millisecond

	store := newMemStore()
	c, err := New(cfg, store, slog.New(slog.NewTextHandler(nopWriter{}, nil)))
	require.NoError(t, err)
	_, err = c.Run(context.Background())
	require.NoError(t, err)
	assert.Equal(t, 1, store.len(), "expected 1 stored page after retrying 5xx")
	assert.Equal(t, int32(3), atomic.LoadInt32(&hits), "expected 3 attempts (2 failures + 1 success)")
}
