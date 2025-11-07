# Architecture & Project Structure Explanation

## Project Structure

```
pubmed-api/
├── cmd/
│   └── api/
│       └── main.go           # Application entry point
├── internal/
│   ├── domain/               # Domain entities and DTOs
│   ├── repo/                 # Repository layer (data access)
│   ├── service/              # Business logic layer
│   ├── http/                 # HTTP handlers and routing
│   └── platform/             # Cross-cutting concerns
├── data/                     # Sample data files
├── api/                      # API specifications (OpenAPI)
├── docs/                     # Documentation
├── terraform/                # Infrastructure as Code
├── Dockerfile                # Container definition
├── Makefile                  # Build automation
├── go.mod                    # Go module definition
└── README.md                 # Project documentation
```

---

## Structure Rationale

### 1. `cmd/` - Command Applications

**Purpose**: Contains the main entry points for different applications in the project.

```
cmd/
└── api/
    └── main.go
```

**Why this structure?**
- **Separation of concerns**: Keeps application entry points separate from library code
- **Go standard**: Follows Go's recommended project layout
- **Multi-binary support**: Easy to add more binaries (e.g., `cmd/migrate`, `cmd/worker`)
- **Build isolation**: Each binary can have its own build configuration
- **Clean imports**: Application code imports from `internal/`, not vice versa

**Considerations**:
- `main.go` is the single entry point that wires everything together
- It's responsible for:
  - Configuration loading
  - Dependency injection
  - Server initialization
  - Graceful shutdown handling

---

### 2. `internal/` - Private Application Code

**Purpose**: Contains all application code that should not be imported by other projects.

**Why `internal/`?**
- **Go visibility rule**: Packages in `internal/` can only be imported by packages within the same parent directory
- **Encapsulation**: Prevents external projects from depending on internal implementation details
- **API stability**: Only exported packages should be imported by external code
- **Refactoring safety**: Internal code can be changed without breaking external dependencies

#### 2.1. `internal/domain/` - Domain Layer

```
internal/domain/
└── article.go
```

**Purpose**: Contains domain entities, value objects, and DTOs.

**Why separate?**
- **Clean Architecture**: Domain layer is the core and should have no dependencies
- **Business rules**: Domain models represent business concepts
- **Reusability**: Domain models are used across all layers
- **Testability**: Pure Go structs, easy to test

**Considerations**:
- Domain models should be independent of frameworks
- No external dependencies (except standard library)
- Contains business types, not infrastructure concerns

#### 2.2. `internal/repo/` - Repository Layer

```
internal/repo/
├── article_repository.go     # Interface definition
└── sqlite_repository.go      # SQLite implementation
```

**Purpose**: Abstracts data access logic.

**Why Repository Pattern?**
- **Dependency Inversion**: Business logic depends on abstractions, not concrete implementations
- **Testability**: Easy to mock repositories for unit testing
- **Flexibility**: Can swap implementations (SQLite → PostgreSQL → MongoDB) without changing business logic
- **Separation**: Data access logic is isolated from business logic

**Considerations**:
- Interface defines the contract (`ArticleRepository`)
- Implementation handles database-specific details
- Service layer depends on interface, not implementation
- Easy to add new implementations (e.g., `postgres_repository.go`, `s3_repository.go`)

#### 2.3. `internal/service/` - Service Layer (Business Logic)

```
internal/service/
├── article_service.go        # Business logic
└── article_service_test.go  # Unit tests
```

**Purpose**: Contains business logic and orchestration.

**Why Service Layer?**
- **Business rules**: Centralizes business logic
- **Coordination**: Orchestrates multiple repository calls
- **Validation**: Validates input and business rules
- **Transaction boundaries**: Manages transaction scope (if needed)

**Considerations**:
- Service layer depends on repository interfaces
- Contains business logic, not HTTP concerns
- Testable in isolation with mock repositories
- Can be reused by different interfaces (HTTP, gRPC, CLI)

#### 2.4. `internal/http/` - HTTP Layer

```
internal/http/
├── handlers.go              # HTTP handlers
├── handlers_test.go         # Handler tests
├── router.go                # Route definitions
└── service_interface.go     # Interface for testing
```

**Purpose**: Handles HTTP-specific concerns.

**Why separate HTTP layer?**
- **Framework independence**: HTTP layer can be swapped (Chi → Gin → Echo)
- **Thin controllers**: Handlers are thin, delegate to service layer
- **HTTP concerns**: Request/response transformation, status codes
- **Middleware**: Request logging, authentication, rate limiting

**Considerations**:
- Handlers are thin, delegate to service layer
- HTTP layer depends on service interface (not concrete type)
- Easy to add middleware (auth, logging, metrics)
- Can be tested independently with mock services

#### 2.5. `internal/platform/` - Platform/Infrastructure Layer

```
internal/platform/
├── config.go                # Configuration management
├── logger.go                # Logging setup
├── aws_s3.go                # AWS S3 integration
├── data_loader.go           # Data loading logic
└── embedded_data.go        # Embedded fallback data
```

**Purpose**: Contains infrastructure and cross-cutting concerns.

**Why Platform Layer?**
- **Cross-cutting concerns**: Configuration, logging, AWS integration
- **Infrastructure**: External service integrations
- **Reusability**: Shared utilities across layers
- **Environment-specific**: Different implementations for dev/prod

**Considerations**:
- Can depend on external libraries (AWS SDK, etc.)
- Contains infrastructure-specific code
- Provides utilities for other layers
- Can be swapped for different environments

---

### 3. `data/` - Data Files

```
data/
└── sample_100_pubmed.jsonl
```

**Purpose**: Contains sample data and test fixtures.

**Why separate?**
- **Version control**: Data files can be tracked separately
- **Testing**: Easy to reference test data
- **Documentation**: Sample data serves as documentation
- **Embedding**: Can be embedded in binary for fallback

**Considerations**:
- Large files should be in `.gitignore` or use Git LFS
- Sample data should be representative
- Can be loaded from S3 in production

---

### 4. `api/` - API Specifications

```
api/
└── openapi.yaml
```

**Purpose**: Contains API specifications (OpenAPI).

**Why separate?**
- **API contract**: Documents the API contract
- **Code generation**: Can generate client SDKs
- **Documentation**: Serves as API documentation
- **Testing**: Can validate API against spec

**Considerations**:
- OpenAPI spec is the source of truth
- Can be used to generate documentation
- Can validate requests/responses
- Future: Add GraphQL schema here

---

### 5. `docs/` - Documentation

```
docs/
├── aws-deploy.md            # AWS deployment guide
└── ARCHITECTURE.md          # This file
```

**Purpose**: Contains project documentation.

**Why separate?**
- **Organization**: Keeps documentation organized
- **Accessibility**: Easy to find and reference
- **Versioning**: Documentation can evolve with code
- **Multiple formats**: Can include diagrams, guides, etc.

**Considerations**:
- Documentation should be kept up-to-date
- Can include architecture diagrams
- Deployment guides, runbooks, etc.

---

### 6. `terraform/` - Infrastructure as Code

```
terraform/
├── main.tf                  # Main resources
├── variables.tf             # Input variables
├── outputs.tf              # Output values
├── README.md               # Terraform documentation
└── .gitignore              # Terraform ignores
```

**Purpose**: Defines AWS infrastructure declaratively.

**Why Terraform?**
- **Infrastructure as Code**: Version-controlled infrastructure
- **Reproducibility**: Same infrastructure across environments
- **Collaboration**: Team can review infrastructure changes
- **State management**: Terraform tracks resource state

**Considerations**:
- Currently includes ECR and S3 (minimal stub)
- Can be extended with ECS, ALB, CloudWatch, etc.
- Should use remote state (S3 backend) in production
- Variables allow environment-specific configurations

---

### 7. `Dockerfile` - Container Definition

**Purpose**: Defines how to build the application container.

**Why separate?**
- **Containerization**: Makes application portable
- **Multi-stage builds**: Smaller final image
- **Security**: Non-root user, minimal base image
- **Reproducibility**: Same build across environments

**Considerations**:
- Multi-stage build reduces image size
- Non-root user for security
- Health check for orchestration
- Can be optimized for caching

---

### 8. `Makefile` - Build Automation

**Purpose**: Provides common commands for development and deployment.

**Why Makefile?**
- **Standardization**: Common commands across team
- **Documentation**: Commands are self-documenting
- **Automation**: Reduces repetitive tasks
- **Consistency**: Same commands across environments

**Considerations**:
- `make run` - Local development
- `make test` - Run tests
- `make docker` - Build container
- `make lint` - Code quality checks

---

## Architecture Principles

### 1. Clean Architecture / Layered Architecture

```
┌─────────────────────────────────────┐
│      HTTP Layer (internal/http)     │  ← Handles HTTP requests
└──────────────┬──────────────────────┘
               │
┌──────────────▼──────────────────────┐
│   Service Layer (internal/service)   │  ← Business logic
└──────────────┬──────────────────────┘
               │
┌──────────────▼──────────────────────┐
│  Repository Layer (internal/repo)   │  ← Data access
└──────────────┬──────────────────────┘
               │
┌──────────────▼──────────────────────┐
│   Domain Layer (internal/domain)     │  ← Core entities
└───────────────────────────────────────┘
```

**Dependency Flow**: Outer layers depend on inner layers, not vice versa.

### 2. Dependency Inversion Principle

- **Service layer** depends on **repository interface**, not implementation
- **HTTP layer** depends on **service interface**, not concrete type
- Allows easy mocking and testing

### 3. Single Responsibility Principle

- Each layer has a single, well-defined responsibility
- Domain: Business entities
- Repository: Data access
- Service: Business logic
- HTTP: Request/response handling

### 4. Interface Segregation

- Small, focused interfaces
- `ArticleRepository` interface defines only what's needed
- Easy to implement and mock

---

## Design Decisions & Tradeoffs

### 1. Why SQLite Instead of PostgreSQL?

**Decision**: Use SQLite for simplicity within timebox.

**Tradeoffs**:
- **Pros**: No external dependencies, easy setup, good for small datasets
- **Cons**: Not suitable for high concurrency, limited full-text search

**Migration Path**: Repository pattern allows easy migration to PostgreSQL later.

### 2. Why In-Memory Database by Default?

**Decision**: Use `:memory:` for development, file-based for production.

**Tradeoffs**:
- **Pros**: Fast, no disk I/O, good for testing
- **Cons**: Data lost on restart

**Production**: Use file-based SQLite or migrate to PostgreSQL.

### 3. Why Repository Pattern?

**Decision**: Abstract data access behind interfaces.

**Tradeoffs**:
- **Pros**: Testable, flexible, easy to swap implementations
- **Cons**: Additional abstraction layer, more code

**Benefit**: Can easily add PostgreSQL, MongoDB, or S3-backed repository.

### 4. Why Chi Router Instead of Gin/Echo?

**Decision**: Use Chi for lightweight, standard library approach.

**Tradeoffs**:
- **Pros**: Lightweight, standard library compatible, minimal dependencies
- **Cons**: Less features than Gin/Echo, more manual setup

**Alternative**: Can easily swap to Gin or Echo if needed.

### 5. Why Not GraphQL or gRPC?

**Decision**: Start with REST API, can add later.

**Tradeoffs**:
- **Pros**: Simple, widely understood, easy to test
- **Cons**: Less flexible than GraphQL, less efficient than gRPC

**Future**: Can add GraphQL endpoint alongside REST.

---

## Scalability Considerations

### Current Architecture (Suitable for ~100 articles)

- SQLite in-memory or file-based
- Single binary deployment
- No caching layer
- Simple search implementation

### Future Enhancements (For Production Scale)

1. **Database Migration**:
   - Move to PostgreSQL with full-text search (`tsvector`)
   - Add read replicas for scaling reads
   - Connection pooling

2. **Caching Layer**:
   - Add Redis for query caching
   - Cache frequently accessed articles
   - Cache statistics

3. **Search Enhancement**:
   - Move to Elasticsearch/OpenSearch for advanced search
   - Implement TF-IDF ranking
   - Add faceted search

4. **API Evolution**:
   - Add GraphQL endpoint
   - Add gRPC for internal services
   - Version API (`/v1/`, `/v2/`)

5. **Infrastructure**:
   - ECS Fargate with auto-scaling
   - Application Load Balancer
   - CloudFront CDN
   - CloudWatch monitoring

---

## Testing Strategy

### Unit Tests
- **Service layer**: Test business logic with mock repositories
- **Repository layer**: Test data access (can use in-memory SQLite)
- **HTTP layer**: Test handlers with mock services

### Integration Tests
- Test full stack with real database
- Test API endpoints end-to-end

### Test Structure
```
internal/
├── service/
│   └── article_service_test.go  # Unit tests
└── http/
    └── handlers_test.go         # Handler tests
```

---

## Security Considerations

1. **Input Validation**: All inputs validated in service layer
2. **SQL Injection**: Using parameterized queries (SQLite driver handles this)
3. **Error Handling**: Don't expose internal errors to clients
4. **Container Security**: Non-root user in Docker
5. **AWS Security**: IAM roles for S3 access, no hardcoded credentials

---

## Deployment Considerations

### Local Development
- Run with `make run`
- Uses local file for data
- In-memory database

### Docker
- Multi-stage build
- Non-root user
- Health check configured

### AWS Deployment
- ECS Fargate or App Runner
- S3 for data storage
- CloudWatch for logs
- ALB for load balancing

---

## Conclusion

This structure provides:
- **Clean separation of concerns**
- **Easy testing and mocking**
- **Flexibility to swap implementations**
- **Scalability path for production**
- **AWS-ready deployment**
- **Maintainable and understandable code**

The architecture follows Go best practices and clean architecture principles, making it suitable for both development and production use.

