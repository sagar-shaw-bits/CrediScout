package config

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDefaultCrawlerConfigShouldSetDefaults(t *testing.T) {
	cfg := DefaultCrawlerConfig()

	assert.Nil(t, cfg.Seeds)
	assert.Nil(t, cfg.AllowedDomains)
	assert.Equal(t, []string{"credit", "card", "payment", "debit", "bank", "loan"}, cfg.URLKeywords)
	assert.NotEmpty(t, cfg.URLBlocklist)
	assert.Equal(t, 200, cfg.MaxPages)
	assert.Equal(t, 3, cfg.MaxDepth)
	assert.Equal(t, 8, cfg.Concurrency)
	assert.Equal(t, 15*time.Second, cfg.RequestTimeout)
	assert.Equal(t, 500*time.Millisecond, cfg.PolitenessDelay)
	assert.Equal(t, "data/documents.jsonl", cfg.OutputFile)
	assert.Equal(t, 2, cfg.MaxRetries)
	assert.Equal(t, 500*time.Millisecond, cfg.RetryBackoff)
	assert.Equal(t, 200, cfg.MinContentLen)
}

func validCrawlerConfig() CrawlerConfig {
	cfg := DefaultCrawlerConfig()
	cfg.Seeds = []string{"https://example.com"}
	return cfg
}

func TestCrawlerConfigValidateWhenValidShouldReturnNil(t *testing.T) {
	assert.Nil(t, validCrawlerConfig().Validate())
}

func TestCrawlerConfigValidateWhenInvalidShouldReturnError(t *testing.T) {
	tests := []struct {
		name    string
		mutate  func(*CrawlerConfig)
		wantErr error
	}{
		{"no seeds", func(c *CrawlerConfig) { c.Seeds = nil }, ErrCrawlerNoSeeds},
		{"non positive max pages", func(c *CrawlerConfig) { c.MaxPages = 0 }, ErrCrawlerMaxPages},
		{"negative max depth", func(c *CrawlerConfig) { c.MaxDepth = -1 }, ErrCrawlerMaxDepth},
		{"non positive concurrency", func(c *CrawlerConfig) { c.Concurrency = 0 }, ErrCrawlerConcurrency},
		{"negative max retries", func(c *CrawlerConfig) { c.MaxRetries = -1 }, ErrCrawlerMaxRetries},
		{"empty output file", func(c *CrawlerConfig) { c.OutputFile = "" }, ErrCrawlerOutputFile},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := validCrawlerConfig()
			tt.mutate(&cfg)

			err := cfg.Validate()
			require.Error(t, err)
			assert.ErrorIs(t, err, tt.wantErr)
		})
	}
}
