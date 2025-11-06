package service

import (
	"context"
	"errors"
	"pubmed-api/internal/domain"
	"strings"
	"testing"
)

// mockRepository is a mock implementation of ArticleRepository
type mockRepository struct {
	articles map[string]*domain.Article
}

func newMockRepository() *mockRepository {
	return &mockRepository{
		articles: make(map[string]*domain.Article),
	}
}

func (m *mockRepository) FindByID(ctx context.Context, pmid string) (*domain.Article, error) {
	article, ok := m.articles[pmid]
	if !ok {
		return nil, errors.New("article not found")
	}
	return article, nil
}

func (m *mockRepository) Search(ctx context.Context, filters *domain.SearchFilters) (*domain.SearchResult, error) {
	// Simple mock search implementation
	var results []*domain.Article
	for _, article := range m.articles {
		matches := true

		if filters.Query != "" {
			queryLower := strings.ToLower(filters.Query)
			titleLower := strings.ToLower(article.Title)
			abstractLower := strings.ToLower(article.Abstract)
			if !strings.Contains(titleLower, queryLower) && !strings.Contains(abstractLower, queryLower) {
				matches = false
			}
		}

		if filters.Year != nil && article.PubYear != *filters.Year {
			matches = false
		}

		if filters.Journal != "" && article.Journal != filters.Journal {
			matches = false
		}

		if filters.Author != "" {
			found := false
			for _, author := range article.Authors {
				if strings.Contains(author, filters.Author) {
					found = true
					break
				}
			}
			if !found {
				matches = false
			}
		}

		if matches {
			results = append(results, article)
		}
	}

	// Apply pagination
	offset := (filters.Page - 1) * filters.PageSize
	end := offset + filters.PageSize
	if end > len(results) {
		end = len(results)
	}

	if offset >= len(results) {
		results = []*domain.Article{}
	} else {
		results = results[offset:end]
	}

	return &domain.SearchResult{
		Items:    results,
		Page:     filters.Page,
		PageSize: filters.PageSize,
		Total:    len(m.articles),
		TookMs:   1,
	}, nil
}

func (m *mockRepository) GetStats(ctx context.Context) (*domain.Stats, error) {
	return &domain.Stats{
		TopJournals:   []domain.JournalCount{{Journal: "Test Journal", Count: 5}},
		YearHistogram: map[int]int{2020: 3, 2021: 2},
	}, nil
}

func (m *mockRepository) LoadData(ctx context.Context, dataPath string) error {
	return nil
}

func TestArticleService_GetArticle(t *testing.T) {
	tests := []struct {
		name    string
		pmid    string
		wantErr bool
		setup   func(*mockRepository)
	}{
		{
			name:    "successful retrieval",
			pmid:    "12345678",
			wantErr: false,
			setup: func(m *mockRepository) {
				m.articles["12345678"] = &domain.Article{
					PMID:     "12345678",
					Title:    "Test Article",
					Abstract: "Test abstract",
					Authors:  []string{"Author A"},
					Journal:  "Test Journal",
					PubYear:  2020,
				}
			},
		},
		{
			name:    "not found",
			pmid:    "99999999",
			wantErr: true,
			setup:   func(m *mockRepository) {},
		},
		{
			name:    "empty pmid",
			pmid:    "",
			wantErr: true,
			setup:   func(m *mockRepository) {},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := newMockRepository()
			tt.setup(mockRepo)
			service := NewArticleService(mockRepo)

			article, err := service.GetArticle(context.Background(), tt.pmid)

			if tt.wantErr {
				if err == nil {
					t.Errorf("expected error but got none")
				}
				if article != nil {
					t.Errorf("expected nil article but got %v", article)
				}
			} else {
				if err != nil {
					t.Errorf("unexpected error: %v", err)
				}
				if article == nil {
					t.Errorf("expected article but got nil")
				}
				if article != nil && article.PMID != tt.pmid {
					t.Errorf("expected pmid %s but got %s", tt.pmid, article.PMID)
				}
			}
		})
	}
}

func TestArticleService_SearchArticles(t *testing.T) {
	mockRepo := newMockRepository()
	mockRepo.articles["1"] = &domain.Article{
		PMID:     "1",
		Title:    "Ibuprofen study",
		Abstract: "This is about ibuprofen",
		Authors:  []string{"Smith J"},
		Journal:  "Medical Journal",
		PubYear:  2020,
	}
	mockRepo.articles["2"] = &domain.Article{
		PMID:     "2",
		Title:    "Acetaminophen research",
		Abstract: "This is about acetaminophen",
		Authors:  []string{"Brown M"},
		Journal:  "Medical Journal",
		PubYear:  2021,
	}

	service := NewArticleService(mockRepo)

	tests := []struct {
		name           string
		filters        *domain.SearchFilters
		expectedCount  int
		expectedTotal  int
	}{
		{
			name: "search by query",
			filters: &domain.SearchFilters{
				Query:    "ibuprofen",
				Page:     1,
				PageSize: 10,
				Sort:     "relevance",
			},
			expectedCount: 1,
			expectedTotal: 2,
		},
		{
			name: "filter by year",
			filters: &domain.SearchFilters{
				Year:     intPtr(2020),
				Page:     1,
				PageSize: 10,
				Sort:     "relevance",
			},
			expectedCount: 1,
			expectedTotal: 2,
		},
		{
			name: "pagination",
			filters: &domain.SearchFilters{
				Page:     1,
				PageSize: 1,
				Sort:     "relevance",
			},
			expectedCount: 1,
			expectedTotal: 2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := service.SearchArticles(context.Background(), tt.filters)
			if err != nil {
				t.Errorf("unexpected error: %v", err)
				return
			}

			if len(result.Items) != tt.expectedCount {
				t.Errorf("expected %d items but got %d", tt.expectedCount, len(result.Items))
			}

			if result.Total != tt.expectedTotal {
				t.Errorf("expected total %d but got %d", tt.expectedTotal, result.Total)
			}
		})
	}
}

func intPtr(i int) *int {
	return &i
}

