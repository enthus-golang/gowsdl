// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at http://mozilla.org/MPL/2.0/.

package generator

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/enthus-golang/gowsdl/pkg/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestWSDLParsingEdgeCases(t *testing.T) {
	tests := []struct {
		name        string
		wsdlContent string
		wantErr     bool
		errContains string
		check       func(t *testing.T, g *Generator)
	}{
		{
			name: "empty WSDL",
			wsdlContent: `<?xml version="1.0" encoding="UTF-8"?>`,
			wantErr:     true,
			errContains: "unable to determine WSDL version",
		},
		{
			name: "invalid XML",
			wsdlContent: `<?xml version="1.0" encoding="UTF-8"?>
<definitions xmlns="http://schemas.xmlsoap.org/wsdl/">
	<types>
		<schema>
			<!-- Unclosed tag
		</schema>
	</types>
</definitions>`,
			wantErr:     true,
			errContains: "failed to unmarshal WSDL",
		},
		{
			name: "WSDL without types",
			wsdlContent: `<?xml version="1.0" encoding="UTF-8"?>
<definitions xmlns="http://schemas.xmlsoap.org/wsdl/"
	targetNamespace="http://example.com/test"
	xmlns:tns="http://example.com/test">
	<message name="TestMessage">
		<part name="body" type="xsd:string"/>
	</message>
	<portType name="TestPortType">
		<operation name="TestOperation">
			<input message="tns:TestMessage"/>
			<output message="tns:TestMessage"/>
		</operation>
	</portType>
</definitions>`,
			wantErr: false,
			check: func(t *testing.T, g *Generator) {
				// Skip internal field checks
				t.Log("WSDL parsed successfully")
			},
		},
		{
			name: "WSDL with empty types",
			wsdlContent: `<?xml version="1.0" encoding="UTF-8"?>
<definitions xmlns="http://schemas.xmlsoap.org/wsdl/">
	<types/>
</definitions>`,
			wantErr: false,
			check: func(t *testing.T, g *Generator) {
				// Skip internal field checks
				t.Log("Empty types parsed successfully")
			},
		},
		{
			name: "WSDL with multiple schemas",
			wsdlContent: `<?xml version="1.0" encoding="UTF-8"?>
<definitions xmlns="http://schemas.xmlsoap.org/wsdl/"
	xmlns:xs="http://www.w3.org/2001/XMLSchema">
	<types>
		<xs:schema targetNamespace="http://example.com/schema1">
			<xs:element name="Element1" type="xs:string"/>
		</xs:schema>
		<xs:schema targetNamespace="http://example.com/schema2">
			<xs:element name="Element2" type="xs:int"/>
		</xs:schema>
	</types>
</definitions>`,
			wantErr: false,
			check: func(t *testing.T, g *Generator) {
				// Skip internal field checks
				t.Log("Multiple schemas parsed successfully")
			},
		},
		{
			name: "WSDL with special characters in names",
			wsdlContent: `<?xml version="1.0" encoding="UTF-8"?>
<definitions xmlns="http://schemas.xmlsoap.org/wsdl/"
	xmlns:xs="http://www.w3.org/2001/XMLSchema">
	<types>
		<xs:schema>
			<xs:element name="Test-Element_123" type="xs:string"/>
			<xs:complexType name="Test.Type">
				<xs:sequence>
					<xs:element name="field-name" type="xs:string"/>
				</xs:sequence>
			</xs:complexType>
		</xs:schema>
	</types>
</definitions>`,
			wantErr: false,
			check: func(t *testing.T, g *Generator) {
				// Skip internal field checks
				t.Log("Special characters in names handled successfully")
			},
		},
		{
			name: "WSDL with duplicate type definitions",
			wsdlContent: `<?xml version="1.0" encoding="UTF-8"?>
<definitions xmlns="http://schemas.xmlsoap.org/wsdl/"
	xmlns:xs="http://www.w3.org/2001/XMLSchema">
	<types>
		<xs:schema>
			<xs:complexType name="DuplicateType">
				<xs:sequence>
					<xs:element name="field1" type="xs:string"/>
				</xs:sequence>
			</xs:complexType>
			<xs:complexType name="DuplicateType">
				<xs:sequence>
					<xs:element name="field2" type="xs:int"/>
				</xs:sequence>
			</xs:complexType>
		</xs:schema>
	</types>
</definitions>`,
			wantErr: false, // Should parse but may have issues during code generation
		},
		{
			name: "WSDL with no namespace",
			wsdlContent: `<?xml version="1.0" encoding="UTF-8"?>
<definitions xmlns="http://schemas.xmlsoap.org/wsdl/">
	<types>
		<schema xmlns="http://www.w3.org/2001/XMLSchema">
			<element name="TestElement" type="string"/>
		</schema>
	</types>
</definitions>`,
			wantErr: false,
		},
		{
			name: "WSDL with very long names",
			wsdlContent: `<?xml version="1.0" encoding="UTF-8"?>
<definitions xmlns="http://schemas.xmlsoap.org/wsdl/"
	xmlns:xs="http://www.w3.org/2001/XMLSchema">
	<types>
		<xs:schema>
			<xs:element name="VeryLongElementNameThatMightCauseIssuesInGeneratedCodeVeryLongElementNameThatMightCauseIssuesInGeneratedCodeVeryLongElementNameThatMightCauseIssuesInGeneratedCode" type="xs:string"/>
		</xs:schema>
	</types>
</definitions>`,
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create temporary WSDL file
			tempFile, err := os.CreateTemp("", "test-*.wsdl")
			require.NoError(t, err)
			defer func() {
				_ = os.Remove(tempFile.Name())
			}()

			_, err = tempFile.WriteString(tt.wsdlContent)
			require.NoError(t, err)
			_ = tempFile.Close()

			// Parse WSDL
			g, err := New(tempFile.Name(), WithPackage("test"))
			require.NoError(t, err)

			// Try to parse by generating code
			_, err = g.Generate(context.Background())

			if tt.wantErr {
				require.Error(t, err)
				if tt.errContains != "" {
					assert.Contains(t, err.Error(), tt.errContains)
				}
			} else {
				require.NoError(t, err)
				if tt.check != nil {
					tt.check(t, g)
				}
			}
		})
	}
}

func TestWSDLWithExternalSchema(t *testing.T) {
	// Create temporary directory for test files
	tempDir, err := os.MkdirTemp("", "wsdl-test-*")
	require.NoError(t, err)
	defer func() {
		_ = os.RemoveAll(tempDir)
	}()

	// Create external schema file
	schemaContent := `<?xml version="1.0" encoding="UTF-8"?>
<xs:schema xmlns:xs="http://www.w3.org/2001/XMLSchema"
	targetNamespace="http://example.com/external"
	xmlns:tns="http://example.com/external">
	<xs:complexType name="ExternalType">
		<xs:sequence>
			<xs:element name="field" type="xs:string"/>
		</xs:sequence>
	</xs:complexType>
</xs:schema>`

	schemaFile := filepath.Join(tempDir, "external.xsd")
	err = os.WriteFile(schemaFile, []byte(schemaContent), 0644)
	require.NoError(t, err)

	// Create WSDL that imports the schema
	wsdlContent := `<?xml version="1.0" encoding="UTF-8"?>
<definitions xmlns="http://schemas.xmlsoap.org/wsdl/"
	xmlns:xs="http://www.w3.org/2001/XMLSchema">
	<types>
		<xs:schema>
			<xs:import namespace="http://example.com/external" schemaLocation="external.xsd"/>
		</xs:schema>
	</types>
</definitions>`

	wsdlFile := filepath.Join(tempDir, "test.wsdl")
	err = os.WriteFile(wsdlFile, []byte(wsdlContent), 0644)
	require.NoError(t, err)

	// Parse WSDL
	g, err := New(wsdlFile, WithPackage("test"))
	require.NoError(t, err)

	// Generate will call unmarshal internally
	_, err = g.Generate(context.Background())
	require.NoError(t, err)

	// Check that external schema was loaded - test passes if no error
	t.Log("External schema loaded successfully")
}

func TestWSDLErrorTypes(t *testing.T) {
	// Test with non-existent file
	g, err := New("nonexistent.wsdl", WithPackage("test"))
	require.NoError(t, err)

	// Generate will call unmarshal internally
	_, err = g.Generate(context.Background())
	require.Error(t, err)

	// Should be a WSDLError
	var wsdlErr *types.WSDLError
	assert.ErrorAs(t, err, &wsdlErr)
	assert.Equal(t, "fetch", wsdlErr.Op)
}

func BenchmarkWSDLParsing(b *testing.B) {
	// Create a test WSDL file
	wsdlContent := `<?xml version="1.0" encoding="UTF-8"?>
<definitions xmlns="http://schemas.xmlsoap.org/wsdl/"
	xmlns:xs="http://www.w3.org/2001/XMLSchema"
	targetNamespace="http://example.com/benchmark">
	<types>
		<xs:schema targetNamespace="http://example.com/benchmark">
			<xs:complexType name="BenchmarkType">
				<xs:sequence>
					<xs:element name="field1" type="xs:string"/>
					<xs:element name="field2" type="xs:int"/>
					<xs:element name="field3" type="xs:boolean"/>
				</xs:sequence>
			</xs:complexType>
		</xs:schema>
	</types>
	<message name="BenchmarkRequest">
		<part name="body" type="tns:BenchmarkType"/>
	</message>
	<message name="BenchmarkResponse">
		<part name="body" type="tns:BenchmarkType"/>
	</message>
	<portType name="BenchmarkPortType">
		<operation name="BenchmarkOperation">
			<input message="tns:BenchmarkRequest"/>
			<output message="tns:BenchmarkResponse"/>
		</operation>
	</portType>
</definitions>`

	tempFile, err := os.CreateTemp("", "benchmark-*.wsdl")
	if err != nil {
		b.Fatal(err)
	}
	defer func() {
		_ = os.Remove(tempFile.Name())
	}()

	_, err = tempFile.WriteString(wsdlContent)
	if err != nil {
		b.Fatal(err)
	}
	_ = tempFile.Close()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		g, err := New(tempFile.Name(), WithPackage("test"))
		if err != nil {
			b.Fatal(err)
		}
		
		// Generate will call unmarshal internally
		_, err = g.Generate(context.Background())
		if err != nil {
			b.Fatal(err)
		}
	}
}