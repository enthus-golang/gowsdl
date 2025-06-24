# WSDL to Go

[![Gitter](https://badges.gitter.im/Join%20Chat.svg)](https://gitter.im/hooklift/gowsdl?utm_source=badge&utm_medium=badge&utm_campaign=pr-badge&utm_content=badge)
[![GoDoc](https://godoc.org/github.com/hooklift/gowsdl?status.svg)](https://godoc.org/github.com/hooklift/gowsdl)
[![Build Status](https://github.com/hooklift/gowsdl/workflows/Test/badge.svg)](https://github.com/hooklift/gowsdl/actions)
[![codecov](https://codecov.io/gh/hooklift/gowsdl/branch/main/graph/badge.svg)](https://codecov.io/gh/hooklift/gowsdl)
[![Go Report Card](https://goreportcard.com/badge/github.com/hooklift/gowsdl)](https://goreportcard.com/report/github.com/hooklift/gowsdl)

Generates Go code from a WSDL file.

### Requirements

* Go 1.21 or higher

### Install

* [Download release](https://github.com/hooklift/gowsdl/releases)
* Download and build locally
    * Go 1.21+: `go install github.com/hooklift/gowsdl/cmd/gowsdl@latest`
* Install from Homebrew: `brew install gowsdl`

### Goals
* Generate idiomatic Go code as much as possible
* Support only Document/Literal wrapped services, which are [WS-I](http://ws-i.org/) compliant
* Support:
	* WSDL 1.1
	* XML Schema 1.0
	* SOAP 1.1
* Resolve external XML Schemas
* Support external and local WSDL

### Caveats
* Please keep in mind that the generated code is just a reflection of what the WSDL is like. If your WSDL has duplicated type definitions, your Go code is going to have the same and may not compile.

### Usage
```
Usage: gowsdl [options] myservice.wsdl
  -o string
        File where the generated code will be saved (default "myservice.go")
  -p string
        Package under which code will be generated (default "myservice")
  -i    Skips TLS Verification
  -v    Shows gowsdl version
  -use-generics
        Generate code using Go generics (requires Go 1.18+)
  ```

### Go Generics Support

Starting with Go 1.18, gowsdl can generate code that uses generics for type-safe SOAP operations. This feature provides:

* Type-safe SOAP clients with compile-time type checking
* Generic result types that can handle both success responses and SOAP faults
* Generic array types for better handling of unbounded elements
* Backward compatibility with non-generic code generation

To enable generic code generation, use the `-use-generics` flag:

```bash
gowsdl -use-generics -p myservice -o myservice.go myservice.wsdl
```

#### Example Usage

With generics enabled, you get additional methods for each operation:

```go
// Standard interface (always generated)
response, err := client.GetUser(&GetUserRequest{UserID: 123})
if err != nil {
    // Handle error
}

// Generic interface (only with -use-generics flag)
result, err := client.GetUserGeneric(&GetUserRequest{UserID: 123})
if err != nil {
    // Handle transport error
}

if result.IsSuccess() {
    user, _ := result.Unwrap()
    // Use user
} else {
    // Handle SOAP fault
    fmt.Printf("SOAP Fault: %s\n", result.Fault.String)
}
```

The generic interface provides better type safety and explicit fault handling, making it easier to distinguish between transport errors and business logic errors (SOAP faults).
