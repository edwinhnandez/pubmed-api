package http

import (
	"context"
	"pubmed-api/internal/domain"
	"pubmed-api/internal/service"
)

// ArticleServiceInterface defines the interface for article service operations
// This allows for easier testing with mocks
type ArticleServiceInterface interface {
	GetArticle(ctx context.Context, pmid string) (*domain.Article, error)
	SearchArticles(ctx context.Context, filters *domain.SearchFilters) (*domain.SearchResult, error)
	GetStats(ctx context.Context) (*domain.Stats, error)
}

// Ensure ArticleService implements the interface
var _ ArticleServiceInterface = (*service.ArticleService)(nil)

