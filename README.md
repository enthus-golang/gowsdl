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
	* WSDL 1.1 and WSDL 2.0
	* XML Schema 1.0
	* SOAP 1.1
* Resolve external XML Schemas
* Support external and local WSDL
* Full namespace handling and preservation
* XSD element reference resolution

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
        Generate code using Go generics (requires Go 1.18+ for generic features)
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
    if result.TransportError {
        // Network/transport error (connection refused, timeout, etc.)
        fmt.Printf("Transport Error: %s\n", result.Fault.String)
    } else {
        // Business logic error from SOAP service
        fmt.Printf("SOAP Fault: %s\n", result.Fault.String)
    }
}
```

The generic interface provides better type safety and explicit fault handling, making it easier to distinguish between transport errors and business logic errors (SOAP faults).

### New Features

#### WSDL 2.0 Support

gowsdl now fully supports both WSDL 1.1 and WSDL 2.0 documents with complete code generation capabilities. Features include:

* **Automatic version detection** - WSDL version is detected from the root element namespace
* **Complete parsing support** for WSDL 2.0 structures:
  - Interface definitions (replacing portType from WSDL 1.1)
  - Endpoint definitions (replacing port from WSDL 1.1)
  - Modern message exchange patterns
  - Enhanced fault handling with input/output fault references
* **Full code generation** - Operations, types, and server code are generated using WSDL 2.0-specific templates
* **Backward compatibility** - All existing WSDL 1.1 functionality remains unchanged

Example usage with WSDL 2.0 files:
```bash
# Works seamlessly with both WSDL 1.1 and 2.0
gowsdl -p myservice -o myservice.go https://example.com/service.wsdl
```

#### Enhanced Namespace Handling

Namespaces are now properly preserved and managed throughout the code generation process:

* XML namespace prefixes are maintained in generated code
* Namespace collisions are automatically resolved
* Generated XML tags include proper namespace information
* Support for qualified and unqualified element/attribute forms

#### XSD Element Reference Resolution

The code generator now properly resolves XSD element references:

* Global element references are fully resolved
* Cross-schema element references are supported
* Properties from referenced elements are properly inherited
* Complex type references within elements are handled correctly

#### Improved Local Schema Resolution

When working with local WSDL files that reference external schemas:

* Relative schema paths are correctly resolved
* Local schema imports work seamlessly
* Recursive schema resolution is supported with cycle detection
* Mixed local and remote schema references are handled

### Implementation Status

This implementation addresses all the features requested in [issue #10](https://github.com/enthus-golang/gowsdl/issues/10), including the original TODOs:

- ✅ **Support for generating namespaces** (main.go:41) - Complete namespace management system
- ✅ **Resolve XSD element references** (main.go:39) - Full element reference resolution  
- ✅ **Local schema resolution** (main.go:37) - Enhanced local and remote schema handling
- ✅ **WSDL 2.0 Support** - Complete parsing and code generation
- ✅ **Better error messages** - Contextual errors with suggestions

All features include comprehensive test coverage and maintain backward compatibility with existing WSDL 1.1 code.
