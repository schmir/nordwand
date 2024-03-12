default:
    just --list

# Run tests
test:
    go test ./...

# Run golangci-lint
lint:
    golangci-lint run

# Run the main program
run:
    go run ./
