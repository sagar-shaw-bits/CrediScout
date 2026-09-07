package crawler

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"log/slog"
	"net/url"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/gocolly/colly/v2"
	"github.com/gocolly/colly/v2/extensions"

	"github.com/sagar-shaw-bits/CrediScout/internal/config"
	"github.com/sagar-shaw-bits/CrediScout/internal/domain"
	"github.com/sagar-shaw-bits/CrediScout/internal/storage"
)

type Stats struct {
	Fetched int
	Stored  int
	Errors  int
}

type Crawler struct {
	cfg           config.CrawlerConfig
	store         storage.Store
	log           *slog.Logger
	allowed       []string
	keywords      []string
	blocklist     []string
	seenContentMu sync.Mutex
	seenContent   map[string]struct{}

	stored  int64
	fetched int64
	errors  int64
	done    int32

	stopOnce sync.Once
	stopCh   chan struct{}
}

func New(cfg config.CrawlerConfig, store storage.Store, log *slog.Logger) (*Crawler, error) {
	if err := cfg.Validate(); err != nil {
		return nil, err
	}
	if log == nil {
		log = slog.Default()
	}

	allowed := make([]string, 0)
	for _, d := range cfg.AllowedDomains {
		if d = strings.TrimSpace(d); d != "" {
			allowed = append(allowed, d)
		}
	}
	if len(allowed) == 0 {
		for _, seed := range cfg.Seeds {
			if u, err := url.Parse(seed); err == nil && u.Hostname() != "" {
				allowed = append(allowed, u.Hostname())
			}
		}
	}

	keywords := make([]string, 0, len(cfg.URLKeywords))
	for _, kw := range cfg.URLKeywords {
		if kw = strings.ToLower(strings.TrimSpace(kw)); kw != "" {
			keywords = append(keywords, kw)
		}
	}

	blocklist := make([]string, 0, len(cfg.URLBlocklist))
	for _, b := range cfg.URLBlocklist {
		if b = strings.ToLower(strings.TrimSpace(b)); b != "" {
			blocklist = append(blocklist, b)
		}
	}

	return &Crawler{
		cfg:         cfg,
		store:       store,
		log:         log,
		allowed:     allowed,
		keywords:    keywords,
		blocklist:   blocklist,
		seenContent: make(map[string]struct{}),
		stopCh:      make(chan struct{}),
	}, nil
}

func (c *Crawler) Run(ctx context.Context) (Stats, error) {
	col := colly.NewCollector(
		colly.AllowedDomains(c.allowed...),
		colly.Async(true),
	)
	col.SetRequestTimeout(c.cfg.RequestTimeout)
	extensions.RandomUserAgent(col)

	err := col.Limit(&colly.LimitRule{
		DomainGlob:  "*",
		Parallelism: c.cfg.Concurrency,
		Delay:       c.cfg.PolitenessDelay,
	})

	if err != nil {
		return Stats{}, err
	}

	go func() {
		<-ctx.Done()
		c.stop()
	}()

	col.OnRequest(func(r *colly.Request) {
		if c.isDone() {
			r.Abort()
		}
	})

	col.OnHTML("a[href]", func(e *colly.HTMLElement) {
		if c.isDone() {
			return
		}
		pageDepth := e.Request.Depth - 1
		if pageDepth >= c.cfg.MaxDepth {
			return
		}
		link := e.Request.AbsoluteURL(e.Attr("href"))
		if !c.wantURL(link) {
			return
		}
		_ = e.Request.Visit(link)
	})

	col.OnHTML("html", func(e *colly.HTMLElement) {
		if c.isDone() {
			return
		}
		atomic.AddInt64(&c.fetched, 1)

		content := extractText(e)
		if len(content) < c.cfg.MinContentLen {
			return
		}

		hash := hashString(content)
		if c.seenContentBefore(hash) {
			c.log.Debug("skipping duplicate content", "url", e.Request.URL.String())
			return
		}

		if n := atomic.AddInt64(&c.stored, 1); n > int64(c.cfg.MaxPages) {
			atomic.AddInt64(&c.stored, -1)
			c.stop()
			return
		}

		rawURL := e.Request.URL.String()
		doc := domain.Document{
			ID:    hashString(rawURL),
			URL:   rawURL,
			Title: strings.TrimSpace(e.ChildText("title")),
			Text:  content,
		}

		if err := c.store.Save(ctx, doc); err != nil {
			atomic.AddInt64(&c.stored, -1)
			atomic.AddInt64(&c.errors, 1)
			c.log.Error("store failed", "url", rawURL, "err", err.Error())
			return
		}

		c.log.Info("stored", "url", rawURL, "depth", e.Request.Depth-1, "title", doc.Title)

		if atomic.LoadInt64(&c.stored) >= int64(c.cfg.MaxPages) {
			c.stop()
		}
	})

	col.OnError(func(r *colly.Response, err error) {
		if c.isDone() {
			return
		}
		status := r.StatusCode
		retryable := status == 0 || status >= 500
		attempt := ctxAttempt(r.Request.Ctx)

		if retryable && attempt < c.cfg.MaxRetries {
			attempt++
			r.Request.Ctx.Put("attempt", strconv.Itoa(attempt))
			backoff := c.cfg.RetryBackoff * time.Duration(int64(1)<<(attempt-1))
			select {
			case <-time.After(backoff):
			case <-ctx.Done():
				return
			}
			c.log.Debug("retrying", "url", r.Request.URL.String(), "status", status, "attempt", attempt)
			_ = r.Request.Retry()
			return
		}

		atomic.AddInt64(&c.errors, 1)
		c.log.Warn("fetch failed", "url", r.Request.URL.String(), "status", status, "err", err.Error())
	})

	for _, seed := range c.cfg.Seeds {
		if c.wantURL(seed) {
			if err := col.Visit(seed); err != nil {
				c.log.Warn("failed to visit seed", "seed", seed, "err", err.Error())
			}
		}
	}

	waitCh := make(chan struct{})
	go func() {
		col.Wait()
		close(waitCh)
	}()
	select {
	case <-waitCh:
	case <-c.stopCh:
	}

	stats := Stats{
		Fetched: int(atomic.LoadInt64(&c.fetched)),
		Stored:  int(atomic.LoadInt64(&c.stored)),
		Errors:  int(atomic.LoadInt64(&c.errors)),
	}
	return stats, ctx.Err()
}

func (c *Crawler) stop() {
	c.stopOnce.Do(func() {
		atomic.StoreInt32(&c.done, 1)
		close(c.stopCh)
	})
}

func (c *Crawler) isDone() bool {
	return atomic.LoadInt32(&c.done) == 1 ||
		atomic.LoadInt64(&c.stored) >= int64(c.cfg.MaxPages)
}

var assetExtensions = []string{
	".pdf", ".jpg", ".jpeg", ".png", ".gif", ".svg", ".webp",
	".css", ".js", ".zip", ".mp4", ".ico", ".woff", ".woff2",
}

func (c *Crawler) wantURL(rawURL string) bool {
	if rawURL == "" {
		return false
	}
	lower := strings.ToLower(rawURL)
	for _, ext := range assetExtensions {
		if strings.HasSuffix(lower, ext) {
			return false
		}
	}
	for _, b := range c.blocklist {
		if strings.Contains(lower, b) {
			return false
		}
	}
	if len(c.keywords) == 0 {
		return true
	}
	for _, kw := range c.keywords {
		if strings.Contains(lower, kw) {
			return true
		}
	}
	return false
}

func (c *Crawler) seenContentBefore(hash string) bool {
	c.seenContentMu.Lock()
	defer c.seenContentMu.Unlock()
	if _, ok := c.seenContent[hash]; ok {
		return true
	}
	c.seenContent[hash] = struct{}{}
	return false
}

func extractText(e *colly.HTMLElement) string {
	var b strings.Builder
	e.ForEach("p, li, h1, h2, h3, h4, td, th", func(_ int, el *colly.HTMLElement) {
		t := strings.TrimSpace(el.Text)
		if t == "" {
			return
		}
		b.WriteString(strings.Join(strings.Fields(t), " "))
		b.WriteString("\n")
	})
	return strings.TrimSpace(b.String())
}

func ctxAttempt(ctx *colly.Context) int {
	if s := ctx.Get("attempt"); s != "" {
		if n, err := strconv.Atoi(s); err == nil {
			return n
		}
	}
	return 0
}

func hashString(input string) string {
	sum := sha256.Sum256([]byte(input))
	return hex.EncodeToString(sum[:])
}
