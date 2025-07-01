// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at http://mozilla.org/MPL/2.0/.

package main

import (
	"errors"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestValidationError(t *testing.T) {
	tests := []struct {
		name     string
		err      *ValidationError
		expected string
	}{
		{
			name: "path traversal",
			err: &ValidationError{
				Field: "path",
				Value: "../../../etc/passwd",
				Err:   errors.New("contains directory traversal sequence '..'"),
			},
			expected: `validation failed for path "../../../etc/passwd": contains directory traversal sequence '..'`,
		},
		{
			name: "empty package name",
			err: &ValidationError{
				Field: "package name",
				Value: "",
				Err:   errors.New("cannot be empty"),
			},
			expected: `validation failed for package name "": cannot be empty`,
		},
		{
			name: "unsafe characters",
			err: &ValidationError{
				Field: "output file",
				Value: "file<>name.go",
				Err:   errors.New("contains unsafe characters"),
			},
			expected: `validation failed for output file "file<>name.go": contains unsafe characters`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.err.Error(); got != tt.expected {
				t.Errorf("ValidationError.Error() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestValidationErrorUnwrap(t *testing.T) {
	baseErr := errors.New("base error")
	validationErr := &ValidationError{
		Field: "test",
		Value: "test-value",
		Err:   baseErr,
	}

	if unwrapped := validationErr.Unwrap(); unwrapped != baseErr {
		t.Errorf("ValidationError.Unwrap() = %v, want %v", unwrapped, baseErr)
	}

	// Test with errors.Is
	if !errors.Is(validationErr, baseErr) {
		t.Error("errors.Is(validationErr, baseErr) = false, want true")
	}
}

func TestValidatePath(t *testing.T) {
	tests := []struct {
		name    string
		path    string
		wantErr bool
	}{
		{
			name:    "valid relative path",
			path:    "output/myservice.go",
			wantErr: false,
		},
		{
			name:    "relative path going up",
			path:    "../wsdl",
			wantErr: false,
		},
		{
			name:    "absolute path is allowed",
			path:    "/etc/passwd",
			wantErr: false,
		},
		{
			name:    "clean path with dots",
			path:    "./output/file.go",
			wantErr: false,
		},
		{
			name:    "empty path",
			path:    "",
			wantErr: true,
		},
		{
			name:    "valid parent directory for output",
			path:    "../output",
			wantErr: false,
		},
		{
			name:    "tmp directory is allowed",
			path:    "/tmp",
			wantErr: false,
		},
		{
			name:    "system directory is allowed",
			path:    "/usr/local/bin",
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validatePath(tt.path)
			if (err != nil) != tt.wantErr {
				t.Errorf("validatePath() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestValidateIdentifier(t *testing.T) {
	tests := []struct {
		name      string
		value     string
		fieldType string
		wantErr   bool
	}{
		{
			name:      "valid package name",
			value:     "myservice",
			fieldType: "package name",
			wantErr:   false,
		},
		{
			name:      "empty value",
			value:     "",
			fieldType: "package name",
			wantErr:   true,
		},
		{
			name:      "unsafe characters",
			value:     "my<service>",
			fieldType: "output file",
			wantErr:   true,
		},
		{
			name:      "path traversal in name",
			value:     "../../../passwd",
			fieldType: "output file",
			wantErr:   true,
		},
		{
			name:      "valid file name",
			value:     "service_client.go",
			fieldType: "output file",
			wantErr:   false,
		},
		{
			name:      "file name with dash",
			value:     "service-client.go",
			fieldType: "output file",
			wantErr:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateIdentifier(tt.value, tt.fieldType)
			if (err != nil) != tt.wantErr {
				t.Errorf("validateIdentifier() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestWriteData(t *testing.T) {
	t.Run("successful write", func(t *testing.T) {
		tmpFile, err := os.CreateTemp("", "test-write-*.tmp")
		assert.NoError(t, err)
		defer func() {
			if err := os.Remove(tmpFile.Name()); err != nil {
				t.Logf("Failed to remove temp file: %v", err)
			}
		}()
		defer func() {
			if err := tmpFile.Close(); err != nil {
				t.Logf("Failed to close temp file: %v", err)
			}
		}()

		data := []byte("test data content")
		err = writeData(tmpFile, data)
		assert.NoError(t, err)

		// Verify data was written correctly
		_, err = tmpFile.Seek(0, 0)
		assert.NoError(t, err)
		readData := make([]byte, len(data))
		n, err := tmpFile.Read(readData)
		assert.NoError(t, err)
		assert.Equal(t, len(data), n)
		assert.Equal(t, data, readData)
	})

	t.Run("write to closed file", func(t *testing.T) {
		tmpFile, err := os.CreateTemp("", "test-write-*.tmp")
		assert.NoError(t, err)
		defer func() {
			if err := os.Remove(tmpFile.Name()); err != nil {
				t.Logf("Failed to remove temp file: %v", err)
			}
		}()
		
		// Close the file before writing
		err = tmpFile.Close()
		assert.NoError(t, err)

		data := []byte("test data")
		err = writeData(tmpFile, data)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to write to file")
	})

	t.Run("write empty data", func(t *testing.T) {
		tmpFile, err := os.CreateTemp("", "test-write-*.tmp")
		assert.NoError(t, err)
		defer func() {
			if err := os.Remove(tmpFile.Name()); err != nil {
				t.Logf("Failed to remove temp file: %v", err)
			}
		}()
		defer func() {
			if err := tmpFile.Close(); err != nil {
				t.Logf("Failed to close temp file: %v", err)
			}
		}()

		data := []byte{}
		err = writeData(tmpFile, data)
		assert.NoError(t, err)
	})

	t.Run("write large data", func(t *testing.T) {
		tmpFile, err := os.CreateTemp("", "test-write-*.tmp")
		assert.NoError(t, err)
		defer func() {
			if err := os.Remove(tmpFile.Name()); err != nil {
				t.Logf("Failed to remove temp file: %v", err)
			}
		}()
		defer func() {
			if err := tmpFile.Close(); err != nil {
				t.Logf("Failed to close temp file: %v", err)
			}
		}()

		// Create a large data buffer (1MB)
		data := make([]byte, 1024*1024)
		for i := range data {
			data[i] = byte(i % 256)
		}

		err = writeData(tmpFile, data)
		assert.NoError(t, err)

		// Verify file size
		stat, err := tmpFile.Stat()
		assert.NoError(t, err)
		assert.Equal(t, int64(len(data)), stat.Size())
	})
}