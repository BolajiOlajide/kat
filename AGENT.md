# Kat Project Agent Configuration

## Build & Test Commands
- Build: `go build -ldflags "-X github.com/BolajiOlajide/kat/internal/version.version=$(VERSION)" ./cmd/kat`
- Install: `go install ./...`
- Run: `go run ./cmd/kat $(ARGS)`
- Run with custom args: `make run ARGS="your args here"`
- Run tests: `go test ./...`
- Run single test: `go test -v ./path/to/package -run TestName`
- Start documentation: `bundle exec jekyll serve`

## Code Style Guidelines
- Indentation: 2 spaces
- Line endings: LF (Unix-style)
- Testing: Use table-driven tests with descriptive names
- Error handling: Use require.NoError(t, err, "descriptive message") in tests
- Imports: Standard Go import organization (stdlib, external, internal)
- Testing framework: Use github.com/stretchr/testify for assertions
- Types: Use descriptive struct names and follow Go naming conventions
- Queries: Use github.com/keegancsmith/sqlf for SQL queries
