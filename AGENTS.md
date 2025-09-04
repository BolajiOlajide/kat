# Kat Project Agent Configuration

## Build & Test Commands
- Build: `mise run build` or `go build -ldflags "-X github.com/BolajiOlajide/kat/internal/version.version=$(VERSION)" ./cmd/kat`
- Build for Windows: `GOOS=windows GOARCH=amd64 go build -o kat.exe ./cmd/kat`
- Install: `mise run install` or `go install ./...`
- Run: `mise run run` or `go run ./cmd/kat $(ARGS)`
- Run with custom args: `ARGS="your args here" mise run run`
- Run tests: `mise run test` or `go test ./...`
- Run single test: `go test -v ./path/to/package -run TestName`
- Start documentation: `mise run start-doc` or `bundle exec jekyll serve`
- Format code: `mise run fmt`
- Type check: `mise run check`

## Code Style Guidelines
- Indentation: 2 spaces
- Line endings: LF (Unix-style)
- Testing: Use table-driven tests with descriptive names
- Error handling: Use require.NoError(t, err, "descriptive message") in tests
- Imports: Standard Go import organization (stdlib, external, internal)
- Testing framework: Use github.com/stretchr/testify for assertions
- Types: Use descriptive struct names and follow Go naming conventions
- Queries: Use github.com/keegancsmith/sqlf for SQL queries
- Errors: Make use of `errors.New` from `github.com/cockroachdb/errors` for creating errors.
