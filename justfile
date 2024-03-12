default: lint test

# Run tests
test:
    go test ./...

# Run golangci-lint
lint:
    golangci-lint run
