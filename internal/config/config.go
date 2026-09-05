package config

import (
	"errors"
	"fmt"
	"time"
)

var (
	ErrCrawlerNoSeeds     = errors.New("At least one seed URL is required")
	ErrCrawlerMaxPages    = errors.New("Max pages must be positive")
	ErrCrawlerMaxDepth    = errors.New("Max depth cannot be negative")
	ErrCrawlerConcurrency = errors.New("Concurrency must be positive")
	ErrCrawlerMaxRetries  = errors.New("Max retries cannot be negative")
	ErrCrawlerOutputFile  = errors.New("Output file is required")
)

type CrawlerConfig struct {
	Seeds           []string
	AllowedDomains  []string
	URLKeywords     []string
	URLBlocklist    []string
	MaxPages        int
	MaxDepth        int
	Concurrency     int
	RequestTimeout  time.Duration
	PolitenessDelay time.Duration
	OutputFile      string
	MaxRetries      int
	RetryBackoff    time.Duration
	MinContentLen   int
}

func DefaultCrawlerConfig() CrawlerConfig {
	return CrawlerConfig{
		Seeds:          nil,
		AllowedDomains: nil,
		URLKeywords:    []string{"credit", "card", "payment", "debit", "bank", "loan"},
		URLBlocklist: []string{
			"action=", "printable=", "oldid=", "useparsoid=", "redlink=",
			"special:", "talk:", "help:", "template:", "category:",
			"portal:", "file:", "user:", "wikipedia:",
			"/hindi/", "/kannada/", "/malayalam/", "/tamil/", "/telugu/",
			"/marathi/", "/gujarati/", "/bengali/", "/punjabi/",
		},
		MaxPages:        200,
		MaxDepth:        3,
		Concurrency:     8,
		RequestTimeout:  15 * time.Second,
		PolitenessDelay: 500 * time.Millisecond,
		OutputFile:      "data/documents.jsonl",
		MaxRetries:      2,
		RetryBackoff:    500 * time.Millisecond,
		MinContentLen:   200,
	}
}

func (c CrawlerConfig) Validate() error {
	if len(c.Seeds) == 0 {
		return ErrCrawlerNoSeeds
	}
	if c.MaxPages <= 0 {
		return fmt.Errorf("%w, got %d", ErrCrawlerMaxPages, c.MaxPages)
	}
	if c.MaxDepth < 0 {
		return fmt.Errorf("%w, got %d", ErrCrawlerMaxDepth, c.MaxDepth)
	}
	if c.Concurrency <= 0 {
		return fmt.Errorf("%w, got %d", ErrCrawlerConcurrency, c.Concurrency)
	}
	if c.MaxRetries < 0 {
		return fmt.Errorf("%w, got %d", ErrCrawlerMaxRetries, c.MaxRetries)
	}
	if c.OutputFile == "" {
		return ErrCrawlerOutputFile
	}
	return nil
}
