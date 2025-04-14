# CLAUDE.md - Commands and Guidelines for Terraform Provider Uptime Kuma

## Build & Install Commands
- `make build` - Build the provider
- `make install` - Build and install the provider (runs `go install`)
- `go install` - Install the provider binary to `$GOPATH/bin`

## Lint & Format Commands
- `make lint` - Run golangci-lint checks
- `make fmt` - Format code with gofmt
- `make generate` - Generate documentation

## Test Commands
- `make test` - Run all unit tests
- `go test -v ./...` - Run all tests with verbose output
- `go test -v ./internal/client` - Run client tests only
- `go test -v ./internal/provider` - Run provider tests only
- `go test -v ./internal/client -run=TestClientAuthentication` - Run a specific test function
- `make testacc` - Run acceptance tests (creates real resources)
- For testing specific acceptance test: `TF_ACC=1 go test -v -run=TestAccExampleResource_* ./internal/provider`

## Code Style Guidelines
- Go ≥ 1.23 required
- Use Go modules for dependency management (`go get`, `go mod tidy`)
- Follow golang-ci linter rules (errcheck, staticcheck, ineffassign, etc.)
- Use gofmt for formatting (-s -w -e flags)
- Use Terraform Plugin Framework patterns for resources/data sources
- Error handling: Use framework.Diagnostics for user-facing errors
- Testing: Write both unit tests and acceptance tests

Documentation is generated via `make generate` and resides in the `docs/` directory.