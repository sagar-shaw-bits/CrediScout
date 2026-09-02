package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/sagar-shaw-bits/CrediScout/internal/config"
	"github.com/sagar-shaw-bits/CrediScout/internal/crawler"
	"github.com/sagar-shaw-bits/CrediScout/internal/storage"
)

func main() {
	fmt.Println("Starting CrediScout crawler...")
	cfg := config.ParseCrawlerFlags()

	log := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}))

	store, err := storage.NewJSONLStore(cfg.OutputFile)
	if err != nil {
		log.Error("unable to initialize storage", "err", err)
		os.Exit(1)
	}
	defer func() {
		if err := store.Close(); err != nil {
			log.Error("unable to close storage", "err", err)
		}
	}()

	c, err := crawler.New(cfg, store, log)
	if err != nil {
		log.Error("unable to initialize crawler", "err", err)
		os.Exit(1)
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	log.Info("crawl started",
		"seeds", cfg.Seeds,
		"max_pages", cfg.MaxPages,
		"concurrency", cfg.Concurrency,
	)
	start := time.Now()

	stats, err := c.Run(ctx)
	if err != nil {
		log.Warn("crawl ended early", "err", err)
	}

	log.Info("crawl finished",
		"stored", stats.Stored,
		"fetched", stats.Fetched,
		"errors", stats.Errors,
		"elapsed", time.Since(start).Round(time.Millisecond).String(),
	)
}
