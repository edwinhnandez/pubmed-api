# PubMed API — Go Technical Challenge

## Overview

A production-ready REST API in Go 1.22+ to search and retrieve PubMed articles, containerized and AWS-ready.

## Features

- **Endpoints:**
  - `GET /healthz` - Health check endpoint
  - `GET /v1/articles` - Search, filter, paginate, and sort articles
  - `GET /v1/articles/{pmid}` - Fetch a single article by PubMed ID
  - `GET /v1/stats` - Get aggregate statistics (top journals, year histogram)

- **Search & Filtering:**
  - Full-text search over title + abstract (case-insensitive)
  - Filter by publication year
  - Filter by journal (exact match)
  - Filter by author (substring match)
  - Pagination (page, page_size, max 50)
  - Sorting (relevance, year_desc, year_asc)

- **Architecture:**
  - Clean layered architecture (domain, repo, service, http, platform)
  - SQLite database with in-memory or file-based storage
  - Repository pattern for data access abstraction
  - Service layer for business logic

- **Configuration:**
  - Environment variable-based configuration
  - Support for S3, local file, or embedded data fallback
  - Structured logging with `log/slog`
  - Graceful shutdown with connection draining

- **Containerization:**
  - Multi-stage Dockerfile
  - Non-root user
  - Health check
  - Small image size

## Quickstart

### Prerequisites

- Go 1.22+
- Docker (optional)
- Make (optional)

### Run Locally

```bash
export PORT=8080
export DATA_PATH=./data/sample_100_pubmed.jsonl
export LOG_LEVEL=info

go run ./cmd/api
```

Or using Make:

```bash
make run
```

### Docker

```bash
# Build image
docker build -t pubmed-api:local .

# Run container
docker run -p 8080:8080 --rm \
  -e DATA_PATH=/app/data/sample_100_pubmed.jsonl \
  -e LOG_LEVEL=info \
  pubmed-api:local
```

Or using Make:

```bash
make docker
make docker-run
```

### Example Requests

```bash
# Health check
curl "http://localhost:8080/healthz"

# Search articles
curl "http://localhost:8080/v1/articles?q=ibuprofen&page=1&page_size=5&sort=relevance"

# Search with filters
curl "http://localhost:8080/v1/articles?q=ibuprofen&year=2020&journal=Medical%20Journal&page=1&page_size=10"

# Get single article
curl "http://localhost:8080/v1/articles/12345678"

# Get statistics
curl "http://localhost:8080/v1/stats"
```

## Environment Variables

| Variable | Description | Default |
|----------|-------------|---------|
| `PORT` | HTTP server port | `8080` |
| `DATA_PATH` | Local JSONL dataset path | `./data/sample_100_pubmed.jsonl` |
| `DATA_S3_URL` | Optional S3 URL to dataset (e.g., `s3://bucket/pubmed.jsonl`) | (empty) |
| `LOG_LEVEL` | Logging level (`debug\|info\|warn\|error`) | `info` |
| `DB_PATH` | SQLite database path (use `:memory:` for in-memory) | `:memory:` |

## Architecture

```
┌─────────────────────────────────────────────────────────┐
│                    HTTP Layer                            │
│  (handlers, routing, middleware, request/response)      │
└────────────────────┬────────────────────────────────────┘
                     │
┌────────────────────▼────────────────────────────────────┐
│                   Service Layer                          │
│  (business logic, validation, search, filters)         │
└────────────────────┬────────────────────────────────────┘
                     │
┌────────────────────▼────────────────────────────────────┐
│                 Repository Layer                          │
│  (data access abstraction, SQLite implementation)       │
└────────────────────┬────────────────────────────────────┘
                     │
┌────────────────────▼────────────────────────────────────┐
│              Platform Layer                              │
│  (config, logging, AWS S3, data loading)                │
└─────────────────────────────────────────────────────────┘
```

### Project Structure

```
.
├── cmd/
│   └── api/
│       └── main.go              # Application entry point
├── internal/
│   ├── domain/                  # Domain entities and DTOs
│   │   └── article.go
│   ├── repo/                    # Repository interfaces and implementations
│   │   ├── article_repository.go
│   │   └── sqlite_repository.go
│   ├── service/                 # Business logic layer
│   │   ├── article_service.go
│   │   └── article_service_test.go
│   └── http/                    # HTTP handlers and routing
│       ├── handlers.go
│       ├── handlers_test.go
│       └── router.go
├── platform/                    # Cross-cutting concerns
│   ├── config.go
│   ├── logger.go
│   ├── aws_s3.go
│   ├── data_loader.go
│   └── embedded_data.go
├── data/                        # Sample data
│   └── sample_100_pubmed.jsonl
├── scripts/                     # Utility scripts
│   └── fetch_pubmed_data.go
├── docs/                        # Documentation
│   └── aws-deploy.md
├── Dockerfile
├── Makefile
├── go.mod
└── README.md
```

## Testing

Run all tests:

```bash
make test
# or
go test ./...
```

The test suite includes:
- Unit tests for service layer (table-driven tests)
- Handler tests for HTTP endpoints
- Mock repositories for isolated testing

## Data Loading

The application supports three data loading strategies (in order of precedence):

1. **S3** - Load from S3 bucket if `DATA_S3_URL` is set
2. **Local File** - Load from local file path if `DATA_PATH` is set and file exists
3. **Embedded** - Use embedded fallback data (via `//go:embed`)

To fetch sample data:

```bash
make fetch-data
```

This will create `data/sample_100_pubmed.jsonl` with ~100 PubMed articles.

## Design Decisions & Tradeoffs

### Database Choice: SQLite

**Why SQLite?**
- Lightweight and portable
- No external dependencies
- Supports both in-memory and file-based storage
- Good performance for small to medium datasets
- Easy to migrate to PostgreSQL later

**Tradeoffs:**
- Not suitable for high-concurrency write workloads
- Limited full-text search capabilities (compared to PostgreSQL's `tsvector`)
- Can be migrated to PostgreSQL with minimal code changes (repository pattern)

### Search Implementation

**Current:** Naive term frequency-based relevance (title matches prioritized)

**Future Enhancements:**
- TF-IDF scoring
- PostgreSQL full-text search (`tsvector`)
- Elasticsearch/OpenSearch for advanced search
- Caching layer (Redis)

### Architecture

- **Clean Layering:** Separates concerns and makes testing easier
- **Repository Pattern:** Allows easy swapping of data sources
- **Interface-based Design:** Enables mocking and testing

## Deployment

See [docs/aws-deploy.md](docs/aws-deploy.md) for AWS deployment guidance (ECS Fargate or App Runner).

## Development

### Build

```bash
make build
```

### Lint

```bash
make lint
```

### Run Tests

```bash
make test
```

## License

This is a technical challenge submission.

