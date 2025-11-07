package service

import (
	"context"
	"fmt"
	"pubmed-api/internal/domain"
	"pubmed-api/internal/repo"
	"strconv"
)

// ArticleService handles business logic for articles
type ArticleService struct {
	repo repo.ArticleRepository
}

// NewArticleService creates a new article service
func NewArticleService(repo repo.ArticleRepository) *ArticleService {
	return &ArticleService{repo: repo}
}

// GetArticle retrieves an article by its PubMed ID
func (s *ArticleService) GetArticle(ctx context.Context, pmid string) (*domain.Article, error) {
	if pmid == "" {
		return nil, fmt.Errorf("pmid is required")
	}

	return s.repo.FindByID(ctx, pmid)
}

// SearchArticles performs a search with filters, pagination, and sorting
func (s *ArticleService) SearchArticles(ctx context.Context, filters *domain.SearchFilters) (*domain.SearchResult, error) {
	// Validate and normalize filters
	if filters.Page < 1 {
		filters.Page = 1
	}

	if filters.PageSize < 1 {
		filters.PageSize = 10
	}

	if filters.PageSize > 50 {
		filters.PageSize = 50
	}

	if filters.Sort == "" {
		filters.Sort = "relevance"
	}

	validSorts := map[string]bool{
		"relevance": true,
		"year_desc": true,
		"year_asc":  true,
	}

	if !validSorts[filters.Sort] {
		filters.Sort = "relevance"
	}

	return s.repo.Search(ctx, filters)
}

// GetStats returns aggregate statistics
func (s *ArticleService) GetStats(ctx context.Context) (*domain.Stats, error) {
	return s.repo.GetStats(ctx)
}

// ParseSearchFilters parses query parameters into SearchFilters
func ParseSearchFilters(queryParams map[string][]string) *domain.SearchFilters {
	filters := &domain.SearchFilters{
		Page:     1,
		PageSize: 10,
		Sort:     "relevance",
	}

	if q := queryParams["q"]; len(q) > 0 && q[0] != "" {
		filters.Query = q[0]
	}

	if yearStr := queryParams["year"]; len(yearStr) > 0 && yearStr[0] != "" {
		if year, err := strconv.Atoi(yearStr[0]); err == nil {
			filters.Year = &year
		}
	}

	if journal := queryParams["journal"]; len(journal) > 0 && journal[0] != "" {
		filters.Journal = journal[0]
	}

	if author := queryParams["author"]; len(author) > 0 && author[0] != "" {
		filters.Author = author[0]
	}

	if pageStr := queryParams["page"]; len(pageStr) > 0 && pageStr[0] != "" {
		if page, err := strconv.Atoi(pageStr[0]); err == nil && page > 0 {
			filters.Page = page
		}
	}

	if pageSizeStr := queryParams["page_size"]; len(pageSizeStr) > 0 && pageSizeStr[0] != "" {
		if pageSize, err := strconv.Atoi(pageSizeStr[0]); err == nil && pageSize > 0 {
			filters.PageSize = pageSize
		}
	}

	if sort := queryParams["sort"]; len(sort) > 0 && sort[0] != "" {
		filters.Sort = sort[0]
	}

	return filters
}
