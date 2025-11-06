package repo

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log/slog"
	"pubmed-api/internal/domain"
	"strings"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

// SQLiteRepository implements ArticleRepository using SQLite
type SQLiteRepository struct {
	db     *sql.DB
	logger *slog.Logger
}

// Assert SQLiteRepository implements ArticleRepository
var _ ArticleRepository = (*SQLiteRepository)(nil)

// NewSQLiteRepository creates a new SQLite repository
func NewSQLiteRepository(dbPath string, logger *slog.Logger) (*SQLiteRepository, error) {
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	repo := &SQLiteRepository{
		db:     db,
		logger: logger,
	}

	if err := repo.initSchema(); err != nil {
		return nil, fmt.Errorf("failed to init schema: %w", err)
	}

	return repo, nil
}

// initSchema creates the articles table if it doesn't exist
func (r *SQLiteRepository) initSchema() error {
	query := `
	CREATE TABLE IF NOT EXISTS articles (
		pmid TEXT PRIMARY KEY,
		title TEXT NOT NULL,
		abstract TEXT,
		authors TEXT NOT NULL,
		journal TEXT NOT NULL,
		pub_year INTEGER,
		mesh_terms TEXT,
		doi TEXT,
		search_text TEXT
	);

	CREATE INDEX IF NOT EXISTS idx_search_text ON articles(search_text);
	CREATE INDEX IF NOT EXISTS idx_pub_year ON articles(pub_year);
	CREATE INDEX IF NOT EXISTS idx_journal ON articles(journal);
	`

	if _, err := r.db.Exec(query); err != nil {
		return fmt.Errorf("failed to create schema: %w", err)
	}

	return nil
}

// LoadData loads articles from a JSONL file
func (r *SQLiteRepository) LoadData(ctx context.Context, dataPath string) error {
	// This will be called from the platform layer that handles S3/local/embedded loading
	// For now, we'll implement a method to insert articles
	return nil
}

// InsertArticles inserts articles into the database
func (r *SQLiteRepository) InsertArticles(ctx context.Context, articles []*domain.Article) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	stmt, err := tx.PrepareContext(ctx, `
		INSERT OR REPLACE INTO articles (pmid, title, abstract, authors, journal, pub_year, mesh_terms, doi, search_text)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
	`)
	if err != nil {
		return fmt.Errorf("failed to prepare statement: %w", err)
	}
	defer stmt.Close()

	for _, article := range articles {
		authorsJSON, _ := json.Marshal(article.Authors)
		meshTermsJSON, _ := json.Marshal(article.MeshTerms)
		searchText := strings.ToLower(article.Title + " " + article.Abstract)

		_, err := stmt.ExecContext(ctx,
			article.PMID,
			article.Title,
			article.Abstract,
			string(authorsJSON),
			article.Journal,
			article.PubYear,
			string(meshTermsJSON),
			article.DOI,
			searchText,
		)
		if err != nil {
			return fmt.Errorf("failed to insert article %s: %w", article.PMID, err)
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	r.logger.Info("loaded articles", "count", len(articles))
	return nil
}

// FindByID retrieves an article by its PubMed ID
func (r *SQLiteRepository) FindByID(ctx context.Context, pmid string) (*domain.Article, error) {
	query := `SELECT pmid, title, abstract, authors, journal, pub_year, mesh_terms, doi
		FROM articles WHERE pmid = ?`

	var article domain.Article
	var authorsJSON, meshTermsJSON string

	err := r.db.QueryRowContext(ctx, query, pmid).Scan(
		&article.PMID,
		&article.Title,
		&article.Abstract,
		&authorsJSON,
		&article.Journal,
		&article.PubYear,
		&meshTermsJSON,
		&article.DOI,
	)
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("article not found: %s", pmid)
	}
	if err != nil {
		return nil, fmt.Errorf("failed to query article: %w", err)
	}

	if err := json.Unmarshal([]byte(authorsJSON), &article.Authors); err != nil {
		return nil, fmt.Errorf("failed to unmarshal authors: %w", err)
	}

	if err := json.Unmarshal([]byte(meshTermsJSON), &article.MeshTerms); err != nil {
		return nil, fmt.Errorf("failed to unmarshal mesh terms: %w", err)
	}

	return &article, nil
}

// Search performs a search with filters and pagination
func (r *SQLiteRepository) Search(ctx context.Context, filters *domain.SearchFilters) (*domain.SearchResult, error) {
	startTime := time.Now()

	// Build WHERE clause
	whereClauses := []string{}
	args := []interface{}{}

	if filters.Query != "" {
		whereClauses = append(whereClauses, "search_text LIKE ?")
		args = append(args, "%"+strings.ToLower(filters.Query)+"%")
	}

	if filters.Year != nil {
		whereClauses = append(whereClauses, "pub_year = ?")
		args = append(args, *filters.Year)
	}

	if filters.Journal != "" {
		whereClauses = append(whereClauses, "journal = ?")
		args = append(args, filters.Journal)
	}

	if filters.Author != "" {
		whereClauses = append(whereClauses, "authors LIKE ?")
		args = append(args, "%"+filters.Author+"%")
	}

	whereClause := ""
	if len(whereClauses) > 0 {
		whereClause = "WHERE " + strings.Join(whereClauses, " AND ")
	}

	// Get total count
	countQuery := "SELECT COUNT(*) FROM articles " + whereClause
	var total int
	if err := r.db.QueryRowContext(ctx, countQuery, args...).Scan(&total); err != nil {
		return nil, fmt.Errorf("failed to count articles: %w", err)
	}

	// Build ORDER BY clause
	orderBy := "pmid ASC"
	switch filters.Sort {
	case "year_desc":
		orderBy = "pub_year DESC, pmid ASC"
	case "year_asc":
		orderBy = "pub_year ASC, pmid ASC"
	case "relevance":
		// Naive relevance: prioritize articles where query appears in title
		if filters.Query != "" {
			orderBy = fmt.Sprintf("CASE WHEN title LIKE '%%%s%%' THEN 1 ELSE 2 END, pmid ASC", strings.ToLower(filters.Query))
		}
	}

	// Build pagination
	offset := (filters.Page - 1) * filters.PageSize
	limit := filters.PageSize

	query := fmt.Sprintf(`
		SELECT pmid, title, abstract, authors, journal, pub_year, mesh_terms, doi
		FROM articles %s ORDER BY %s LIMIT ? OFFSET ?
	`, whereClause, orderBy)

	args = append(args, limit, offset)

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to query articles: %w", err)
	}
	defer rows.Close()

	var articles []*domain.Article
	for rows.Next() {
		var article domain.Article
		var authorsJSON, meshTermsJSON string

		if err := rows.Scan(
			&article.PMID,
			&article.Title,
			&article.Abstract,
			&authorsJSON,
			&article.Journal,
			&article.PubYear,
			&meshTermsJSON,
			&article.DOI,
		); err != nil {
			return nil, fmt.Errorf("failed to scan article: %w", err)
		}

		if err := json.Unmarshal([]byte(authorsJSON), &article.Authors); err != nil {
			return nil, fmt.Errorf("failed to unmarshal authors: %w", err)
		}

		if err := json.Unmarshal([]byte(meshTermsJSON), &article.MeshTerms); err != nil {
			return nil, fmt.Errorf("failed to unmarshal mesh terms: %w", err)
		}

		articles = append(articles, &article)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("failed to iterate rows: %w", err)
	}

	tookMs := time.Since(startTime).Milliseconds()

	return &domain.SearchResult{
		Items:    articles,
		Page:     filters.Page,
		PageSize: filters.PageSize,
		Total:    total,
		TookMs:   tookMs,
	}, nil
}

// GetStats returns aggregate statistics
func (r *SQLiteRepository) GetStats(ctx context.Context) (*domain.Stats, error) {
	// Top journals
	journalQuery := `
		SELECT journal, COUNT(*) as count
		FROM articles
		GROUP BY journal
		ORDER BY count DESC
		LIMIT 5
	`

	rows, err := r.db.QueryContext(ctx, journalQuery)
	if err != nil {
		return nil, fmt.Errorf("failed to query top journals: %w", err)
	}
	defer rows.Close()

	var topJournals []domain.JournalCount
	for rows.Next() {
		var jc domain.JournalCount
		if err := rows.Scan(&jc.Journal, &jc.Count); err != nil {
			return nil, fmt.Errorf("failed to scan journal count: %w", err)
		}
		topJournals = append(topJournals, jc)
	}

	// Year histogram
	yearQuery := `
		SELECT pub_year, COUNT(*) as count
		FROM articles
		WHERE pub_year IS NOT NULL
		GROUP BY pub_year
		ORDER BY pub_year
	`

	rows, err = r.db.QueryContext(ctx, yearQuery)
	if err != nil {
		return nil, fmt.Errorf("failed to query year histogram: %w", err)
	}
	defer rows.Close()

	yearHistogram := make(map[int]int)
	for rows.Next() {
		var year, count int
		if err := rows.Scan(&year, &count); err != nil {
			return nil, fmt.Errorf("failed to scan year count: %w", err)
		}
		yearHistogram[year] = count
	}

	return &domain.Stats{
		TopJournals:   topJournals,
		YearHistogram: yearHistogram,
	}, nil
}

// Close closes the database connection
func (r *SQLiteRepository) Close() error {
	return r.db.Close()
}

