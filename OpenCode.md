# Sprint Planning - OpenCode Guidelines

## Build Commands
- `make build`: Generate templ templates and build Go application
- `make run`: Run the application
- `make dev-run`: Run the application in development mode
- `make watch` or `npm run dev`: Watch for changes and rebuild
- `npm run build:tailwind`: Build Tailwind CSS

## Test Commands
- `make test`: Run all tests
- `go test ./test/services/room-service_test.go -v`: Run a single test file
- `go test ./test/services -run TestRoomService_CreateRoom -v`: Run a specific test

## Code Style Guidelines
- **Imports**: Group standard library, third-party, and internal imports
- **Formatting**: Use gofmt for Go files
- **Types**: Use strong typing with structs and interfaces
- **Naming**: Use camelCase for variables, PascalCase for exported functions/types
- **Error Handling**: Return errors, use context for cancellation
- **Architecture**: Follow clean architecture with server/service/database layers
- **Components**: Use templ for HTML templating with component-based approach
- **CSS**: Use Tailwind CSS for styling
- **JavaScript**: Use vanilla JS with Web Components