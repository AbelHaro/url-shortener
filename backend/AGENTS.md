# AGENTS.md

This file contains guidelines and commands for agentic coding agents working in this Go URL shortener repository.

## Build/Test Commands

### Running Tests
- Run all tests: `go test ./...`
- Run tests in specific package: `go test ./internal/service/url`
- Run single test function: `go test ./internal/service/url -run TestService_Store`
- Run with verbose output: `go test -v ./...`
- Run with coverage: `go test -cover ./...`

### Build Commands
- Build the application: `go build ./cmd/api`
- Build and run: `go run ./cmd/api`
- Build with race detection: `go build -race ./cmd/api`

### Linting and Formatting
- Format code: `go fmt ./...`
- Run vet: `go vet ./...`
- Run staticcheck (if installed): `staticcheck ./...`

## Code Style Guidelines

### Package Structure
- Domain-driven design with clear separation: `domain/`, `repository/`, `service/`, `delivery/http/`
- Each entity has its own subdirectory under `repository/` and `service/`
- Repository interfaces in `repository/{entity}/repository.go`
- Repository implementations in `repository/{entity}/postgres.go`
- Mocks in `repository/{entity}/mock.go`
- Services in `service/{entity}/service.go`

### File Naming Conventions
- Use `repository.go` for repository interfaces
- Use `postgres.go` for PostgreSQL implementations
- Use `mock.go` for mock implementations
- Use `service.go` for service implementations
- Use `*_test.go` for test files
- File name should match the type or be descriptive of its purpose

### Type Naming Conventions
- Repository interface: `Repository` (not `URLRepository`)
- PostgreSQL implementation: `PostgresRepository` (not `PostgresURLRepository`)
- Mock implementation: `MockRepository` (not `MockURLRepository`)
- Service: `Service` (not `URLService`)

### Import Organization
- Group imports in three sections: standard library, third-party, internal packages
- Use blank lines between groups
- Use import aliases when there's a naming conflict (e.g., `urlRepo "github.com/.../repository/url"`)

### Naming Conventions
- Use PascalCase for exported types, functions, and constants
- Use camelCase for unexported identifiers
- Repository interfaces: `Repository` with methods like `Store()`, `FindByShortURL()`
- Service structs: `Service` with methods like `Store()`, `FindByShortURL()`
- Error variables in domain: `ErrURLNotFound`, `ErrInvalidURL`

### Error Handling
- Define domain-specific errors in `domain/errors.go`
- Use `errors.Is()` for error comparison in handlers
- Return errors from service layer, handle in delivery layer
- Use `domain.ErrInternal` for unexpected errors

### Domain Models
- Use UUIDs for entity IDs with `github.com/google/uuid`
- Struct tags for JSON and GORM: `json:"id" gorm:"primaryKey;type:uuid;default:uuid_generate_v4()"`
- Time fields with GORM defaults: `gorm:"not null;default:now()"`

### HTTP Handlers
- Use Gin framework with structured error responses
- Return appropriate HTTP status codes (201 for creation, 404 for not found, 400 for bad requests)
- Use request/response DTOs in `delivery/http/dto.go`
- Include Swagger annotations for API documentation

### Testing
- Use table-driven tests for multiple scenarios
- Mock repositories using custom mock implementations
- Test both success and error cases
- Use `t.Fatalf()` for setup errors, `t.Errorf()` for test failures

### Database
- Use GORM for ORM with PostgreSQL driver
- Use GORM generics: `gorm.G[domain.Model](db).First(ctx)`
- Repository pattern with interfaces for testability
- Use migrations or schema definitions in models

### Configuration
- Use `github.com/joho/godotenv` for environment variables
- Load config in `server/app.go` before initializing services
- Use `.env` file with example in `env.example`

### Application Structure
- Main entry point in `cmd/api/main.go`
- Application bootstrap in `server/app.go`
- Dependency injection in `server.NewApp()`
- Use `log.Fatalf()` for unrecoverable startup errors

### Development Workflow
- Run `go fmt ./...` before committing
- Run `go vet ./...` to check for issues
- Run relevant tests after making changes
- Use `go test -v ./internal/service/url` for focused testing

### API Design
- RESTful endpoints with version prefix `/api/v1/`
- Use appropriate HTTP methods (GET, POST, DELETE)
- Return JSON responses with consistent structure
- Include health check endpoint `/health`

### Security
- Validate URLs using `net/url.ParseRequestURI()`
- Use UUIDs to prevent enumeration attacks
- Handle errors without exposing internal details
- Use HTTPS in production (not configured here)

### Performance
- Use connection pooling with GORM
- Consider caching for frequently accessed URLs
- Use appropriate indexes in database (e.g., unique constraint on short_url)

### Documentation
- Use Go doc comments for exported functions
- Include Swagger annotations in HTTP handlers
- Keep README.md updated with setup instructions
