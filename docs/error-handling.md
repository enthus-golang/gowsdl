# Error Handling in gowsdl

## Overview

`gowsdl` uses Go 1.13+ error wrapping patterns to provide detailed context about errors. This allows consumers to understand what went wrong and where, and to handle specific error types programmatically.

## Error Types

### WSDLError

Represents errors related to WSDL processing.

```go
type WSDLError struct {
    Op   string // operation that failed
    Path string // file path or URL
    Err  error  // underlying error
}
```

Common operations (`Op` values):
- `"fetch"` - Failed to download or read WSDL file
- `"parse"` - Failed to parse WSDL XML
- `"resolve_schemas"` - Failed to resolve external schemas

### SchemaError

Represents errors related to XSD schema processing.

```go
type SchemaError struct {
    Op     string // operation that failed
    Schema string // schema file or URL
    Err    error  // underlying error
}
```

Common operations (`Op` values):
- `"parse_reference"` - Failed to parse schema reference
- `"fetch"` - Failed to download schema file
- `"parse"` - Failed to parse schema XML
- `"resolve_nested"` - Failed to resolve nested schemas

### ValidationError (CLI only)

Represents input validation errors in the command-line tool.

```go
type ValidationError struct {
    Field string // field that failed validation
    Value string // invalid value
    Err   error  // underlying error
}
```

Common validation errors:
- Directory traversal attempts
- Empty required fields
- Unsafe characters in file names
- Absolute paths where relative paths are required

## Error Handling Patterns

### Checking for Specific Error Types

Use `errors.As` to check for specific error types:

```go
import "github.com/hooklift/gowsdl"

g, err := gowsdl.NewGoWSDL(wsdlPath, "myservice", false, true)
if err != nil {
    var wsdlErr *gowsdl.WSDLError
    if errors.As(err, &wsdlErr) {
        log.Printf("WSDL error in operation %s for %s: %v", 
            wsdlErr.Op, wsdlErr.Path, wsdlErr.Err)
    }
    return err
}
```

### Checking for Underlying Errors

Use `errors.Is` to check for specific underlying errors:

```go
gocode, err := g.Start()
if err != nil {
    if errors.Is(err, io.EOF) {
        // Handle unexpected EOF
    }
    return err
}
```

### Error Context

All errors include context about what operation failed and what resource was being processed:

```
wsdl parse "http://example.com/service.wsdl": failed to unmarshal WSDL: XML syntax error on line 42
schema fetch "http://example.com/types.xsd": connection timeout
validation failed for path "../../../etc/passwd": contains directory traversal sequence '..'
```

## Best Practices

1. **Always check errors**: Don't ignore returned errors
2. **Use error types for handling**: Use `errors.As` when you need to handle specific error conditions
3. **Log full error context**: The error messages include valuable debugging information
4. **Wrap errors in your code**: When building on top of gowsdl, wrap errors with additional context:
   ```go
   if err != nil {
       return fmt.Errorf("generating client for service %s: %w", serviceName, err)
   }
   ```

## Migration from Older Versions

If you're upgrading from an older version of gowsdl:

1. Replace direct error comparisons:
   ```go
   // Old
   if err == io.EOF {
   
   // New
   if errors.Is(err, io.EOF) {
   ```

2. Use error type assertions for better error handling:
   ```go
   // Old
   if err != nil {
       log.Fatal(err)
   }
   
   // New
   if err != nil {
       var wsdlErr *gowsdl.WSDLError
       if errors.As(err, &wsdlErr) {
           // Handle WSDL-specific error
       }
       return err
   }
   ```

## Goroutine Error Handling

The `StartWithContext` method properly aggregates errors from concurrent goroutines. If any code generation step fails, the entire operation fails with a descriptive error:

- Single error: The exact error is returned
- Multiple errors: A combined error message listing all failures

This ensures that partial code generation failures are detected and reported properly.