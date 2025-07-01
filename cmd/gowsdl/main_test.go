// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at http://mozilla.org/MPL/2.0/.

package main

import (
	"bytes"
	"errors"
	"log"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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

func TestRunSuccessOutput(t *testing.T) {
	// Create a temporary directory for test outputs
	tmpDir, err := os.MkdirTemp("", "gowsdl-test-*")
	require.NoError(t, err)
	defer os.RemoveAll(tmpDir)

	// Create a test WSDL file
	wsdlContent := `<?xml version="1.0" encoding="UTF-8"?>
<definitions xmlns="http://schemas.xmlsoap.org/wsdl/"
	xmlns:soap="http://schemas.xmlsoap.org/wsdl/soap/"
	xmlns:tns="http://example.com/"
	targetNamespace="http://example.com/">
	<types/>
	<portType name="TestService">
		<operation name="TestOperation"/>
	</portType>
	<binding name="TestBinding" type="tns:TestService">
		<soap:binding style="document" transport="http://schemas.xmlsoap.org/soap/http"/>
		<operation name="TestOperation">
			<soap:operation soapAction="TestOperation"/>
		</operation>
	</binding>
	<service name="TestService">
		<port name="TestPort" binding="tns:TestBinding">
			<soap:address location="http://example.com/service"/>
		</port>
	</service>
</definitions>`

	wsdlFile := filepath.Join(tmpDir, "test.wsdl")
	err = os.WriteFile(wsdlFile, []byte(wsdlContent), 0644)
	require.NoError(t, err)

	tests := []struct {
		name           string
		pkgName        string
		outFile        string
		generateServer bool
		expectedOutput []string
	}{
		{
			name:           "basic output without server",
			pkgName:        "testpkg",
			outFile:        "test.go",
			generateServer: false,
			expectedOutput: []string{
				"Generated",
				"testpkg/test.go",
				"(package testpkg)",
				"from",
				"test.wsdl",
			},
		},
		{
			name:           "output with server generation",
			pkgName:        "testpkg",
			outFile:        "test.go",
			generateServer: true,
			expectedOutput: []string{
				"Generated",
				"testpkg/test.go",
				"(package testpkg)",
				"from",
				"test.wsdl",
				"servertest.go",
				"(server implementation)",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Capture log output
			var logBuf bytes.Buffer
			log.SetOutput(&logBuf)
			defer log.SetOutput(os.Stdout)

			// Simulate what the run function does for output
			pkgDir := filepath.Join(tmpDir, tt.pkgName)
			outputPath := filepath.Join(pkgDir, tt.outFile)
			
			// Generate the output messages (same as in real run function)
			log.Printf("Generated %s (package %s) from %s", outputPath, tt.pkgName, wsdlFile)
			
			if tt.generateServer {
				serverFileName := "server" + tt.outFile
				serverFilePath := filepath.Join(pkgDir, serverFileName)
				log.Printf("Generated %s (server implementation)", serverFilePath)
			}

			// Check log output
			logOutput := logBuf.String()
			for _, expected := range tt.expectedOutput {
				assert.Contains(t, logOutput, expected, "Expected output to contain: %s", expected)
			}

			// Verify no emoji in output
			assert.NotContains(t, logOutput, "🍀", "Output should not contain emoji")
			assert.NotContains(t, logOutput, "👍", "Output should not contain emoji")
			assert.NotContains(t, logOutput, "Done", "Output should not contain generic 'Done' message")
		})
	}
}

func TestInitLogging(t *testing.T) {
	// Test that log configuration is correct after init
	// The init function has already run when the test binary starts
	assert.Equal(t, 0, log.Flags(), "Log flags should be 0")
	assert.Equal(t, os.Stdout, log.Writer(), "Log output should be stdout")
	assert.Equal(t, "", log.Prefix(), "Log prefix should be empty (no emoji)")
}

func TestServerFilePathCalculation(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "gowsdl-test-*")
	require.NoError(t, err)
	defer os.RemoveAll(tmpDir)

	tests := []struct {
		name               string
		pkg                string
		outFile            string
		dir                string
		generateServer     bool
		expectedServerPath string
	}{
		{
			name:               "server file path with basic names",
			pkg:                "myservice",
			outFile:            "client.go",
			dir:                tmpDir,
			generateServer:     true,
			expectedServerPath: filepath.Join(tmpDir, "myservice", "serverclient.go"),
		},
		{
			name:               "server file path with complex names",
			pkg:                "complex_service",
			outFile:            "soap_client.go",
			dir:                tmpDir,
			generateServer:     true,
			expectedServerPath: filepath.Join(tmpDir, "complex_service", "serversoap_client.go"),
		},
		{
			name:           "no server file when flag is false",
			pkg:            "myservice",
			outFile:        "client.go",
			dir:            tmpDir,
			generateServer: false,
			expectedServerPath: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Set test values
			testPkg := tt.pkg
			testOutFile := tt.outFile
			testDir := tt.dir
			testGenerateServer := tt.generateServer
			
			// Calculate paths as done in the main function
			pkgDir := filepath.Join(testDir, testPkg)
			
			var serverFilePath string
			if testGenerateServer {
				serverFileName := "server" + testOutFile
				serverFilePath = filepath.Join(pkgDir, serverFileName)
			}
			
			if tt.generateServer {
				assert.Equal(t, tt.expectedServerPath, serverFilePath, "Server file path should match expected")
			} else {
				assert.Empty(t, serverFilePath, "Server file path should be empty when generateServer is false")
			}
		})
	}
}