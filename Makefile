GHACCOUNT := hooklift
NAME := gowsdl
VERSION := v0.5.0

include common.mk

.PHONY: test test-coverage test-integration test-benchmarks lint clean deps update-deps

# Run all tests
test:
	go test -v -race ./...

# Run tests with coverage
test-coverage:
	go test -v -race -coverprofile=coverage.out -covermode=atomic ./...
	go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report generated: coverage.html"
	@go tool cover -func=coverage.out | grep total | awk '{print "Total coverage: " $$3}'

# Run integration tests
test-integration:
	go test -tags=integration -v ./...

# Run benchmarks
test-benchmarks:
	go test -bench=. -benchmem ./...

# Run linter
lint:
	golangci-lint run

# Clean generated files
clean:
	rm -f coverage.out coverage.html

deps:
	go mod download
	go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
	go get github.com/c4milo/github-release
	go get github.com/mitchellh/gox
	go get github.com/hooklift/assert

# Update dependencies
update-deps:
	go get -u ./...
	go mod tidy
