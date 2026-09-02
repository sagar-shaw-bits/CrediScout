package config

import (
	"flag"
	"strings"
)

func ParseCrawlerFlags() CrawlerConfig {
	cfg := DefaultCrawlerConfig()

	var (
		seeds     string
		domains   string
		keywords  string
		blocklist string
	)
	flag.StringVar(&seeds, "seeds", "https://en.wikipedia.org/wiki/Credit_card",
		"comma-separated seed URLs")
	flag.StringVar(&domains, "domains", "en.wikipedia.org",
		"comma-separated allowed domains (defaults to seed hosts)")
	flag.StringVar(&keywords, "keywords", "credit,card,payment,debit,bank,loan",
		"comma-separated URL keywords to stay on-topic (empty to allow all)")
	flag.StringVar(&blocklist, "blocklist", "",
		"comma-separated URL substrings to skip (empty keeps built-in defaults)")
	flag.IntVar(&cfg.MaxPages, "max-pages", cfg.MaxPages, "maximum pages to store")
	flag.IntVar(&cfg.MaxDepth, "max-depth", cfg.MaxDepth, "maximum link depth from seeds")
	flag.IntVar(&cfg.Concurrency, "concurrency", cfg.Concurrency, "number of parallel fetch workers")
	flag.DurationVar(&cfg.PolitenessDelay, "delay", cfg.PolitenessDelay, "minimum delay between requests to the same host")
	flag.DurationVar(&cfg.RequestTimeout, "timeout", cfg.RequestTimeout, "per-request timeout")
	flag.IntVar(&cfg.MaxRetries, "max-retries", cfg.MaxRetries, "retries on network error or 5xx")
	flag.DurationVar(&cfg.RetryBackoff, "retry-backoff", cfg.RetryBackoff, "initial backoff before a retry (doubles each attempt)")
	flag.StringVar(&cfg.OutputFile, "out", cfg.OutputFile, "JSONL file to write crawled documents")
	flag.Parse()

	cfg.Seeds = splitCSV(seeds)
	cfg.AllowedDomains = splitCSV(domains)
	cfg.URLKeywords = splitCSV(keywords)
	if bl := splitCSV(blocklist); bl != nil {
		cfg.URLBlocklist = bl
	}

	return cfg
}

func splitCSV(s string) []string {
	if strings.TrimSpace(s) == "" {
		return nil
	}
	parts := strings.Split(s, ",")
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		if p = strings.TrimSpace(p); p != "" {
			out = append(out, p)
		}
	}
	return out
}
