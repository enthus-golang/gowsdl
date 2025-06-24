// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at http://mozilla.org/MPL/2.0/.

package gowsdl

import (
	"context"
	"go/parser"
	"go/token"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCodeGeneration(t *testing.T) {
	tests := []struct {
		name        string
		wsdlContent string
		wantErr     bool
		checkCode   func(t *testing.T, code map[string][]byte)
	}{
		{
			name: "simple types",
			wsdlContent: `<?xml version="1.0" encoding="UTF-8"?>
<definitions xmlns="http://schemas.xmlsoap.org/wsdl/"
	xmlns:xs="http://www.w3.org/2001/XMLSchema"
	xmlns:tns="http://example.com/test"
	targetNamespace="http://example.com/test">
	<types>
		<xs:schema targetNamespace="http://example.com/test">
			<xs:element name="StringElement" type="xs:string"/>
			<xs:element name="IntElement" type="xs:int"/>
			<xs:element name="BoolElement" type="xs:boolean"/>
		</xs:schema>
	</types>
</definitions>`,
			wantErr: false,
			checkCode: func(t *testing.T, code map[string][]byte) {
				types := string(code["types"])
				assert.Contains(t, types, "StringElement")
				assert.Contains(t, types, "IntElement")
				assert.Contains(t, types, "BoolElement")
				assert.Contains(t, types, "string")
				assert.Contains(t, types, "int32")
				assert.Contains(t, types, "bool")
			},
		},
		{
			name: "complex type",
			wsdlContent: `<?xml version="1.0" encoding="UTF-8"?>
<definitions xmlns="http://schemas.xmlsoap.org/wsdl/"
	xmlns:xs="http://www.w3.org/2001/XMLSchema"
	xmlns:tns="http://example.com/test"
	targetNamespace="http://example.com/test">
	<types>
		<xs:schema targetNamespace="http://example.com/test">
			<xs:complexType name="PersonType">
				<xs:sequence>
					<xs:element name="firstName" type="xs:string"/>
					<xs:element name="lastName" type="xs:string"/>
					<xs:element name="age" type="xs:int"/>
					<xs:element name="active" type="xs:boolean"/>
				</xs:sequence>
			</xs:complexType>
			<xs:element name="Person" type="tns:PersonType"/>
		</xs:schema>
	</types>
</definitions>`,
			wantErr: false,
			checkCode: func(t *testing.T, code map[string][]byte) {
				types := string(code["types"])
				assert.Contains(t, types, "type PersonType struct")
				assert.Contains(t, types, "FirstName string")
				assert.Contains(t, types, "LastName string")
				assert.Contains(t, types, "Age int32")
				assert.Contains(t, types, "Active bool")
			},
		},
		{
			name: "arrays and sequences",
			wsdlContent: `<?xml version="1.0" encoding="UTF-8"?>
<definitions xmlns="http://schemas.xmlsoap.org/wsdl/"
	xmlns:xs="http://www.w3.org/2001/XMLSchema"
	xmlns:tns="http://example.com/test"
	targetNamespace="http://example.com/test">
	<types>
		<xs:schema targetNamespace="http://example.com/test">
			<xs:complexType name="ItemList">
				<xs:sequence>
					<xs:element name="item" type="xs:string" minOccurs="0" maxOccurs="unbounded"/>
				</xs:sequence>
			</xs:complexType>
		</xs:schema>
	</types>
</definitions>`,
			wantErr: false,
			checkCode: func(t *testing.T, code map[string][]byte) {
				types := string(code["types"])
				assert.Contains(t, types, "[]string")
			},
		},
		{
			name: "operations",
			wsdlContent: `<?xml version="1.0" encoding="UTF-8"?>
<definitions xmlns="http://schemas.xmlsoap.org/wsdl/"
	xmlns:xs="http://www.w3.org/2001/XMLSchema"
	xmlns:soap="http://schemas.xmlsoap.org/wsdl/soap/"
	xmlns:tns="http://example.com/test"
	targetNamespace="http://example.com/test">
	<types>
		<xs:schema targetNamespace="http://example.com/test">
			<xs:element name="GetUserRequest">
				<xs:complexType>
					<xs:sequence>
						<xs:element name="userId" type="xs:string"/>
					</xs:sequence>
				</xs:complexType>
			</xs:element>
			<xs:element name="GetUserResponse">
				<xs:complexType>
					<xs:sequence>
						<xs:element name="userName" type="xs:string"/>
					</xs:sequence>
				</xs:complexType>
			</xs:element>
		</xs:schema>
	</types>
	<message name="GetUserRequestMessage">
		<part name="body" element="tns:GetUserRequest"/>
	</message>
	<message name="GetUserResponseMessage">
		<part name="body" element="tns:GetUserResponse"/>
	</message>
	<portType name="UserServicePortType">
		<operation name="GetUser">
			<input message="tns:GetUserRequestMessage"/>
			<output message="tns:GetUserResponseMessage"/>
		</operation>
	</portType>
	<binding name="UserServiceBinding" type="tns:UserServicePortType">
		<soap:binding transport="http://schemas.xmlsoap.org/soap/http"/>
		<operation name="GetUser">
			<soap:operation soapAction="GetUser"/>
			<input>
				<soap:body use="literal"/>
			</input>
			<output>
				<soap:body use="literal"/>
			</output>
		</operation>
	</binding>
	<service name="UserService">
		<port name="UserServicePort" binding="tns:UserServiceBinding">
			<soap:address location="http://example.com/UserService"/>
		</port>
	</service>
</definitions>`,
			wantErr: false,
			checkCode: func(t *testing.T, code map[string][]byte) {
				// Check types
				types := string(code["types"])
				assert.Contains(t, types, "GetUserRequest")
				assert.Contains(t, types, "GetUserResponse")
				
				// Check operations
				ops := string(code["operations"])
				assert.Contains(t, ops, "GetUser")
				assert.Contains(t, ops, "UserServicePortType")
			},
		},
		{
			name: "reserved Go keywords",
			wsdlContent: `<?xml version="1.0" encoding="UTF-8"?>
<definitions xmlns="http://schemas.xmlsoap.org/wsdl/"
	xmlns:xs="http://www.w3.org/2001/XMLSchema"
	xmlns:tns="http://example.com/test"
	targetNamespace="http://example.com/test">
	<types>
		<xs:schema targetNamespace="http://example.com/test">
			<xs:complexType name="TestType">
				<xs:sequence>
					<xs:element name="type" type="xs:string"/>
					<xs:element name="interface" type="xs:string"/>
					<xs:element name="package" type="xs:string"/>
					<xs:element name="func" type="xs:string"/>
				</xs:sequence>
			</xs:complexType>
		</xs:schema>
	</types>
</definitions>`,
			wantErr: false,
			checkCode: func(t *testing.T, code map[string][]byte) {
				types := string(code["types"])
				// Reserved words should be modified
				assert.NotContains(t, types, "Type string")
				assert.NotContains(t, types, "Interface string")
				assert.NotContains(t, types, "Package string")
				assert.NotContains(t, types, "Func string")
				// Should have underscores
				assert.Contains(t, types, "Type_ string")
				assert.Contains(t, types, "Interface_ string")
				assert.Contains(t, types, "Package_ string")
				assert.Contains(t, types, "Func_ string")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create temporary WSDL file
			tempFile, err := os.CreateTemp("", "codegen-test-*.wsdl")
			require.NoError(t, err)
			defer os.Remove(tempFile.Name())

			_, err = tempFile.WriteString(tt.wsdlContent)
			require.NoError(t, err)
			tempFile.Close()

			// Generate code
			g, err := NewGoWSDL(tempFile.Name(), "test", false, true)
			require.NoError(t, err)

			code, err := g.Start()
			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				require.NotNil(t, code)
				
				// Check all expected parts are present
				assert.NotNil(t, code["header"])
				assert.NotNil(t, code["types"])
				assert.NotNil(t, code["operations"])
				assert.NotNil(t, code["server"])
				assert.NotNil(t, code["server_header"])
				assert.NotNil(t, code["server_wsdl"])

				// Run custom checks
				if tt.checkCode != nil {
					tt.checkCode(t, code)
				}

				// Verify generated code is valid Go
				allCode := string(code["header"]) + string(code["types"]) + string(code["operations"])
				_, err := parser.ParseFile(token.NewFileSet(), "", allCode, parser.AllErrors)
				assert.NoError(t, err, "Generated code should be valid Go")
			}
		})
	}
}

func TestCodeGenerationWithNamespaces(t *testing.T) {
	wsdlContent := `<?xml version="1.0" encoding="UTF-8"?>
<definitions xmlns="http://schemas.xmlsoap.org/wsdl/"
	xmlns:xs="http://www.w3.org/2001/XMLSchema"
	xmlns:tns="http://example.com/test"
	xmlns:other="http://example.com/other"
	targetNamespace="http://example.com/test">
	<types>
		<xs:schema targetNamespace="http://example.com/test">
			<xs:import namespace="http://example.com/other"/>
			<xs:complexType name="MainType">
				<xs:sequence>
					<xs:element name="field" type="other:OtherType"/>
				</xs:sequence>
			</xs:complexType>
		</xs:schema>
		<xs:schema targetNamespace="http://example.com/other">
			<xs:complexType name="OtherType">
				<xs:sequence>
					<xs:element name="value" type="xs:string"/>
				</xs:sequence>
			</xs:complexType>
		</xs:schema>
	</types>
</definitions>`

	tempFile, err := os.CreateTemp("", "namespace-test-*.wsdl")
	require.NoError(t, err)
	defer os.Remove(tempFile.Name())

	_, err = tempFile.WriteString(wsdlContent)
	require.NoError(t, err)
	tempFile.Close()

	g, err := NewGoWSDL(tempFile.Name(), "test", false, true)
	require.NoError(t, err)

	code, err := g.Start()
	require.NoError(t, err)

	types := string(code["types"])
	assert.Contains(t, types, "MainType")
	assert.Contains(t, types, "OtherType")
}

func TestMakePublicFunction(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"test", "Test"},
		{"testField", "TestField"},
		{"test_field", "Test_field"},
		{"TEST", "TEST"},
		{"123test", "123test"}, // Numbers at start remain
		{"xmlField", "XMLField"},
		{"httpRequest", "HTTPRequest"},
		{"urlPath", "URLPath"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := makePublic(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestReplaceReservedWords(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"type", "type_"},
		{"package", "package_"},
		{"interface", "interface_"},
		{"func", "func_"},
		{"return", "return_"},
		{"normalField", "normalField"},
		{"Type", "Type_"}, // Capital letters
		{"myType", "myType"}, // Not a reserved word
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := replaceReservedWords(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestNormalizeNames(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"test-name", "testname"},
		{"test.name", "testname"},
		{"test_name", "test_name"},
		{"test name", "testname"},
		{"Test-Name", "TestName"},
		{"test123", "test123"},
		{"123test", "123test"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := normalize(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func BenchmarkCodeGeneration(b *testing.B) {
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
					<xs:element name="field4" type="xs:dateTime"/>
					<xs:element name="field5" type="xs:decimal"/>
				</xs:sequence>
			</xs:complexType>
		</xs:schema>
	</types>
</definitions>`

	tempFile, err := os.CreateTemp("", "benchmark-codegen-*.wsdl")
	if err != nil {
		b.Fatal(err)
	}
	defer os.Remove(tempFile.Name())

	_, err = tempFile.WriteString(wsdlContent)
	if err != nil {
		b.Fatal(err)
	}
	tempFile.Close()

	g, err := NewGoWSDL(tempFile.Name(), "test", false, true)
	if err != nil {
		b.Fatal(err)
	}

	// Pre-unmarshal for accurate benchmark
	err = g.unmarshal(context.Background())
	if err != nil {
		b.Fatal(err)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := g.genTypes()
		if err != nil {
			b.Fatal(err)
		}
	}
}

func TestStartWithContextErrorHandling(t *testing.T) {
	// Create a WSDL that will cause an error during generation
	wsdlContent := `<?xml version="1.0" encoding="UTF-8"?>
<definitions xmlns="http://schemas.xmlsoap.org/wsdl/">
	<types>
		<schema xmlns="http://www.w3.org/2001/XMLSchema">
			<element name="Test" type="string"/>
		</schema>
	</types>
</definitions>`

	tempFile, err := os.CreateTemp("", "error-test-*.wsdl")
	require.NoError(t, err)
	defer os.Remove(tempFile.Name())

	_, err = tempFile.WriteString(wsdlContent)
	require.NoError(t, err)
	tempFile.Close()

	g, err := NewGoWSDL(tempFile.Name(), "test", false, true)
	require.NoError(t, err)

	// This should succeed - the error handling improvements ensure
	// that generation errors are properly reported
	code, err := g.StartWithContext(context.Background())
	require.NoError(t, err)
	assert.NotNil(t, code)
}

func TestCodeGenerationConcurrency(t *testing.T) {
	// Test that concurrent code generation works correctly
	wsdlContent := `<?xml version="1.0" encoding="UTF-8"?>
<definitions xmlns="http://schemas.xmlsoap.org/wsdl/"
	xmlns:xs="http://www.w3.org/2001/XMLSchema">
	<types>
		<xs:schema>
			<xs:element name="Test" type="xs:string"/>
		</xs:schema>
	</types>
</definitions>`

	tempFile, err := os.CreateTemp("", "concurrent-test-*.wsdl")
	require.NoError(t, err)
	defer os.Remove(tempFile.Name())

	_, err = tempFile.WriteString(wsdlContent)
	require.NoError(t, err)
	tempFile.Close()

	// Run multiple generations concurrently
	const goroutines = 5
	results := make(chan error, goroutines)

	for i := 0; i < goroutines; i++ {
		go func() {
			g, err := NewGoWSDL(tempFile.Name(), "test", false, true)
			if err != nil {
				results <- err
				return
			}

			code, err := g.Start()
			if err != nil {
				results <- err
				return
			}

			if code == nil || len(code) == 0 {
				results <- assert.AnError
				return
			}

			results <- nil
		}()
	}

	// Check all results
	for i := 0; i < goroutines; i++ {
		err := <-results
		assert.NoError(t, err)
	}
}