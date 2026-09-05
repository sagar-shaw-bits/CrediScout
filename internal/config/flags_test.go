package config

import (
	"flag"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSplitCSV(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  []string
	}{
		{"empty string", "", nil},
		{"only whitespace", "   ", nil},
		{"single value", "a", []string{"a"}},
		{"multiple values", "a,b,c", []string{"a", "b", "c"}},
		{"trims spaces", " a , b ,c ", []string{"a", "b", "c"}},
		{"skips empty parts", "a,,b,", []string{"a", "b"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, splitCSV(tt.input))
		})
	}
}

func TestParseCrawlerFlagsWhenPresentShouldOverwriteDefaults(t *testing.T) {
	oldArgs := os.Args
	defer func() { os.Args = oldArgs }()
	flag.CommandLine = flag.NewFlagSet(os.Args[0], flag.ExitOnError)

	os.Args = []string{
		"cmd",
		"--seeds", "https://a.com,https://b.com",
		"--domains", "a.com,b.com",
		"--keywords", "credit,card",
		"--blocklist", "talk:,file:",
		"--max-pages", "50",
		"--max-depth", "2",
		"--concurrency", "4",
		"--delay", "250ms",
		"--timeout", "10s",
		"--max-retries", "1",
		"--retry-backoff", "1s",
		"--out", "out.jsonl",
	}

	cfg := ParseCrawlerFlags()

	assert.Equal(t, []string{"https://a.com", "https://b.com"}, cfg.Seeds)
	assert.Equal(t, []string{"a.com", "b.com"}, cfg.AllowedDomains)
	assert.Equal(t, []string{"credit", "card"}, cfg.URLKeywords)
	assert.Equal(t, []string{"talk:", "file:"}, cfg.URLBlocklist)
	assert.Equal(t, 50, cfg.MaxPages)
	assert.Equal(t, 2, cfg.MaxDepth)
	assert.Equal(t, 4, cfg.Concurrency)
	assert.Equal(t, "out.jsonl", cfg.OutputFile)
}

func TestParseCrawlerFlagsWhenBlocklistEmptyShouldKeepDefaults(t *testing.T) {
	oldArgs := os.Args
	defer func() { os.Args = oldArgs }()
	flag.CommandLine = flag.NewFlagSet(os.Args[0], flag.ExitOnError)

	os.Args = []string{"cmd", "--seeds", "https://a.com"}

	cfg := ParseCrawlerFlags()

	assert.Equal(t, DefaultCrawlerConfig().URLBlocklist, cfg.URLBlocklist)
}
