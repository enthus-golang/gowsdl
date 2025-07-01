// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at http://mozilla.org/MPL/2.0/.
/*

Gowsdl generates Go code from a WSDL file.

This project is originally intended to generate Go clients for WS-* services.

Usage: gowsdl [options] myservice.wsdl
  -o string
        File where the generated code will be saved (default "myservice.go")
  -p string
        Package under which code will be generated (default "myservice")
  -v    Shows gowsdl version

Features

Supports only Document/Literal wrapped services, which are WS-I (http://ws-i.org/) compliant.

Attempts to generate idiomatic Go code as much as possible.

Supports WSDL 1.1, XML Schema 1.0, SOAP 1.1.

Resolves external XML Schemas

Supports providing WSDL HTTP URL as well as a local WSDL file.

Not supported

UDDI.

TODO

Add support for filters to allow the user to change the generated code.

If WSDL file is local, resolve external XML schemas locally too instead of failing due to not having a URL to download them from.

Resolve XSD element references.

Support for generating namespaces.

Make code generation agnostic so generating code to other programming languages is feasible through plugins.

*/

package main

import (
	"bytes"
	"crypto/tls"
	"errors"
	"flag"
	"fmt"
	"go/format"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	gen "github.com/enthus-golang/gowsdl"
)

// Version is initialized in compilation time by go build.
var Version string

// Name is initialized in compilation time by go build.
var Name string

var vers = flag.Bool("v", false, "Shows gowsdl version")
var pkg = flag.String("p", "myservice", "Package under which code will be generated")
var outFile = flag.String("o", "myservice.go", "File where the generated code will be saved")
var dir = flag.String("d", "./", "Directory under which package directory will be created")
var insecure = flag.Bool("i", false, "Skips TLS Verification")
var makePublic = flag.Bool("make-public", true, "Make the generated types public/exported")
var useGenerics = flag.Bool("use-generics", false, "Generate code using Go generics (requires Go 1.18+)")
var generateServer = flag.Bool("generate-server", false, "Generate server implementation code")

// HTTP client configuration flags
var httpTimeout = flag.Duration("http-timeout", 30*time.Second, "Timeout for HTTP requests")
var httpRetries = flag.Int("http-retries", 0, "Number of retries for failed requests")
var httpRetryDelay = flag.Duration("http-retry-delay", 1*time.Second, "Delay between retries")
var httpRateLimit = flag.Int("http-rate-limit", 0, "Rate limit for HTTP requests per second (0 = no limit)")
var httpUserAgent = flag.String("http-user-agent", "gowsdl/1.0", "User-Agent header for requests")
var httpProxy = flag.String("http-proxy", "", "HTTP proxy URL")
var tlsCACert = flag.String("tls-ca-cert", "", "Path to CA certificate file")
var tlsClientCert = flag.String("tls-client-cert", "", "Path to client certificate file (requires -tls-client-key)")
var tlsClientKey = flag.String("tls-client-key", "", "Path to client key file (requires -tls-client-cert)")
var tlsMinVersion = flag.String("tls-min-version", "1.2", "Minimum TLS version (1.0, 1.1, 1.2, 1.3)")

// ValidationError represents an input validation error
type ValidationError struct {
	Field string // field that failed validation
	Value string // invalid value
	Err   error  // underlying error
}

func (e *ValidationError) Error() string {
	return fmt.Sprintf("validation failed for %s %q: %v", e.Field, e.Value, e.Err)
}

func (e *ValidationError) Unwrap() error {
	return e.Err
}

func init() {
	log.SetFlags(0)
	log.SetOutput(os.Stdout)
	log.SetPrefix("🍀  ")
}

// validatePath validates a file path to prevent directory traversal attacks
func validatePath(path string) error {
	// Clean and normalize the path
	cleaned := filepath.Clean(path)
	
	// Check for directory traversal attempts
	if strings.Contains(cleaned, "..") {
		return &ValidationError{
			Field: "path",
			Value: path,
			Err:   errors.New("contains directory traversal sequence '..'"),
		}
	}
	
	// Check for absolute paths that might overwrite system files
	if filepath.IsAbs(cleaned) {
		return &ValidationError{
			Field: "path",
			Value: path,
			Err:   errors.New("absolute paths not allowed for security"),
		}
	}
	
	return nil
}

// validateIdentifier validates package names and file names
func validateIdentifier(name, fieldType string) error {
	if name == "" {
		return &ValidationError{
			Field: fieldType,
			Value: name,
			Err:   errors.New("cannot be empty"),
		}
	}
	
	// Check for common unsafe characters
	if strings.ContainsAny(name, "/<>:\"|?*") {
		return &ValidationError{
			Field: fieldType,
			Value: name,
			Err:   errors.New("contains unsafe characters"),
		}
	}
	
	// Check for relative path components
	if strings.Contains(name, "..") || strings.Contains(name, "./") {
		return &ValidationError{
			Field: fieldType,
			Value: name,
			Err:   errors.New("contains path traversal sequences"),
		}
	}
	
	return nil
}

// writeData is a helper function to write data to a file with proper error handling
func writeData(file *os.File, data []byte) error {
	if _, err := file.Write(data); err != nil {
		return fmt.Errorf("failed to write to file: %w", err)
	}
	return nil
}

func main() {
	if err := run(); err != nil {
		// Check for specific error types to provide better error messages
		var validationErr *ValidationError
		if errors.As(err, &validationErr) {
			log.Fatalf("Validation error: %v", validationErr)
		}
		
		var wsdlErr *gen.WSDLError
		if errors.As(err, &wsdlErr) {
			log.Fatalf("WSDL error: %v", wsdlErr)
		}
		
		var schemaErr *gen.SchemaError  
		if errors.As(err, &schemaErr) {
			log.Fatalf("Schema error: %v", schemaErr)
		}
		
		log.Fatalf("Error: %v", err)
	}
}

func run() error {
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: %s [options] myservice.wsdl\n", os.Args[0])
		flag.PrintDefaults()
	}

	flag.Parse()

	// Show app version
	if *vers {
		log.Println(Version)
		return nil
	}

	if len(os.Args) < 2 {
		flag.Usage()
		return nil
	}

	wsdlPath := os.Args[len(os.Args)-1]

	if *outFile == wsdlPath {
		return errors.New("output file cannot be the same WSDL file")
	}

	// Validate inputs for security
	if err := validateIdentifier(*pkg, "package name"); err != nil {
		return fmt.Errorf("invalid package name: %w", err)
	}
	
	if err := validateIdentifier(*outFile, "output file"); err != nil {
		return fmt.Errorf("invalid output file: %w", err)
	}
	
	if err := validatePath(*dir); err != nil {
		return fmt.Errorf("invalid directory: %w", err)
	}

	// Create HTTP client configuration
	httpConfig := gen.DefaultHTTPClientConfig()
	httpConfig.Timeout = *httpTimeout
	httpConfig.MaxRetries = *httpRetries
	httpConfig.RetryDelay = *httpRetryDelay
	httpConfig.RateLimitPerSecond = *httpRateLimit
	httpConfig.UserAgent = *httpUserAgent
	httpConfig.ProxyURL = *httpProxy
	httpConfig.InsecureSkipVerify = *insecure

	// Configure TLS
	if *tlsMinVersion != "" {
		var minVersion uint16
		switch *tlsMinVersion {
		case "1.0":
			minVersion = tls.VersionTLS10
		case "1.1":
			minVersion = tls.VersionTLS11
		case "1.2":
			minVersion = tls.VersionTLS12
		case "1.3":
			minVersion = tls.VersionTLS13
		default:
			return fmt.Errorf("invalid TLS version: %s", *tlsMinVersion)
		}
		if httpConfig.TLSConfig == nil {
			httpConfig.TLSConfig = &tls.Config{}
		}
		httpConfig.TLSConfig.MinVersion = minVersion
	}

	// Add CA certificate if provided
	if *tlsCACert != "" {
		if _, err := httpConfig.WithCACert(*tlsCACert); err != nil {
			return fmt.Errorf("failed to load CA certificate: %w", err)
		}
	}

	// Add client certificate if provided
	if *tlsClientCert != "" && *tlsClientKey != "" {
		if _, err := httpConfig.WithClientCert(*tlsClientCert, *tlsClientKey); err != nil {
			return fmt.Errorf("failed to load client certificate: %w", err)
		}
	} else if *tlsClientCert != "" || *tlsClientKey != "" {
		return errors.New("both -tls-client-cert and -tls-client-key must be provided together")
	}

	// Create options for the generator
	opts := []gen.Option{
		gen.WithPackage(*pkg),
		gen.WithHTTPConfig(httpConfig),
		gen.WithExportAllTypes(*makePublic),
		gen.WithServerGeneration(*generateServer),
	}
	if *useGenerics {
		opts = append(opts, gen.WithGenerics())
	}

	// load wsdl
	gowsdl, err := gen.NewGoWSDL(wsdlPath, opts...)
	if err != nil {
		return fmt.Errorf("failed to initialize WSDL generator: %w", err)
	}

	// generate code
	gocode, err := gowsdl.Start()
	if err != nil {
		return fmt.Errorf("failed to generate code: %w", err)
	}

	pkgDir := filepath.Join(*dir, *pkg)
	
	// Create directory with proper permissions and handle existing directories
	err = os.MkdirAll(pkgDir, 0755)
	if err != nil {
		return fmt.Errorf("failed to create package directory: %w", err)
	}

	// Validate the final output file path
	outputPath := filepath.Join(pkgDir, *outFile)
	if err := validatePath(outputPath); err != nil {
		return fmt.Errorf("invalid output path: %w", err)
	}

	file, err := os.Create(outputPath)
	if err != nil {
		return fmt.Errorf("failed to create output file: %w", err)
	}
	defer func() {
		if err := file.Close(); err != nil {
			log.Printf("Error closing file: %v", err)
		}
	}()

	data := new(bytes.Buffer)
	data.Write(gocode["header"])
	data.Write(gocode["types"])
	data.Write(gocode["operations"])

	// go fmt the generated code
	source, err := format.Source(data.Bytes())
	if err != nil {
		// Write unformatted code and return error with context
		if writeErr := writeData(file, data.Bytes()); writeErr != nil {
			return fmt.Errorf("failed to format code and failed to write unformatted code: format error: %w, write error: %v", err, writeErr)
		}
		return fmt.Errorf("failed to format generated code: %w", err)
	}

	if err := writeData(file, source); err != nil {
		return fmt.Errorf("failed to write formatted code: %w", err)
	}

	// Generate server file only if generateServer flag is set
	if *generateServer {
		serverFileName := "server" + *outFile
		if err := validateIdentifier(serverFileName, "server file name"); err != nil {
			return fmt.Errorf("invalid server file name: %w", err)
		}
		
		serverFilePath := filepath.Join(pkgDir, serverFileName)
		if err := validatePath(serverFilePath); err != nil {
			return fmt.Errorf("invalid server file path: %w", err)
		}
		
		serverFile, err := os.Create(serverFilePath)
		if err != nil {
			return fmt.Errorf("failed to create server file: %w", err)
		}
		defer func() {
			if err := serverFile.Close(); err != nil {
				log.Printf("Error closing server file: %v", err)
			}
		}()

		serverData := new(bytes.Buffer)
		serverData.Write(gocode["server_header"])
		serverData.Write(gocode["server_wsdl"])
		serverData.Write(gocode["server"])

		serverSource, err := format.Source(serverData.Bytes())
		if err != nil {
			// Write unformatted server code and return error with context
			if writeErr := writeData(serverFile, serverData.Bytes()); writeErr != nil {
				return fmt.Errorf("failed to format server code and failed to write unformatted server code: format error: %w, write error: %v", err, writeErr)
			}
			return fmt.Errorf("failed to format generated server code: %w", err)
		}
		
		if err := writeData(serverFile, serverSource); err != nil {
			return fmt.Errorf("failed to write formatted server code: %w", err)
		}
	}

	log.Println("Done 👍")
	return nil
}
