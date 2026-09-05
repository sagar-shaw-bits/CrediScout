package config

import (
	"flag"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDefaultServerConfigShouldSetDefaults(t *testing.T) {
	cfg := DefaultServerConfig()

	assert.Equal(t, ":8080", cfg.Addr)
	assert.Equal(t, "data/index.jsonl", cfg.IndexFile)
	assert.Equal(t, "http://localhost:11434", cfg.OllamaURL)
	assert.Equal(t, "nomic-embed-text", cfg.EmbedModel)
	assert.Equal(t, "llama3.2", cfg.ChatModel)
	assert.Equal(t, 5, cfg.TopK)
	assert.Equal(t, 120*time.Second, cfg.RequestTimeout)
}

func TestServerConfigValidateWhenValidShouldReturnNil(t *testing.T) {
	assert.Nil(t, DefaultServerConfig().Validate())
}

func TestServerConfigValidateWhenInvalidShouldReturnError(t *testing.T) {
	tests := []struct {
		name    string
		mutate  func(*ServerConfig)
		wantErr error
	}{
		{"empty addr", func(c *ServerConfig) { c.Addr = "" }, ErrServerAddr},
		{"empty index file", func(c *ServerConfig) { c.IndexFile = "" }, ErrServerIndexFile},
		{"empty ollama url", func(c *ServerConfig) { c.OllamaURL = "" }, ErrServerOllamaURL},
		{"empty embed model", func(c *ServerConfig) { c.EmbedModel = "" }, ErrServerEmbedModel},
		{"empty chat model", func(c *ServerConfig) { c.ChatModel = "" }, ErrServerChatModel},
		{"non positive top-k", func(c *ServerConfig) { c.TopK = 0 }, ErrServerTopK},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := DefaultServerConfig()
			tt.mutate(&cfg)

			err := cfg.Validate()
			require.Error(t, err)
			assert.ErrorIs(t, err, tt.wantErr)
		})
	}
}

func TestEnvStr(t *testing.T) {
	t.Run("when env is set should return value", func(t *testing.T) {
		t.Setenv("TEST_ENV_STR", "value")
		assert.Equal(t, "value", envStr("TEST_ENV_STR", "fallback"))
	})

	t.Run("when env is empty should return fallback", func(t *testing.T) {
		t.Setenv("TEST_ENV_STR", "")
		assert.Equal(t, "fallback", envStr("TEST_ENV_STR", "fallback"))
	})
}

func TestEnvInt(t *testing.T) {
	t.Run("when env is a valid int should return value", func(t *testing.T) {
		t.Setenv("TEST_ENV_INT", "42")
		assert.Equal(t, 42, envInt("TEST_ENV_INT", 7))
	})

	t.Run("when env is not a number should return fallback", func(t *testing.T) {
		t.Setenv("TEST_ENV_INT", "abc")
		assert.Equal(t, 7, envInt("TEST_ENV_INT", 7))
	})

	t.Run("when env is empty should return fallback", func(t *testing.T) {
		t.Setenv("TEST_ENV_INT", "")
		assert.Equal(t, 7, envInt("TEST_ENV_INT", 7))
	})
}

func TestParseServerFlagsWhenPresentShouldOverwriteDefaults(t *testing.T) {
	oldArgs := os.Args
	defer func() { os.Args = oldArgs }()
	flag.CommandLine = flag.NewFlagSet(os.Args[0], flag.ExitOnError)

	os.Args = []string{
		"cmd",
		"--addr", ":9090",
		"--index", "custom-index.jsonl",
		"--ollama-url", "http://ollama:1234",
		"--embed-model", "test-embed",
		"--chat-model", "test-chat",
		"--top-k", "9",
		"--timeout", "30s",
	}

	cfg := ParseServerFlags()

	assert.Equal(t, ":9090", cfg.Addr)
	assert.Equal(t, "custom-index.jsonl", cfg.IndexFile)
	assert.Equal(t, "http://ollama:1234", cfg.OllamaURL)
	assert.Equal(t, "test-embed", cfg.EmbedModel)
	assert.Equal(t, "test-chat", cfg.ChatModel)
	assert.Equal(t, 9, cfg.TopK)
	assert.Equal(t, 30*time.Second, cfg.RequestTimeout)
}

func TestParseServerFlagsWhenEnvSetShouldUseEnv(t *testing.T) {
	oldArgs := os.Args
	defer func() { os.Args = oldArgs }()
	flag.CommandLine = flag.NewFlagSet(os.Args[0], flag.ExitOnError)

	t.Setenv("SERVER_ADDR", ":7000")
	t.Setenv("TOP_K", "3")
	os.Args = []string{"cmd"}

	cfg := ParseServerFlags()

	assert.Equal(t, ":7000", cfg.Addr)
	assert.Equal(t, 3, cfg.TopK)
}
