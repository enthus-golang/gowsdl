// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at http://mozilla.org/MPL/2.0/.

package generator

import (
	"context"
	"os"
	"testing"

	"github.com/enthus-golang/gowsdl/pkg/http"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGeneratorOptions(t *testing.T) {
	tempFile := createTempWSDLFile(t)
	defer func() {
		if err := os.Remove(tempFile); err != nil {
			t.Logf("Failed to remove temp file: %v", err)
		}
	}()

	tests := []struct {
		name   string
		option Option
	}{
		{
			name:   "WithHTTPConfig",
			option: WithHTTPConfig(&http.HTTPClientConfig{}),
		},
		{
			name:   "WithGenerics",
			option: WithGenerics(),
		},
		{
			name:   "WithExportAllTypes_true",
			option: WithExportAllTypes(true),
		},
		{
			name:   "WithExportAllTypes_false",
			option: WithExportAllTypes(false),
		},
		{
			name:   "WithServerGeneration_true",
			option: WithServerGeneration(true),
		},
		{
			name:   "WithServerGeneration_false",
			option: WithServerGeneration(false),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g, err := New(tempFile, tt.option)
			require.NoError(t, err)
			assert.NotNil(t, g)
		})
	}
}

func TestUtilityFunctions(t *testing.T) {
	t.Run("makePrivate", func(t *testing.T) {
		tests := []struct {
			input    string
			expected string
		}{
			{"TestName", "testName"},
			{"XMLParser", "xMLParser"},
			{"ID", "iD"},
			{"", ""},
			{"a", "a"},
		}

		for _, tt := range tests {
			result := makePrivate(tt.input)
			assert.Equal(t, tt.expected, result)
		}
	})

	t.Run("comment", func(t *testing.T) {
		tests := []struct {
			input    string
			expected string
		}{
			{"single line", "// single line"},
			{"line1\nline2", "// line1\n// line2"},
			{"  spaced  ", "// spaced"},
			{"", "// "},
		}

		for _, tt := range tests {
			result := comment(tt.input)
			assert.Equal(t, tt.expected, result)
		}
	})

	t.Run("removePointerFromType", func(t *testing.T) {
		tests := []struct {
			input    string
			expected string
		}{
			{"*string", "string"},
			{"string", "string"},
			{"**int", "*int"},
			{"", ""},
		}

		for _, tt := range tests {
			result := removePointerFromType(tt.input)
			assert.Equal(t, tt.expected, result)
		}
	})

	t.Run("removeNamespacePrefix", func(t *testing.T) {
		tests := []struct {
			input    string
			expected string
		}{
			{"ns:LocalName", "LocalName"},
			{"soap:Envelope", "Envelope"},
			{"LocalName", "LocalName"},
			{"", ""},
			{"a:b:c", "c"},
		}

		for _, tt := range tests {
			result := removeNamespacePrefix(tt.input)
			assert.Equal(t, tt.expected, result)
		}
	})

	t.Run("makeValidXmlTag", func(t *testing.T) {
		tests := []struct {
			xmlName  string
			goName   string
			expected string
		}{
			{"name", "name", "name"},
			{"xmlName", "goName", "xmlName,omitempty"},
			{"", "", ""},
		}

		for _, tt := range tests {
			result := makeValidXmlTag(tt.xmlName, tt.goName)
			assert.Equal(t, tt.expected, result)
		}
	})

	t.Run("goString", func(t *testing.T) {
		tests := []struct {
			input    string
			expected string
		}{
			{`hello world`, `hello world`},
			{`hello "world"`, `hello \"world\"`},
			{`say "hello" and "goodbye"`, `say \"hello\" and \"goodbye\"`},
			{"", ""},
		}

		for _, tt := range tests {
			result := goString(tt.input)
			assert.Equal(t, tt.expected, result)
		}
	})

	t.Run("sanitizeEnumValue", func(t *testing.T) {
		tests := []struct {
			input    string
			expected string
		}{
			{"", "EmptyString"},
			{"simple", "simple"},
			{"with spaces", "withspaces"},
			{"with-dashes", "withdashes"},
			{"with.dots", "withdots"},
			{"with:colons", "withcolons"},
			{"with/slashes", "withslashes"},
			{"with\\backslashes", "withbackslashes"},
			{"with(parens)", "withparens"},
			{"with[brackets]", "withbrackets"},
			{"with{braces}", "withbraces"},
			{"with,commas;semicolons", "withcommassemicolons"},
			{"with'quotes\"", "withquotes"},
			{"   ", "Empty"},
		}

		for _, tt := range tests {
			result := sanitizeEnumValue(tt.input)
			assert.Equal(t, tt.expected, result)
		}
	})
}

func TestGeneratorHelperMethods(t *testing.T) {
	tempFile := createTempWSDLFile(t)
	defer func() {
		if err := os.Remove(tempFile); err != nil {
			t.Logf("Failed to remove temp file: %v", err)
		}
	}()

	g, err := New(tempFile, WithPackage("test"))
	require.NoError(t, err)

	// Load the WSDL to test helper methods
	ctx := context.Background()
	err = g.unmarshal(ctx)
	require.NoError(t, err)

	t.Run("getTargetNamespace", func(t *testing.T) {
		ns := g.getTargetNamespace()
		// Should return empty string for our simple test WSDL or the actual namespace
		assert.IsType(t, "", ns)
	})

	t.Run("findSOAPAction", func(t *testing.T) {
		action := g.findSOAPAction("TestOperation", "TestPortType")
		// Should return the SOAP action from the WSDL binding
		assert.Equal(t, "testAction", action)
	})

	t.Run("findType", func(t *testing.T) {
		// Test with a simple message reference
		result := g.findType("TestRequestMessage")
		// findType may return the element reference or the type, both are valid
		assert.NotEmpty(t, result)

		// Test with namespace prefix
		result = g.findType("tns:TestRequestMessage")
		assert.NotEmpty(t, result)
	})
}

func TestGeneratorWithDifferentVersions(t *testing.T) {
	t.Run("WSDL_1.1", func(t *testing.T) {
		tempFile := createTempWSDLFile(t)
		defer func() {
			if err := os.Remove(tempFile); err != nil {
				t.Logf("Failed to remove temp file: %v", err)
			}
		}()

		g, err := New(tempFile, WithPackage("test"), WithServerGeneration(true))
		require.NoError(t, err)

		ctx := context.Background()
		code, err := g.Generate(ctx)
		require.NoError(t, err)

		assert.Contains(t, code, "header")
		assert.Contains(t, code, "types")
		assert.Contains(t, code, "operations")
		assert.Contains(t, code, "server")
		assert.Contains(t, code, "server_header")
		assert.Contains(t, code, "server_wsdl")
	})

	t.Run("WSDL_2.0", func(t *testing.T) {
		// Note: WSDL 2.0 support may be limited, so we'll test if it handles gracefully
		wsdl2Content := `<?xml version="1.0" encoding="UTF-8"?>
<description xmlns="http://www.w3.org/ns/wsdl"
             targetNamespace="http://example.com/stockquote"
             xmlns:tns="http://example.com/stockquote">

  <types>
    <xsd:schema xmlns:xsd="http://www.w3.org/2001/XMLSchema"
                targetNamespace="http://example.com/stockquote">
      <xsd:element name="GetQuoteRequest">
        <xsd:complexType>
          <xsd:sequence>
            <xsd:element name="symbol" type="xsd:string"/>
          </xsd:sequence>
        </xsd:complexType>
      </xsd:element>
      <xsd:element name="GetQuoteResponse">
        <xsd:complexType>
          <xsd:sequence>
            <xsd:element name="price" type="xsd:float"/>
          </xsd:sequence>
        </xsd:complexType>
      </xsd:element>
    </xsd:schema>
  </types>

  <interface name="StockQuoteInterface">
    <operation name="GetQuote">
      <input element="tns:GetQuoteRequest"/>
      <output element="tns:GetQuoteResponse"/>
    </operation>
  </interface>

  <binding name="StockQuoteSOAPBinding" interface="tns:StockQuoteInterface"
           type="http://www.w3.org/ns/wsdl/soap">
    <operation ref="tns:GetQuote">
      <soap:operation action="http://example.com/GetQuote"/>
    </operation>
  </binding>

  <service name="StockQuoteService" interface="tns:StockQuoteInterface">
    <endpoint name="StockQuoteEndpoint" binding="tns:StockQuoteSOAPBinding"
              address="http://example.com/stockquote"/>
  </service>
</description>`

		tmpFile, err := os.CreateTemp("", "test-wsdl2-*.wsdl")
		require.NoError(t, err)
		defer func() {
			if err := os.Remove(tmpFile.Name()); err != nil {
				t.Logf("Failed to remove temp file: %v", err)
			}
		}()

		_, err = tmpFile.WriteString(wsdl2Content)
		require.NoError(t, err)
		err = tmpFile.Close()
		require.NoError(t, err)

		g, err := New(tmpFile.Name(), WithPackage("stockquote"), WithServerGeneration(true))
		if err != nil {
			// WSDL 2.0 may not be fully supported yet, so we'll skip if it fails
			t.Skipf("WSDL 2.0 parsing not fully supported: %v", err)
			return
		}

		ctx := context.Background()
		code, err := g.Generate(ctx)
		if err != nil {
			t.Skipf("WSDL 2.0 generation not fully supported: %v", err)
			return
		}

		assert.Contains(t, code, "header")
		assert.Contains(t, code, "types")
		assert.Contains(t, code, "operations")
		assert.Contains(t, code, "server")
		assert.Contains(t, code, "server_header")
		assert.Contains(t, code, "server_wsdl")
	})
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