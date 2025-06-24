// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at http://mozilla.org/MPL/2.0/.

package gowsdl

import (
	"errors"
	"fmt"
	"testing"
)

func TestWSDLError(t *testing.T) {
	tests := []struct {
		name     string
		err      *WSDLError
		expected string
	}{
		{
			name: "with path",
			err: &WSDLError{
				Op:   "parse",
				Path: "/path/to/file.wsdl",
				Err:  errors.New("invalid XML"),
			},
			expected: `wsdl parse "/path/to/file.wsdl": invalid XML`,
		},
		{
			name: "without path",
			err: &WSDLError{
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
	wsdlErr := &WSDLError{
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
		err      *SchemaError
		expected string
	}{
		{
			name: "with schema name",
			err: &SchemaError{
				Op:     "validate",
				Schema: "http://example.com/schema.xsd",
				Err:    errors.New("invalid type"),
			},
			expected: `schema validate "http://example.com/schema.xsd": invalid type`,
		},
		{
			name: "without schema name",
			err: &SchemaError{
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
	schemaErr := &SchemaError{
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
	wsdlErr := &WSDLError{
		Op:   "download",
		Path: "http://example.com/service.wsdl",
		Err:  fmt.Errorf("network error"),
	}

	wrappedErr := fmt.Errorf("failed to process WSDL: %w", wsdlErr)

	var targetWSDLErr *WSDLError
	if !errors.As(wrappedErr, &targetWSDLErr) {
		t.Error("errors.As failed to extract WSDLError from chain")
	}

	if targetWSDLErr.Op != "download" {
		t.Errorf("extracted error has wrong Op: got %v, want download", targetWSDLErr.Op)
	}

	// Test Schema error chain
	schemaErr := &SchemaError{
		Op:     "resolve",
		Schema: "types.xsd",
		Err:    fmt.Errorf("file not found"),
	}

	wrappedSchemaErr := fmt.Errorf("schema processing failed: %w", schemaErr)

	var targetSchemaErr *SchemaError
	if !errors.As(wrappedSchemaErr, &targetSchemaErr) {
		t.Error("errors.As failed to extract SchemaError from chain")
	}

	if targetSchemaErr.Schema != "types.xsd" {
		t.Errorf("extracted error has wrong Schema: got %v, want types.xsd", targetSchemaErr.Schema)
	}
}

func TestMultipleErrorsInStartWithContext(t *testing.T) {
	// This test verifies that multiple errors in goroutines are properly collected
	// The actual test would require mocking or a test WSDL that triggers errors
	// For now, this is a placeholder showing the expected behavior
	t.Skip("Integration test for StartWithContext error aggregation")
}