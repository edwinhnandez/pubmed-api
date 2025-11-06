package http

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"pubmed-api/internal/domain"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"log/slog"
)

// mockService is a mock implementation of ArticleServiceInterface
type mockService struct {
	articles map[string]*domain.Article
	stats    *domain.Stats
}

// Ensure mockService implements ArticleServiceInterface
var _ ArticleServiceInterface = (*mockService)(nil)

func newMockService() *mockService {
	return &mockService{
		articles: map[string]*domain.Article{
			"12345678": {
				PMID:     "12345678",
				Title:    "Test Article",
				Abstract: "Test abstract",
				Authors:  []string{"Author A"},
				Journal:  "Test Journal",
				PubYear:  2020,
			},
		},
		stats: &domain.Stats{
			TopJournals:   []domain.JournalCount{{Journal: "Test Journal", Count: 5}},
			YearHistogram: map[int]int{2020: 3},
		},
	}
}

func (m *mockService) GetArticle(ctx context.Context, pmid string) (*domain.Article, error) {
	article, ok := m.articles[pmid]
	if !ok {
		return nil, errors.New("article not found")
	}
	return article, nil
}

func (m *mockService) SearchArticles(ctx context.Context, filters *domain.SearchFilters) (*domain.SearchResult, error) {
	var results []*domain.Article
	for _, article := range m.articles {
		if filters.Query == "" || contains(article.Title, filters.Query) || contains(article.Abstract, filters.Query) {
			results = append(results, article)
		}
	}

	return &domain.SearchResult{
		Items:    results,
		Page:     filters.Page,
		PageSize: filters.PageSize,
		Total:    len(results),
		TookMs:   1,
	}, nil
}

func (m *mockService) GetStats(ctx context.Context) (*domain.Stats, error) {
	return m.stats, nil
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(substr) == 0)
}

func TestHandler_Healthz(t *testing.T) {
	logger := slog.Default()
	// Healthz doesn't use service, so we can pass nil
	handler := &Handler{
		service: nil,
		logger:  logger,
	}

	req := httptest.NewRequest("GET", "/healthz", nil)
	w := httptest.NewRecorder()

	handler.Healthz(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Header().Get("Content-Type"), "application/json")

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)
	assert.Equal(t, "pubmed-api", response["service"])
	assert.Equal(t, "ok", response["status"])
}

func TestHandler_GetArticle(t *testing.T) {
	logger := slog.Default()
	mockSvc := newMockService()
	handler := &Handler{
		service: mockSvc,
		logger:  logger,
	}

	tests := []struct {
		name       string
		pmid       string
		statusCode int
	}{
		{
			name:       "successful retrieval",
			pmid:       "12345678",
			statusCode: http.StatusOK,
		},
		{
			name:       "not found",
			pmid:       "99999999",
			statusCode: http.StatusNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest("GET", "/v1/articles/"+tt.pmid, nil)
			w := httptest.NewRecorder()

			// Create a chi router context
			rctx := chi.NewRouteContext()
			rctx.URLParams.Add("pmid", tt.pmid)
			req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

			handler.GetArticle(w, req)

			assert.Equal(t, tt.statusCode, w.Code)

			if tt.statusCode == http.StatusOK {
				var article domain.Article
				err := json.Unmarshal(w.Body.Bytes(), &article)
				require.NoError(t, err)
				assert.Equal(t, tt.pmid, article.PMID)
			}
		})
	}
}

func TestHandler_GetArticles(t *testing.T) {
	logger := slog.Default()
	mockSvc := newMockService()
	handler := &Handler{
		service: mockSvc,
		logger:  logger,
	}

	req := httptest.NewRequest("GET", "/v1/articles?q=test&page=1&page_size=10", nil)
	w := httptest.NewRecorder()

	handler.GetArticles(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var result domain.SearchResult
	err := json.Unmarshal(w.Body.Bytes(), &result)
	require.NoError(t, err)
	assert.GreaterOrEqual(t, result.Total, 0)
}

func TestHandler_GetStats(t *testing.T) {
	logger := slog.Default()
	mockSvc := newMockService()
	handler := &Handler{
		service: mockSvc,
		logger:  logger,
	}

	req := httptest.NewRequest("GET", "/v1/stats", nil)
	w := httptest.NewRecorder()

	handler.GetStats(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var stats domain.Stats
	err := json.Unmarshal(w.Body.Bytes(), &stats)
	require.NoError(t, err)
	assert.NotNil(t, stats.TopJournals)
	assert.NotNil(t, stats.YearHistogram)
}

