package repo

import (
	"context"
	"pubmed-api/internal/domain"
)

// ArticleRepository defines the interface for article data access
type ArticleRepository interface {
	// FindByID retrieves an article by its PubMed ID
	FindByID(ctx context.Context, pmid string) (*domain.Article, error)

	// Search performs a search with filters and pagination
	Search(ctx context.Context, filters *domain.SearchFilters) (*domain.SearchResult, error)

	// GetStats returns aggregate statistics
	GetStats(ctx context.Context) (*domain.Stats, error)

	// LoadData loads articles from a data source (file, S3, etc.)
	LoadData(ctx context.Context, dataPath string) error
}

