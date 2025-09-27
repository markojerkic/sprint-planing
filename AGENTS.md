# Agent Guidelines for Sprint Planning

## Build/Test Commands
- `make build` - Build the application (runs templ generate + Go build)
- `make test` - Run all tests with verbose output (`go test ./... -v`)
- `go test ./test/services/room-service_test.go -v` - Run single test file
- `make watch` - Start development server with live reload (starts DB + bun dev)
- `bun dev` - Run frontend development with parallel processes (templ, tailwind, go)

## Code Style Guidelines

### Go
- Use `gorm.Model` for database models with proper struct tags
- Package imports: stdlib first, then external, then internal (separated by blank lines)
- Error handling: return errors explicitly, use `log/slog` for structured logging
- Naming: PascalCase for exported, camelCase for unexported, use descriptive names
- Use testcontainers for integration tests with proper cleanup in `TearDownSubTest`

### Frontend (JS/TS)
- ES6+ target, use TypeScript checking (`checkJs: true`)
- Use kebab-case for file names, camelCase for variables
- HTMX for interactivity, vanilla JS for components

### Imports
- Go: group stdlib, external, internal packages with blank line separation
- JS/TS: ES6 modules, avoid CommonJS