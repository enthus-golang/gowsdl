# Claude Development Guidelines for gowsdl

This document provides development guidelines and codebase information for AI assistants working on the gowsdl project.

## Project Overview

gowsdl is a WSDL-to-Go code generator that creates idiomatic Go code from WSDL (Web Services Description Language) files. It supports both WSDL 1.1 and WSDL 2.0, with modern Go features including generics support.

## Architecture Overview

### Package Structure

The project follows the standard Go project layout:

- **`cmd/gowsdl/`**: Command-line interface entry point
- **`pkg/`**: Core library packages
  - `core/`: Core utilities, error handling, namespace management
  - `generator/`: Code generation logic and templates
  - `http/`: HTTP client with security features
  - `parser/`: WSDL and XSD parsers
  - `soap/`: SOAP protocol implementation
  - `types/`: Type mapping and conversion
  - `utils/`: General utility functions
- **`soap/`**: Legacy SOAP package (maintained for backward compatibility)
- **`fixtures/`**: Test WSDL/XSD files
- **`example/`**: Usage examples
- **`docs/`**: Documentation

### Key Components

1. **Parser Module** (`pkg/parser/`)
   - WSDL 1.1 parser
   - WSDL 2.0 parser
   - XSD schema parser
   - Version detection logic

2. **Generator Module** (`pkg/generator/`)
   - Template-based code generation
   - Support for generics
   - Server-side code generation
   - Namespace handling

3. **HTTP Client** (`pkg/http/`)
   - Configurable timeouts and retries
   - TLS/SSL configuration
   - Proxy support
   - Rate limiting

4. **SOAP Implementation** (`pkg/soap/`)
   - Client implementation
   - Authentication mechanisms
   - Encoding/decoding utilities
   - Type definitions

## Development Practices

### Code Style

1. **Follow Go idioms**: Use standard Go conventions and patterns
2. **Error handling**: Use wrapped errors with context (`fmt.Errorf` with `%w`)
3. **Naming**: Use descriptive names, follow Go naming conventions
4. **Comments**: Add godoc comments for all exported types and functions
5. **No unnecessary comments**: Don't add inline comments unless absolutely necessary

### Security Considerations

1. **Path validation**: Always validate file paths to prevent directory traversal
2. **Input sanitization**: Validate package names and other user inputs
3. **TLS defaults**: Use secure defaults for HTTPS connections
4. **No hardcoded secrets**: Never include API keys or credentials in code

### Testing Requirements

1. **Unit tests**: Write tests for all new functionality
2. **Coverage target**: Maintain 80%+ code coverage
3. **Integration tests**: Add integration tests for end-to-end scenarios
4. **Test naming**: Use descriptive test names (e.g., `TestParser_WSDL20_ComplexTypes`)

### Build and Test Commands

```bash
# Build the project
go build ./cmd/gowsdl

# Run all tests
go test -v -race ./...

# Run tests with coverage
go test -v -race -coverprofile=coverage.out -covermode=atomic ./...
go tool cover -html=coverage.out -o coverage.html
go tool cover -func=coverage.out | grep total

# Run integration tests
go test -tags=integration -v ./...

# Run benchmarks
go test -bench=. -benchmem ./...

# Run linter (install: go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest)
golangci-lint run

# Update dependencies
go get -u ./...
go mod tidy

# Install the tool locally
go install ./cmd/gowsdl
```

## Common Development Tasks

### Adding a New Feature

1. Create feature branch from `main`
2. Add tests first (TDD approach)
3. Implement the feature
4. Ensure all tests pass
5. Run linter and fix any issues
6. Update documentation if needed

### Modifying Templates

Templates are located in `pkg/generator/templates/`. When modifying:

1. Understand the existing template structure
2. Make changes incrementally
3. Test with various WSDL files
4. Ensure backward compatibility

### Working with WSDL Versions

The codebase supports both WSDL 1.1 and 2.0:

- Version detection: `pkg/parser/version.go`
- WSDL 1.1 parser: `pkg/parser/wsdl11.go`
- WSDL 2.0 parser: `pkg/parser/wsdl20.go`

### Error Handling Pattern

```go
// Use custom error types for specific errors
return &core.ValidationError{
    Field: "packageName",
    Value: pkg,
    Message: "invalid package name",
}

// Wrap errors with context
if err != nil {
    return fmt.Errorf("failed to parse WSDL: %w", err)
}
```

## Important Files

- `cmd/gowsdl/main.go`: CLI entry point and flag handling
- `pkg/generator/generator.go`: Main code generation logic
- `pkg/parser/parser.go`: WSDL parsing interface
- `pkg/http/client.go`: HTTP client configuration
- `.github/workflows/`: CI/CD configuration

## Dependencies

- Go 1.21+ required
- Minimal external dependencies
- Main testing dependency: `github.com/stretchr/testify`

## CI/CD

- GitHub Actions workflow: `.github/workflows/`
- Automated testing on pull requests
- Code coverage reporting via codecov
- Release automation

## Debugging Tips

1. Use verbose logging during development
2. Test with various WSDL files from `fixtures/`
3. Check generated code compiles and runs correctly
4. Use `go test -v` for detailed test output

## Performance Considerations

1. Minimize memory allocations in hot paths
2. Use buffered I/O for file operations
3. Implement caching where appropriate
4. Run benchmarks to verify performance

## Backward Compatibility

Always maintain backward compatibility:

1. Don't change existing public APIs
2. Deprecate features instead of removing
3. Keep legacy `soap/` package functional
4. Test with existing user code

## Release Process

1. Update version in code/tags
2. Update CHANGELOG
3. Create git tag
4. GitHub Actions handles the rest

## Common Issues and Solutions

### Issue: Generated code doesn't compile
- Check for namespace conflicts
- Verify type mappings are correct
- Ensure templates generate valid Go code

### Issue: WSDL parsing fails
- Validate WSDL against schema
- Check for unsupported features
- Enable debug logging

### Issue: Performance problems
- Profile the code
- Check for excessive allocations
- Optimize template rendering

## Contributing Guidelines

1. Fork the repository
2. Create feature branch
3. Write tests first
4. Implement feature
5. Ensure tests pass
6. Run linter
7. Submit pull request

## Resources

- [WSDL 1.1 Specification](https://www.w3.org/TR/wsdl)
- [WSDL 2.0 Specification](https://www.w3.org/TR/wsdl20/)
- [SOAP 1.1 Specification](https://www.w3.org/TR/2000/NOTE-SOAP-20000508/)
- [Go Code Review Comments](https://github.com/golang/go/wiki/CodeReviewComments)