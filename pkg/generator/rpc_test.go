package generator

import (
	"context"
	"os"
	"testing"

	"github.com/enthus-golang/gowsdl/pkg/parser"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRPCStyleWSDLGeneration(t *testing.T) {
	// WSDL content with RPC-style operations having multiple primitive-typed parts
	wsdl := `<?xml version="1.0" encoding="UTF-8"?>
<definitions xmlns="http://schemas.xmlsoap.org/wsdl/" 
             xmlns:soap="http://schemas.xmlsoap.org/wsdl/soap/" 
             xmlns:tns="http://example.com/rpc"
             xmlns:xsd="http://www.w3.org/2001/XMLSchema"
             targetNamespace="http://example.com/rpc">
    
    <types>
        <xsd:schema targetNamespace="http://example.com/rpc">
            <xsd:complexType name="SearchResult">
                <xsd:sequence>
                    <xsd:element name="count" type="xsd:int"/>
                    <xsd:element name="results" type="xsd:string" maxOccurs="unbounded"/>
                </xsd:sequence>
            </xsd:complexType>
        </xsd:schema>
    </types>
    
    <message name="searchRequest">
        <part name="query" type="xsd:string"/>
        <part name="maxResults" type="xsd:int"/>
        <part name="category" type="xsd:string"/>
    </message>
    
    <message name="searchResponse">
        <part name="result" type="tns:SearchResult"/>
    </message>
    
    <message name="singlePartPrimitive">
        <part name="id" type="xsd:string"/>
    </message>
    
    <portType name="SearchService">
        <operation name="search">
            <input message="tns:searchRequest"/>
            <output message="tns:searchResponse"/>
        </operation>
    </portType>
    
    <binding name="SearchServiceBinding" type="tns:SearchService">
        <soap:binding style="rpc" transport="http://schemas.xmlsoap.org/soap/http"/>
        <operation name="search">
            <soap:operation soapAction="http://example.com/rpc/search"/>
            <input>
                <soap:body use="literal"/>
            </input>
            <output>
                <soap:body use="literal"/>
            </output>
        </operation>
    </binding>
    
    <service name="SearchService">
        <port name="SearchServicePort" binding="tns:SearchServiceBinding">
            <soap:address location="http://example.com/rpc/search"/>
        </port>
    </service>
</definitions>`

	// Write WSDL to a temporary file
	tmpFile := t.TempDir() + "/test.wsdl"
	err := os.WriteFile(tmpFile, []byte(wsdl), 0644)
	require.NoError(t, err)

	gen, err := New(tmpFile, WithPackage("test"))
	require.NoError(t, err)
	
	code, err := gen.Generate(context.TODO())
	require.NoError(t, err)
	
	// Combine all generated code for testing
	var generatedCode string
	for _, content := range code {
		generatedCode += string(content) + "\n"
	}
	
	// Check that wrapper types are generated for RPC-style messages
	assert.Contains(t, generatedCode, "type SearchRequest struct {")
	assert.Contains(t, generatedCode, "Query string `xml:\"query\" json:\"query\"`")
	assert.Contains(t, generatedCode, "MaxResults int32 `xml:\"maxResults\" json:\"maxResults\"`")
	assert.Contains(t, generatedCode, "Category string `xml:\"category\" json:\"category\"`")
	
	// Check that the operations use the generated types
	assert.Contains(t, generatedCode, "Search(ctx context.Context, request *SearchRequest) (*SearchResponse, error)")
	
	// Ensure no undefined String_ type is referenced
	assert.NotContains(t, generatedCode, "String_")
	
	// Check for single part primitive message handling
	assert.Contains(t, generatedCode, "type SinglePartPrimitive struct {")
	assert.Contains(t, generatedCode, "Id string `xml:\"id\" json:\"id\"`")
}

func TestDocumentStyleWSDLNotAffected(t *testing.T) {
	// WSDL content with document-style operations
	wsdl := `<?xml version="1.0" encoding="UTF-8"?>
<definitions xmlns="http://schemas.xmlsoap.org/wsdl/" 
             xmlns:soap="http://schemas.xmlsoap.org/wsdl/soap/" 
             xmlns:tns="http://example.com/doc"
             xmlns:xsd="http://www.w3.org/2001/XMLSchema"
             targetNamespace="http://example.com/doc">
    
    <types>
        <xsd:schema targetNamespace="http://example.com/doc">
            <xsd:element name="GetUserRequest">
                <xsd:complexType>
                    <xsd:sequence>
                        <xsd:element name="userId" type="xsd:string"/>
                    </xsd:sequence>
                </xsd:complexType>
            </xsd:element>
            
            <xsd:element name="GetUserResponse">
                <xsd:complexType>
                    <xsd:sequence>
                        <xsd:element name="user" type="tns:User"/>
                    </xsd:sequence>
                </xsd:complexType>
            </xsd:element>
            
            <xsd:complexType name="User">
                <xsd:sequence>
                    <xsd:element name="id" type="xsd:string"/>
                    <xsd:element name="name" type="xsd:string"/>
                </xsd:sequence>
            </xsd:complexType>
        </xsd:schema>
    </types>
    
    <message name="GetUserRequestMessage">
        <part name="parameters" element="tns:GetUserRequest"/>
    </message>
    
    <message name="GetUserResponseMessage">
        <part name="parameters" element="tns:GetUserResponse"/>
    </message>
    
    <portType name="UserService">
        <operation name="getUser">
            <input message="tns:GetUserRequestMessage"/>
            <output message="tns:GetUserResponseMessage"/>
        </operation>
    </portType>
    
    <binding name="UserServiceBinding" type="tns:UserService">
        <soap:binding style="document" transport="http://schemas.xmlsoap.org/soap/http"/>
        <operation name="getUser">
            <soap:operation soapAction="http://example.com/doc/getUser"/>
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
            <soap:address location="http://example.com/doc/user"/>
        </port>
    </service>
</definitions>`

	// Write WSDL to a temporary file
	tmpFile := t.TempDir() + "/test.wsdl"
	err := os.WriteFile(tmpFile, []byte(wsdl), 0644)
	require.NoError(t, err)

	gen, err := New(tmpFile, WithPackage("test"))
	require.NoError(t, err)
	
	code, err := gen.Generate(context.TODO())
	require.NoError(t, err)
	
	// Combine all generated code for testing
	var generatedCode string
	for _, content := range code {
		generatedCode += string(content) + "\n"
	}
	
	// Check that document-style operations still work correctly
	assert.Contains(t, generatedCode, "type GetUserRequest struct {")
	assert.Contains(t, generatedCode, "UserId string")
	assert.Contains(t, generatedCode, "GetUser(ctx context.Context, request *GetUserRequest) (*GetUserResponse, error)")
	
	// Document-style messages should not generate wrapper types
	assert.NotContains(t, generatedCode, "type GetUserRequestMessage struct")
}

func TestIsRPCStyleMessage(t *testing.T) {
	gen := &Generator{}
	
	// Test cases
	tests := []struct {
		name     string
		message  *parser.WSDLMessage
		expected bool
	}{
		{
			name:     "nil message",
			message:  nil,
			expected: false,
		},
		{
			name: "multiple parts",
			message: &parser.WSDLMessage{
				Parts: []*parser.WSDLPart{
					{Name: "param1", Type: "xsd:string"},
					{Name: "param2", Type: "xsd:int"},
				},
			},
			expected: true,
		},
		{
			name: "single primitive type part",
			message: &parser.WSDLMessage{
				Parts: []*parser.WSDLPart{
					{Name: "id", Type: "xsd:string"},
				},
			},
			expected: true,
		},
		{
			name: "single element part",
			message: &parser.WSDLMessage{
				Parts: []*parser.WSDLPart{
					{Name: "body", Element: "tns:RequestBody"},
				},
			},
			expected: false,
		},
		{
			name: "empty parts",
			message: &parser.WSDLMessage{
				Parts: []*parser.WSDLPart{},
			},
			expected: false,
		},
	}
	
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			result := gen.isRPCStyleMessage(tc.message)
			assert.Equal(t, tc.expected, result)
		})
	}
}

func TestMessageTypeGeneration(t *testing.T) {
	// Test that different message types generate the correct wrapper types
	wsdl := `<?xml version="1.0" encoding="UTF-8"?>
<definitions xmlns="http://schemas.xmlsoap.org/wsdl/" 
             xmlns:soap="http://schemas.xmlsoap.org/wsdl/soap/"
             xmlns:tns="http://test.com"
             xmlns:xsd="http://www.w3.org/2001/XMLSchema"
             targetNamespace="http://test.com">
    
    <types>
        <xsd:schema targetNamespace="http://test.com">
            <xsd:element name="RequestBody">
                <xsd:complexType>
                    <xsd:sequence>
                        <xsd:element name="data" type="xsd:string"/>
                    </xsd:sequence>
                </xsd:complexType>
            </xsd:element>
        </xsd:schema>
    </types>
    
    <message name="multiPartMessage">
        <part name="param1" type="xsd:string"/>
        <part name="param2" type="xsd:int"/>
    </message>
    
    <message name="singlePrimitiveMessage">
        <part name="id" type="xsd:string"/>
    </message>
    
    <message name="singleElementMessage">
        <part name="body" element="tns:RequestBody"/>
    </message>
    
    <portType name="TestService">
        <operation name="multiPartOp">
            <input message="tns:multiPartMessage"/>
            <output message="tns:multiPartMessage"/>
        </operation>
        <operation name="singlePrimitiveOp">
            <input message="tns:singlePrimitiveMessage"/>
            <output message="tns:singlePrimitiveMessage"/>
        </operation>
        <operation name="singleElementOp">
            <input message="tns:singleElementMessage"/>
            <output message="tns:singleElementMessage"/>
        </operation>
    </portType>
    
    <binding name="TestServiceBinding" type="tns:TestService">
        <soap:binding style="rpc" transport="http://schemas.xmlsoap.org/soap/http"/>
        <operation name="multiPartOp">
            <soap:operation soapAction=""/>
        </operation>
        <operation name="singlePrimitiveOp">
            <soap:operation soapAction=""/>
        </operation>
        <operation name="singleElementOp">
            <soap:operation soapAction=""/>
        </operation>
    </binding>
</definitions>`

	tmpFile := t.TempDir() + "/test.wsdl"
	err := os.WriteFile(tmpFile, []byte(wsdl), 0644)
	require.NoError(t, err)

	gen, err := New(tmpFile, WithPackage("test"))
	require.NoError(t, err)
	
	code, err := gen.Generate(context.TODO())
	require.NoError(t, err)
	
	// Combine all generated code for testing
	var generatedCode string
	for _, content := range code {
		generatedCode += string(content) + "\n"
	}
	
	// Check that wrapper types are generated for RPC-style messages
	assert.Contains(t, generatedCode, "type MultiPartMessage struct {")
	assert.Contains(t, generatedCode, "Param1 string `xml:\"param1\" json:\"param1\"`")
	assert.Contains(t, generatedCode, "Param2 int32 `xml:\"param2\" json:\"param2\"`")
	
	assert.Contains(t, generatedCode, "type SinglePrimitiveMessage struct {")
	assert.Contains(t, generatedCode, "Id string `xml:\"id\" json:\"id\"`")
	
	// Document-style message should not generate wrapper type, but should have RequestBody type
	assert.NotContains(t, generatedCode, "type SingleElementMessage struct {")
	assert.Contains(t, generatedCode, "type RequestBody struct {")
	
	// Check operations use correct types
	assert.Contains(t, generatedCode, "multiPartOp(ctx context.Context, request *MultiPartMessage) (*MultiPartMessage, error)")
	assert.Contains(t, generatedCode, "singlePrimitiveOp(ctx context.Context, request *SinglePrimitiveMessage) (*SinglePrimitiveMessage, error)")
	assert.Contains(t, generatedCode, "singleElementOp(ctx context.Context, request *RequestBody) (*RequestBody, error)")
}