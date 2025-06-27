// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at http://mozilla.org/MPL/2.0/.

package gowsdl

import (
	"context"
	"errors"
	"strings"

	"github.com/enthus-golang/gowsdl/pkg/generator"
	"github.com/enthus-golang/gowsdl/pkg/utils"
)

// GoWSDL is the main facade for WSDL code generation
type GoWSDL struct {
	generator *generator.Generator
}

// Option defines a function that configures GoWSDL
type Option func(*GoWSDL) error

// WithHTTPConfig sets the HTTP client configuration
func WithHTTPConfig(config *HTTPClientConfig) Option {
	return func(g *GoWSDL) error {
		return generator.WithHTTPConfig(config)(g.generator)
	}
}

// WithGenerics enables Go generics code generation
func WithGenerics() Option {
	return func(g *GoWSDL) error {
		return generator.WithGenerics()(g.generator)
	}
}

// WithExportAllTypes enables exporting all types (making them public)
func WithExportAllTypes(export bool) Option {
	return func(g *GoWSDL) error {
		return generator.WithExportAllTypes(export)(g.generator)
	}
}

// WithPackage sets the package name for generated code
func WithPackage(pkg string) Option {
	return func(g *GoWSDL) error {
		return generator.WithPackage(pkg)(g.generator)
	}
}

// NewGoWSDL creates a new GoWSDL instance with the provided options
func NewGoWSDL(file string, opts ...Option) (*GoWSDL, error) {
	file = strings.TrimSpace(file)
	if file == "" {
		return nil, errors.New("WSDL file is required to generate Go proxy")
	}

	// Convert options to generator options
	genOpts := make([]generator.Option, 0, len(opts))
	
	// Create generator
	gen, err := generator.New(file, genOpts...)
	if err != nil {
		return nil, err
	}

	g := &GoWSDL{
		generator: gen,
	}

	// Apply options
	for _, opt := range opts {
		if err := opt(g); err != nil {
			return nil, err
		}
	}

	return g, nil
}

// Legacy constructors for backward compatibility
func NewGoWSDLWithConfig(file, pkg string, httpConfig *HTTPClientConfig, exportAllTypes bool) (*GoWSDL, error) {
	opts := []Option{
		WithPackage(pkg),
		WithHTTPConfig(httpConfig),
		WithExportAllTypes(exportAllTypes),
	}
	return NewGoWSDL(file, opts...)
}

func NewGoWSDLWithOptions(file, pkg string, httpConfig *HTTPClientConfig, exportAllTypes bool, useGenerics bool) (*GoWSDL, error) {
	opts := []Option{
		WithPackage(pkg),
		WithHTTPConfig(httpConfig),
		WithExportAllTypes(exportAllTypes),
	}
	if useGenerics {
		opts = append(opts, WithGenerics())
	}
	return NewGoWSDL(file, opts...)
}

// StartWithContext initiates the code generation process with context support
func (g *GoWSDL) StartWithContext(ctx context.Context) (map[string][]byte, error) {
	return g.generator.Generate(ctx)
}

// Start initiates the code generation process
func (g *GoWSDL) Start() (map[string][]byte, error) {
	return g.StartWithContext(context.Background())
}

// Utility functions exported for backward compatibility
var (
	MakePublic = utils.MakePublic
)