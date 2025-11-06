package domain

// Article represents a PubMed article entity
type Article struct {
	PMID      string   `json:"pmid"`
	Title     string   `json:"title"`
	Abstract  string   `json:"abstract"`
	Authors   []string `json:"authors"`
	Journal   string   `json:"journal"`
	PubYear   int      `json:"pub_year"`
	MeshTerms []string `json:"mesh_terms"`
	DOI       string   `json:"doi,omitempty"`
}

// SearchFilters represents search and filter parameters
type SearchFilters struct {
	Query    string
	Year     *int
	Journal  string
	Author   string
	Page     int
	PageSize int
	Sort     string
}

// SearchResult represents paginated search results
type SearchResult struct {
	Items    []*Article `json:"items"`
	Page     int        `json:"page"`
	PageSize int        `json:"page_size"`
	Total    int        `json:"total"`
	TookMs   int64      `json:"took_ms"`
}

// Stats represents aggregate statistics
type Stats struct {
	TopJournals   []JournalCount `json:"top_journals"`
	YearHistogram map[int]int    `json:"year_histogram"`
}

// JournalCount represents journal count statistics
type JournalCount struct {
	Journal string `json:"journal"`
	Count   int    `json:"count"`
}
