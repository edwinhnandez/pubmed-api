package platform

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"os"
	"pubmed-api/internal/domain"
	"pubmed-api/internal/repo"
	"strings"
)

// LoadArticles loads articles from various sources (S3, local file, or embedded)
func LoadArticles(ctx context.Context, repo repo.ArticleRepository, cfg *Config, logger *slog.Logger) error {
	var data []byte
	var err error
	var source string

	// Priority 1: S3
	if cfg.DataS3URL != "" {
		data, err = LoadFromS3(ctx, cfg.DataS3URL, logger)
		if err != nil {
			logger.Warn("failed to load from S3, falling back", "error", err)
		} else {
			source = "S3"
		}
	}

	// Priority 2: Local file
	if data == nil {
		if _, err := os.Stat(cfg.DataPath); err == nil {
			data, err = os.ReadFile(cfg.DataPath)
			if err != nil {
				logger.Warn("failed to load from local file, falling back", "error", err)
			} else {
				source = "local file"
			}
		}
	}

	// Priority 3: Embedded fallback
	if data == nil {
		data = embeddedData
		source = "embedded"
		logger.Info("using embedded fallback data")
	}

	if data == nil {
		return fmt.Errorf("no data source available")
	}

	logger.Info("loading articles", "source", source)

	articles, err := parseJSONL(data)
	if err != nil {
		return fmt.Errorf("failed to parse JSONL: %w", err)
	}

	// Insert into repository - use type assertion with interface check
	// We'll need to add a method to insert articles in the interface
	// For now, we'll use a type assertion approach
	type articleInserter interface {
		InsertArticles(ctx context.Context, articles []*domain.Article) error
	}
	
	if inserter, ok := repo.(articleInserter); ok {
		if err := inserter.InsertArticles(ctx, articles); err != nil {
			return fmt.Errorf("failed to insert articles: %w", err)
		}
	}

	logger.Info("articles loaded successfully", "count", len(articles))
	return nil
}

// parseJSONL parses JSONL (JSON Lines) format
func parseJSONL(data []byte) ([]*domain.Article, error) {
	var articles []*domain.Article

	scanner := bufio.NewScanner(strings.NewReader(string(data)))
	for scanner.Scan() {
		line := scanner.Bytes()
		if len(line) == 0 {
			continue
		}

		var article domain.Article
		if err := json.Unmarshal(line, &article); err != nil {
			return nil, fmt.Errorf("failed to unmarshal article: %w", err)
		}

		articles = append(articles, &article)
	}

	if err := scanner.Err(); err != nil && err != io.EOF {
		return nil, fmt.Errorf("failed to scan JSONL: %w", err)
	}

	return articles, nil
}

