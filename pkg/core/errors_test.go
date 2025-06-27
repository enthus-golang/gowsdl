// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at http://mozilla.org/MPL/2.0/.

package core

import (
	"errors"
	"fmt"
	"testing"

	"github.com/enthus-golang/gowsdl/pkg/types"
)

func TestWSDLError(t *testing.T) {
	tests := []struct {
		name     string
		err      *types.WSDLError
		expected string
	}{
		{
			name: "with path",
			err: &types.WSDLError{
				Op:   "parse",
				Path: "/path/to/file.wsdl",
				Err:  errors.New("invalid XML"),
			},
			expected: `wsdl parse "/path/to/file.wsdl": invalid XML`,
		},
		{
			name: "without path",
			err: &types.WSDLError{
				Op:  "generate",
				Err: errors.New("template error"),
			},
			expected: "wsdl generate: template error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.err.Error(); got != tt.expected {
				t.Errorf("WSDLError.Error() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestWSDLErrorUnwrap(t *testing.T) {
	baseErr := errors.New("base error")
	wsdlErr := &types.WSDLError{
		Op:   "test",
		Path: "test.wsdl",
		Err:  baseErr,
	}

	if unwrapped := wsdlErr.Unwrap(); unwrapped != baseErr {
		t.Errorf("WSDLError.Unwrap() = %v, want %v", unwrapped, baseErr)
	}

	// Test with errors.Is
	if !errors.Is(wsdlErr, baseErr) {
		t.Error("errors.Is(wsdlErr, baseErr) = false, want true")
	}
}

func TestSchemaError(t *testing.T) {
	tests := []struct {
		name     string
		err      *types.SchemaError
		expected string
	}{
		{
			name: "with schema name",
			err: &types.SchemaError{
				Op:     "validate",
				Schema: "http://example.com/schema.xsd",
				Err:    errors.New("invalid type"),
			},
			expected: `schema validate "http://example.com/schema.xsd": invalid type`,
		},
		{
			name: "without schema name",
			err: &types.SchemaError{
				Op:  "parse",
				Err: errors.New("syntax error"),
			},
			expected: "schema parse: syntax error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.err.Error(); got != tt.expected {
				t.Errorf("SchemaError.Error() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestSchemaErrorUnwrap(t *testing.T) {
	baseErr := errors.New("base error")
	schemaErr := &types.SchemaError{
		Op:     "test",
		Schema: "test.xsd",
		Err:    baseErr,
	}

	if unwrapped := schemaErr.Unwrap(); unwrapped != baseErr {
		t.Errorf("SchemaError.Unwrap() = %v, want %v", unwrapped, baseErr)
	}

	// Test with errors.Is
	if !errors.Is(schemaErr, baseErr) {
		t.Error("errors.Is(schemaErr, baseErr) = false, want true")
	}
}

// ValidationError tests are in cmd/gowsdl/main_test.go since the type is defined there

func TestErrorsAs(t *testing.T) {
	// Test WSDL error chain
	wsdlErr := &types.WSDLError{
		Op:   "download",
		Path: "http://example.com/service.wsdl",
		Err:  fmt.Errorf("network error"),
	}

	wrappedErr := fmt.Errorf("failed to process WSDL: %w", wsdlErr)

	var targetWSDLErr *types.WSDLError
	if !errors.As(wrappedErr, &targetWSDLErr) {
		t.Error("errors.As failed to extract WSDLError from chain")
	}

	if targetWSDLErr.Op != "download" {
		t.Errorf("extracted error has wrong Op: got %v, want download", targetWSDLErr.Op)
	}

	// Test Schema error chain
	schemaErr := &types.SchemaError{
		Op:     "resolve",
		Schema: "types.xsd",
		Err:    fmt.Errorf("file not found"),
	}

	wrappedSchemaErr := fmt.Errorf("schema processing failed: %w", schemaErr)

	var targetSchemaErr *types.SchemaError
	if !errors.As(wrappedSchemaErr, &targetSchemaErr) {
		t.Error("errors.As failed to extract SchemaError from chain")
	}

	if targetSchemaErr.Schema != "types.xsd" {
		t.Errorf("extracted error has wrong Schema: got %v, want types.xsd", targetSchemaErr.Schema)
	}
}

func TestMultipleErrorsInStartWithContext(t *testing.T) {
	t.Log("Integration test for StartWithContext error aggregation")
	t.Skip("TODO: Implement this test after GoWSDL is properly refactored")
}