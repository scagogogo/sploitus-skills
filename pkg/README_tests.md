# Unit Tests for Sploitus Crawler

This directory contains unit tests for the various components of the Sploitus Crawler package.

## Test Structure

The tests are organized according to the package structure:

- `pkg/sploitus/client_test.go`: Tests for the Sploitus API client
- `pkg/sploitus/utils_test.go`: Tests for utility functions
- `pkg/types/types_test.go`: Tests for type definitions and JSON serialization

## Running Tests

You can run all the tests using the Go test command:

```bash
# Run all tests
go test ./...

# Run tests with verbose output
go test -v ./...

# Run tests for a specific package
go test -v ./pkg/sploitus
go test -v ./pkg/types

# Run a specific test
go test -v ./pkg/sploitus -run TestNewClient
```

## Test Coverage

To generate test coverage reports:

```bash
# Generate coverage report
go test ./... -coverprofile=coverage.out

# View coverage in browser
go tool cover -html=coverage.out

# View coverage in terminal
go tool cover -func=coverage.out
```

## Adding New Tests

When adding new features to the codebase, please also add corresponding tests. Follow the existing test patterns:

1. Test basic functionality
2. Test edge cases
3. Test error handling

Each test function should be named `TestXxx` where `Xxx` is the name of the functionality being tested.

## Mocking HTTP Requests

The client tests use Go's `httptest` package to mock HTTP requests and avoid actual API calls during testing. This allows the tests to be run in any environment without requiring actual connectivity to Sploitus API. 