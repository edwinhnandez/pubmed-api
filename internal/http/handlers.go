package http

import (
	"encoding/json"
	"net/http"
	"pubmed-api/internal/service"
	"time"

	"log/slog"

	"github.com/go-chi/chi/v5"
)

// Handler handles HTTP requests
type Handler struct {
	service ArticleServiceInterface
	logger  *slog.Logger
}

// NewHandler creates a new HTTP handler
func NewHandler(service ArticleServiceInterface, logger *slog.Logger) *Handler {
	return &Handler{
		service: service,
		logger:  logger,
	}
}

// Healthz handles health check requests
func (h *Handler) Healthz(w http.ResponseWriter, r *http.Request) {
	response := map[string]interface{}{
		"service": "pubmed-api",
		"version": "1.0.0",
		"status":  "ok",
		"uptime":  time.Since(startTime).Seconds(),
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

var startTime = time.Now()

// GetArticles handles GET /v1/articles requests
func (h *Handler) GetArticles(w http.ResponseWriter, r *http.Request) {
	filters := service.ParseSearchFilters(r.URL.Query())

	result, err := h.service.SearchArticles(r.Context(), filters)
	if err != nil {
		h.logger.Error("failed to search articles", "error", err)
		h.writeError(w, http.StatusInternalServerError, "failed to search articles")
		return
	}

	h.writeJSON(w, http.StatusOK, result)
}

// GetArticle handles GET /v1/articles/{pmid} requests
func (h *Handler) GetArticle(w http.ResponseWriter, r *http.Request) {
	pmid := chi.URLParam(r, "pmid")
	if pmid == "" {
		h.writeError(w, http.StatusBadRequest, "pmid is required")
		return
	}

	article, err := h.service.GetArticle(r.Context(), pmid)
	if err != nil {
		h.logger.Error("failed to get article", "pmid", pmid, "error", err)
		h.writeError(w, http.StatusNotFound, "article not found")
		return
	}

	h.writeJSON(w, http.StatusOK, article)
}

// GetStats handles GET /v1/stats requests
func (h *Handler) GetStats(w http.ResponseWriter, r *http.Request) {
	stats, err := h.service.GetStats(r.Context())
	if err != nil {
		h.logger.Error("failed to get stats", "error", err)
		h.writeError(w, http.StatusInternalServerError, "failed to get stats")
		return
	}

	h.writeJSON(w, http.StatusOK, stats)
}

// writeJSON writes a JSON response
func (h *Handler) writeJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(data); err != nil {
		h.logger.Error("failed to encode JSON response", "error", err)
	}
}

// writeError writes an error response
func (h *Handler) writeError(w http.ResponseWriter, status int, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(map[string]string{"error": message})
}
