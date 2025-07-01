package generator

import (
	"context"
	"os"
	"testing"

	"github.com/enthus-golang/gowsdl/pkg/parser"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSOAPActionInGeneratedCode(t *testing.T) {
	// Create a temporary WSDL file
	wsdlContent := `<?xml version="1.0" encoding="UTF-8"?>
<wsdl:definitions xmlns:soap="http://schemas.xmlsoap.org/wsdl/soap/" 
                  xmlns:wsdl="http://schemas.xmlsoap.org/wsdl/" 
                  xmlns:xsd="http://www.w3.org/2001/XMLSchema" 
                  targetNamespace="http://example.com/test" 
                  xmlns:tns="http://example.com/test" 
                  name="TestService">
	
	<wsdl:types>
		<xsd:schema targetNamespace="http://example.com/test">
			<xsd:element name="TestRequest" type="xsd:string"/>
			<xsd:element name="TestResponse" type="xsd:string"/>
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
			<soap:operation soapAction="http://example.com/test/TestOperation"/>
			<wsdl:input>
				<soap:body use="literal"/>
			</wsdl:input>
			<wsdl:output>
				<soap:body use="literal"/>
			</wsdl:output>
		</wsdl:operation>
	</wsdl:binding>
	
	<wsdl:service name="TestService">
		<wsdl:port name="TestPort" binding="tns:TestBinding">
			<soap:address location="http://example.com/test"/>
		</wsdl:port>
	</wsdl:service>
</wsdl:definitions>`

	// Write to temp file
	tmpFile, err := os.CreateTemp("", "test-*.wsdl")
	require.NoError(t, err)
	defer func() {
		_ = os.Remove(tmpFile.Name())
	}()

	_, err = tmpFile.WriteString(wsdlContent)
	require.NoError(t, err)
	err = tmpFile.Close()
	require.NoError(t, err)

	// Create generator
	gen, err := New(tmpFile.Name(), WithPackage("testpkg"))
	require.NoError(t, err)

	// Generate code
	result, err := gen.Generate(context.Background())
	require.NoError(t, err)

	// Check that operations file contains SOAP action
	operationsCode, ok := result["operations"]
	require.True(t, ok, "operations file should be generated")

	// Verify SOAP action is included in CallContext
	assert.Contains(t, string(operationsCode), `CallContext(ctx, "http://example.com/test/TestOperation", request, response)`)
}

func TestFindSOAPAction(t *testing.T) {
	tests := []struct {
		name        string
		operation   string
		portType    string
		wsdl        *parser.WSDL
		expected    string
	}{
		{
			name:      "finds soap action for matching operation",
			operation: "TestOp",
			portType:  "TestPortType",
			wsdl: &parser.WSDL{
				Binding: []*parser.WSDLBinding{
					{
						Name: "TestBinding",
						Type: "tns:TestPortType",
						Operations: []*parser.WSDLOperation{
							{
								Name: "TestOp",
								SOAPOperation: parser.WSDLSOAPOperation{
									SOAPAction: "http://example.com/TestOp",
								},
							},
						},
					},
				},
			},
			expected: "http://example.com/TestOp",
		},
		{
			name:      "handles port type without suffix",
			operation: "TestOp",
			portType:  "TestPort",
			wsdl: &parser.WSDL{
				Binding: []*parser.WSDLBinding{
					{
						Name: "TestBinding",
						Type: "tns:TestPortType",
						Operations: []*parser.WSDLOperation{
							{
								Name: "TestOp",
								SOAPOperation: parser.WSDLSOAPOperation{
									SOAPAction: "http://example.com/TestOp",
								},
							},
						},
					},
				},
			},
			expected: "http://example.com/TestOp",
		},
		{
			name:      "returns empty string when operation not found",
			operation: "NonExistent",
			portType:  "TestPortType",
			wsdl: &parser.WSDL{
				Binding: []*parser.WSDLBinding{
					{
						Name: "TestBinding",
						Type: "tns:TestPortType",
						Operations: []*parser.WSDLOperation{
							{
								Name: "TestOp",
								SOAPOperation: parser.WSDLSOAPOperation{
									SOAPAction: "http://example.com/TestOp",
								},
							},
						},
					},
				},
			},
			expected: "",
		},
		{
			name:      "returns empty string when wsdl is nil",
			operation: "TestOp",
			portType:  "TestPortType",
			wsdl:      nil,
			expected:  "",
		},
		{
			name:      "handles empty soap action",
			operation: "TestOp",
			portType:  "TestPortType",
			wsdl: &parser.WSDL{
				Binding: []*parser.WSDLBinding{
					{
						Name: "TestBinding",
						Type: "tns:TestPortType",
						Operations: []*parser.WSDLOperation{
							{
								Name: "TestOp",
								SOAPOperation: parser.WSDLSOAPOperation{
									SOAPAction: "",
								},
							},
						},
					},
				},
			},
			expected: "",
		},
		{
			name:      "handles complex suffix variations",
			operation: "TestOp",
			portType:  "MyService",
			wsdl: &parser.WSDL{
				Binding: []*parser.WSDLBinding{
					{
						Name: "MyServiceBinding",
						Type: "tns:MyServicePortType",
						Operations: []*parser.WSDLOperation{
							{
								Name: "TestOp",
								SOAPOperation: parser.WSDLSOAPOperation{
									SOAPAction: "http://example.com/MyServiceTestOp",
								},
							},
						},
					},
				},
			},
			expected: "http://example.com/MyServiceTestOp",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gen := &Generator{wsdl: tt.wsdl}
			result := gen.findSOAPAction(tt.operation, tt.portType)
			assert.Equal(t, tt.expected, result)
		})
	}
}