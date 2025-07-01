// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at http://mozilla.org/MPL/2.0/.

package gowsdl

import (
	"context"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestWithOptions(t *testing.T) {
	tests := []struct {
		name   string
		option Option
	}{
		{
			name: "WithHTTPConfig",
			option: WithHTTPConfig(&HTTPClientConfig{
				Timeout: 30,
			}),
		},
		{
			name:   "WithGenerics",
			option: WithGenerics(),
		},
		{
			name:   "WithExportAllTypes",
			option: WithExportAllTypes(true),
		},
		{
			name:   "WithPackage",
			option: WithPackage("testpkg"),
		},
		{
			name:   "WithServerGeneration",
			option: WithServerGeneration(true),
		},
	}

	// Create a temporary WSDL file for testing
	tempFile := createTempWSDLFile(t)
	defer func() {
		if err := os.Remove(tempFile); err != nil {
			t.Logf("Failed to remove temp file: %v", err)
		}
	}()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gowsdl, err := NewGoWSDL(tempFile, tt.option)
			require.NoError(t, err)
			assert.NotNil(t, gowsdl)
			assert.NotNil(t, gowsdl.generator)
		})
	}
}

func TestNewGoWSDL(t *testing.T) {
	tests := []struct {
		name    string
		file    string
		opts    []Option
		wantErr bool
	}{
		{
			name: "valid_wsdl_file",
			file: createTempWSDLFile(t),
			opts: []Option{WithPackage("test")},
		},
		{
			name:    "empty_file_path",
			file:    "",
			wantErr: true,
		},
		{
			name:    "whitespace_file_path",
			file:    "   ",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.name == "valid_wsdl_file" {
				defer func() {
					if err := os.Remove(tt.file); err != nil {
						t.Logf("Failed to remove temp file: %v", err)
					}
				}()
			}

			gowsdl, err := NewGoWSDL(tt.file, tt.opts...)
			if tt.wantErr {
				require.Error(t, err)
				assert.Nil(t, gowsdl)
			} else {
				require.NoError(t, err)
				assert.NotNil(t, gowsdl)
				assert.NotNil(t, gowsdl.generator)
			}
		})
	}
}

func TestNewGoWSDLWithConfig(t *testing.T) {
	tempFile := createTempWSDLFile(t)
	defer func() {
		if err := os.Remove(tempFile); err != nil {
			t.Logf("Failed to remove temp file: %v", err)
		}
	}()

	httpConfig := &HTTPClientConfig{
		Timeout: 30,
	}

	gowsdl, err := NewGoWSDLWithConfig(tempFile, "testpkg", httpConfig, true)
	require.NoError(t, err)
	assert.NotNil(t, gowsdl)
	assert.NotNil(t, gowsdl.generator)
}

func TestNewGoWSDLWithOptions(t *testing.T) {
	tempFile := createTempWSDLFile(t)
	defer func() {
		if err := os.Remove(tempFile); err != nil {
			t.Logf("Failed to remove temp file: %v", err)
		}
	}()

	httpConfig := &HTTPClientConfig{
		Timeout: 30,
	}

	tests := []struct {
		name        string
		useGenerics bool
	}{
		{
			name:        "with_generics",
			useGenerics: true,
		},
		{
			name:        "without_generics",
			useGenerics: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gowsdl, err := NewGoWSDLWithOptions(tempFile, "testpkg", httpConfig, true, tt.useGenerics)
			require.NoError(t, err)
			assert.NotNil(t, gowsdl)
			assert.NotNil(t, gowsdl.generator)
		})
	}
}

func TestGoWSDLStart(t *testing.T) {
	tempFile := createTempWSDLFile(t)
	defer func() {
		if err := os.Remove(tempFile); err != nil {
			t.Logf("Failed to remove temp file: %v", err)
		}
	}()

	gowsdl, err := NewGoWSDL(tempFile, WithPackage("test"))
	require.NoError(t, err)

	code, err := gowsdl.Start()
	require.NoError(t, err)
	assert.NotNil(t, code)
	assert.Contains(t, code, "header")
	assert.Contains(t, code, "types")
	assert.Contains(t, code, "operations")
}

func TestGoWSDLStartWithContext(t *testing.T) {
	tempFile := createTempWSDLFile(t)
	defer func() {
		if err := os.Remove(tempFile); err != nil {
			t.Logf("Failed to remove temp file: %v", err)
		}
	}()

	gowsdl, err := NewGoWSDL(tempFile, WithPackage("test"))
	require.NoError(t, err)

	ctx := context.Background()
	code, err := gowsdl.StartWithContext(ctx)
	require.NoError(t, err)
	assert.NotNil(t, code)
	assert.Contains(t, code, "header")
	assert.Contains(t, code, "types")
	assert.Contains(t, code, "operations")
}

func TestGoWSDLWithServerGeneration(t *testing.T) {
	tempFile := createTempWSDLFile(t)
	defer func() {
		if err := os.Remove(tempFile); err != nil {
			t.Logf("Failed to remove temp file: %v", err)
		}
	}()

	gowsdl, err := NewGoWSDL(tempFile, WithPackage("test"), WithServerGeneration(true))
	require.NoError(t, err)

	code, err := gowsdl.Start()
	require.NoError(t, err)
	assert.NotNil(t, code)
	assert.Contains(t, code, "header")
	assert.Contains(t, code, "types")
	assert.Contains(t, code, "operations")
	assert.Contains(t, code, "server")
	assert.Contains(t, code, "server_header")
	assert.Contains(t, code, "server_wsdl")
}

func createTempWSDLFile(t *testing.T) string {
	t.Helper()

	wsdlContent := `<?xml version="1.0" encoding="UTF-8"?>
<wsdl:definitions xmlns:wsdl="http://schemas.xmlsoap.org/wsdl/"
	xmlns:soap="http://schemas.xmlsoap.org/wsdl/soap/"
	xmlns:tns="http://example.com/test"
	targetNamespace="http://example.com/test">

	<wsdl:types>
		<xsd:schema xmlns:xsd="http://www.w3.org/2001/XMLSchema"
			targetNamespace="http://example.com/test">
			<xsd:element name="TestRequest">
				<xsd:complexType>
					<xsd:sequence>
						<xsd:element name="input" type="xsd:string"/>
					</xsd:sequence>
				</xsd:complexType>
			</xsd:element>
			<xsd:element name="TestResponse">
				<xsd:complexType>
					<xsd:sequence>
						<xsd:element name="output" type="xsd:string"/>
					</xsd:sequence>
				</xsd:complexType>
			</xsd:element>
		</xsd:schema>
	</wsdl:types>

	<wsdl:message name="TestRequestMessage">
		<wsdl:part name="parameters" element="tns:TestRequest"/>
	</wsdl:message>
	<wsdl:message name="TestResponseMessage">
		<wsdl:part name="parameters" element="tns:TestResponse"/>
	</wsdl:message>

	<wsdl:portType name="TestPortType">
		<wsdl:operation name="TestOperation">
			<wsdl:input message="tns:TestRequestMessage"/>
			<wsdl:output message="tns:TestResponseMessage"/>
		</wsdl:operation>
	</wsdl:portType>

	<wsdl:binding name="TestBinding" type="tns:TestPortType">
		<soap:binding style="document" transport="http://schemas.xmlsoap.org/soap/http"/>
		<wsdl:operation name="TestOperation">
			<soap:operation soapAction="testAction"/>
			<wsdl:input><soap:body use="literal"/></wsdl:input>
			<wsdl:output><soap:body use="literal"/></wsdl:output>
		</wsdl:operation>
	</wsdl:binding>

	<wsdl:service name="TestService">
		<wsdl:port name="TestPort" binding="tns:TestBinding">
			<soap:address location="http://example.com/test"/>
		</wsdl:port>
	</wsdl:service>
</wsdl:definitions>`

	tmpFile, err := os.CreateTemp("", "test-*.wsdl")
	require.NoError(t, err)

	_, err = tmpFile.WriteString(wsdlContent)
	require.NoError(t, err)

	err = tmpFile.Close()
	require.NoError(t, err)

	return tmpFile.Name()
}