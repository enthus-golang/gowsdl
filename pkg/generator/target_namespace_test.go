package generator

import (
	"context"
	"os"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTargetNamespaceMethodGeneration(t *testing.T) {
	tests := []struct {
		name                   string
		wsdl                   string
		expectedInCheckV1      bool
		expectedInResponse     bool
		expectedNamespace      string
		elementFormDefault     string
	}{
		{
			name: "unqualified_schema_with_message_elements",
			wsdl: `<?xml version="1.0" encoding="UTF-8"?>
<definitions xmlns="http://schemas.xmlsoap.org/wsdl/"
             xmlns:soap="http://schemas.xmlsoap.org/wsdl/soap/"
             xmlns:tns="http://example.com/test"
             targetNamespace="http://example.com/test">
    <types>
        <schema xmlns="http://www.w3.org/2001/XMLSchema"
                targetNamespace="http://example.com/test"
                elementFormDefault="unqualified">
            <element name="TestRequest" type="tns:TestRequestType"/>
            <element name="TestResponse" type="tns:TestResponseType"/>
            <complexType name="TestRequestType">
                <sequence>
                    <element name="input" type="string"/>
                </sequence>
            </complexType>
            <complexType name="TestResponseType">
                <sequence>
                    <element name="output" type="string"/>
                </sequence>
            </complexType>
        </schema>
    </types>
    <message name="TestRequestMessage">
        <part name="body" element="tns:TestRequest"/>
    </message>
    <message name="TestResponseMessage">
        <part name="body" element="tns:TestResponse"/>
    </message>
    <portType name="TestPortType">
        <operation name="Test">
            <input message="tns:TestRequestMessage"/>
            <output message="tns:TestResponseMessage"/>
        </operation>
    </portType>
</definitions>`,
			expectedInCheckV1:  true,
			expectedInResponse: true,
			expectedNamespace:  "http://example.com/test",
			elementFormDefault: "unqualified",
		},
		{
			name: "qualified_schema_no_target_namespace_method",
			wsdl: `<?xml version="1.0" encoding="UTF-8"?>
<definitions xmlns="http://schemas.xmlsoap.org/wsdl/"
             xmlns:soap="http://schemas.xmlsoap.org/wsdl/soap/"
             xmlns:tns="http://example.com/qualified"
             targetNamespace="http://example.com/qualified">
    <types>
        <schema xmlns="http://www.w3.org/2001/XMLSchema"
                targetNamespace="http://example.com/qualified"
                elementFormDefault="qualified">
            <element name="QualifiedRequest" type="tns:QualifiedRequestType"/>
            <element name="QualifiedResponse" type="tns:QualifiedResponseType"/>
            <complexType name="QualifiedRequestType">
                <sequence>
                    <element name="data" type="string"/>
                </sequence>
            </complexType>
            <complexType name="QualifiedResponseType">
                <sequence>
                    <element name="result" type="string"/>
                </sequence>
            </complexType>
        </schema>
    </types>
    <message name="QualifiedRequestMessage">
        <part name="body" element="tns:QualifiedRequest"/>
    </message>
    <message name="QualifiedResponseMessage">
        <part name="body" element="tns:QualifiedResponse"/>
    </message>
    <portType name="QualifiedPortType">
        <operation name="Process">
            <input message="tns:QualifiedRequestMessage"/>
            <output message="tns:QualifiedResponseMessage"/>
        </operation>
    </portType>
</definitions>`,
			expectedInCheckV1:  false, // No TargetNamespace method for qualified schemas
			expectedInResponse: false,
			expectedNamespace:  "http://example.com/qualified",
			elementFormDefault: "qualified",
		},
		{
			name: "complex_types_not_used_in_messages",
			wsdl: `<?xml version="1.0" encoding="UTF-8"?>
<definitions xmlns="http://schemas.xmlsoap.org/wsdl/"
             xmlns:soap="http://schemas.xmlsoap.org/wsdl/soap/"
             xmlns:tns="http://example.com/unused"
             targetNamespace="http://example.com/unused">
    <types>
        <schema xmlns="http://www.w3.org/2001/XMLSchema"
                targetNamespace="http://example.com/unused"
                elementFormDefault="unqualified">
            <complexType name="UnusedType">
                <sequence>
                    <element name="field" type="string"/>
                </sequence>
            </complexType>
            <element name="UsedRequest" type="tns:UsedRequestType"/>
            <complexType name="UsedRequestType">
                <sequence>
                    <element name="input" type="string"/>
                </sequence>
            </complexType>
        </schema>
    </types>
    <message name="UsedRequestMessage">
        <part name="body" element="tns:UsedRequest"/>
    </message>
    <portType name="TestPortType">
        <operation name="Test">
            <input message="tns:UsedRequestMessage"/>
        </operation>
    </portType>
</definitions>`,
			expectedInCheckV1:  true, // UsedRequestType should have TargetNamespace
			expectedInResponse: false, // UnusedType should NOT have TargetNamespace
			expectedNamespace:  "http://example.com/unused",
			elementFormDefault: "unqualified",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a temporary file with the WSDL content
			tempFile, err := os.CreateTemp("", "test-*.wsdl")
			require.NoError(t, err)
			defer os.Remove(tempFile.Name())
			defer tempFile.Close()
			
			_, err = tempFile.Write([]byte(tt.wsdl))
			require.NoError(t, err)
			
			// Create generator with the temp file
			g, err := New(tempFile.Name(), WithPackage("test"))
			require.NoError(t, err)
			
			files, err := g.Generate(context.Background())
			require.NoError(t, err)
			require.NotNil(t, files)
			
			typesContent, exists := files["types"]
			require.True(t, exists, "types file should be generated")
			
			typesStr := string(typesContent)
			
			// Check for TargetNamespace method based on test expectations
			if tt.expectedInCheckV1 {
				// For unqualified schemas, we expect TargetNamespace method
				assert.Contains(t, typesStr, "func (t ", "Should contain method receiver")
				assert.Contains(t, typesStr, "TargetNamespace() string", "Should contain TargetNamespace method signature")
				assert.Contains(t, typesStr, "return \""+tt.expectedNamespace+"\"", "Should return correct namespace")
			}
			
			// Verify that qualified schemas don't generate TargetNamespace method
			if tt.elementFormDefault == "qualified" {
				assert.NotContains(t, typesStr, "TargetNamespace() string", "Qualified schemas should not have TargetNamespace method")
			}
		})
	}
}

func TestTargetNamespaceMethodForMessageElements(t *testing.T) {
	// Test specifically for the checkV1 scenario from the user's WSDL
	wsdl := `<?xml version="1.0" encoding="UTF-8"?>
<definitions xmlns="http://schemas.xmlsoap.org/wsdl/"
             xmlns:soap="http://schemas.xmlsoap.org/wsdl/soap/"
             xmlns:tns="http://b2b.also.com/ActWebServices.PriceAvailability"
             targetNamespace="http://b2b.also.com/ActWebServices.PriceAvailability">
    <types>
        <schema xmlns="http://www.w3.org/2001/XMLSchema"
                targetNamespace="http://b2b.also.com/ActWebServices.PriceAvailability"
                elementFormDefault="unqualified">
            <element name="checkV1" type="tns:CheckV1Type"/>
            <element name="checkV1Response" type="tns:CheckV1ResponseType"/>
            <complexType name="CheckV1Type">
                <sequence>
                    <element name="PartnerCountryCode" type="string"/>
                    <element name="StockType" type="string"/>
                </sequence>
            </complexType>
            <complexType name="CheckV1ResponseType">
                <sequence>
                    <element name="ErrorMessage" type="string" minOccurs="0"/>
                </sequence>
            </complexType>
        </schema>
    </types>
    <message name="checkV1Request">
        <part name="parameters" element="tns:checkV1"/>
    </message>
    <message name="checkV1Response">
        <part name="parameters" element="tns:checkV1Response"/>
    </message>
    <portType name="PriceAvailabilityPortType">
        <operation name="checkV1">
            <input message="tns:checkV1Request"/>
            <output message="tns:checkV1Response"/>
        </operation>
    </portType>
</definitions>`

	// Create a temporary file with the WSDL content
	tempFile, err := os.CreateTemp("", "test-*.wsdl")
	require.NoError(t, err)
	defer os.Remove(tempFile.Name())
	defer tempFile.Close()
	
	_, err = tempFile.Write([]byte(wsdl))
	require.NoError(t, err)
	
	// Create generator with the temp file
	g, err := New(tempFile.Name(), WithPackage("test"))
	require.NoError(t, err)
	
	files, err := g.Generate(context.Background())
	require.NoError(t, err)
	
	typesContent := string(files["types"])
	
	// Check that the wrapper types have TargetNamespace method
	// The generator creates element wrapper types (CheckV1) that embed the complex types (CheckV1Type)
	assert.Contains(t, typesContent, "type CheckV1 struct")
	assert.Contains(t, typesContent, "type CheckV1Type struct")
	
	// Check for the TargetNamespace method on the wrapper type CheckV1
	assert.Contains(t, typesContent, "func (t CheckV1) TargetNamespace() string", "CheckV1 should have TargetNamespace method")
	assert.Contains(t, typesContent, `return "http://b2b.also.com/ActWebServices.PriceAvailability"`, "Should return correct namespace")
	
	// Check that CheckV1Response also has TargetNamespace method
	assert.Contains(t, typesContent, "type CheckV1Response struct")
	assert.Contains(t, typesContent, "func (t CheckV1Response) TargetNamespace() string", "CheckV1Response should have TargetNamespace method")
	
	// Verify XMLName is set correctly on wrapper types
	assert.Contains(t, typesContent, `XMLName xml.Name `+"`"+`xml:"tns:checkV1"`+"`", "CheckV1 should have XMLName with tns: prefix")
	assert.Contains(t, typesContent, `XMLName xml.Name `+"`"+`xml:"checkV1Response"`+"`", "CheckV1Response should have XMLName without prefix")
}

func TestTargetNamespaceIntegrationWithSOAPClient(t *testing.T) {
	// Test that generated types with TargetNamespace method work with SOAP client
	wsdl := `<?xml version="1.0" encoding="UTF-8"?>
<definitions xmlns="http://schemas.xmlsoap.org/wsdl/"
             xmlns:soap="http://schemas.xmlsoap.org/wsdl/soap/"
             xmlns:tns="http://integration.test/soap"
             targetNamespace="http://integration.test/soap">
    <types>
        <schema xmlns="http://www.w3.org/2001/XMLSchema"
                targetNamespace="http://integration.test/soap"
                elementFormDefault="unqualified">
            <element name="Echo" type="tns:EchoType"/>
            <element name="EchoResponse" type="tns:EchoResponseType"/>
            <complexType name="EchoType">
                <sequence>
                    <element name="Message" type="string"/>
                </sequence>
            </complexType>
            <complexType name="EchoResponseType">
                <sequence>
                    <element name="Result" type="string"/>
                </sequence>
            </complexType>
        </schema>
    </types>
    <message name="EchoRequest">
        <part name="parameters" element="tns:Echo"/>
    </message>
    <message name="EchoResponse">
        <part name="parameters" element="tns:EchoResponse"/>
    </message>
    <portType name="EchoServicePortType">
        <operation name="Echo">
            <input message="tns:EchoRequest"/>
            <output message="tns:EchoResponse"/>
        </operation>
    </portType>
    <binding name="EchoServiceBinding" type="tns:EchoServicePortType">
        <soap:binding style="document" transport="http://schemas.xmlsoap.org/soap/http"/>
        <operation name="Echo">
            <soap:operation soapAction="Echo"/>
            <input>
                <soap:body use="literal"/>
            </input>
            <output>
                <soap:body use="literal"/>
            </output>
        </operation>
    </binding>
    <service name="EchoService">
        <port name="EchoServicePort" binding="tns:EchoServiceBinding">
            <soap:address location="http://localhost:8080/echo"/>
        </port>
    </service>
</definitions>`

	// Create a temporary file with the WSDL content
	tempFile, err := os.CreateTemp("", "test-*.wsdl")
	require.NoError(t, err)
	defer os.Remove(tempFile.Name())
	defer tempFile.Close()
	
	_, err = tempFile.Write([]byte(wsdl))
	require.NoError(t, err)
	
	// Create generator with the temp file
	g, err := New(tempFile.Name(), WithPackage("test"))
	require.NoError(t, err)
	
	files, err := g.Generate(context.Background())
	require.NoError(t, err)
	
	// Verify types file contains TargetNamespace method
	typesContent := string(files["types"])
	// The TargetNamespace method is on the wrapper type Echo, not EchoType
	assert.Contains(t, typesContent, "func (t Echo) TargetNamespace() string")
	assert.Contains(t, typesContent, `return "http://integration.test/soap"`)
	
	// Verify XMLName is generated with proper namespace prefix
	assert.Contains(t, typesContent, `XMLName xml.Name `+"`"+`xml:"tns:Echo"`+"`")
	
	// Verify response uses local name for flexibility
	assert.Contains(t, typesContent, `XMLName xml.Name `+"`"+`xml:"EchoResponse"`+"`")
}

func TestTargetNamespaceNotGeneratedForNonMessageTypes(t *testing.T) {
	// Ensure TargetNamespace is NOT generated for types that aren't used in messages
	wsdl := `<?xml version="1.0" encoding="UTF-8"?>
<definitions xmlns="http://schemas.xmlsoap.org/wsdl/"
             xmlns:soap="http://schemas.xmlsoap.org/wsdl/soap/"
             xmlns:tns="http://example.com/types"
             targetNamespace="http://example.com/types">
    <types>
        <schema xmlns="http://www.w3.org/2001/XMLSchema"
                targetNamespace="http://example.com/types"
                elementFormDefault="unqualified">
            <!-- Type used in message -->
            <element name="UsedElement" type="tns:UsedType"/>
            <complexType name="UsedType">
                <sequence>
                    <element name="field1" type="string"/>
                </sequence>
            </complexType>
            
            <!-- Type NOT used in message -->
            <complexType name="UnusedType">
                <sequence>
                    <element name="field2" type="string"/>
                </sequence>
            </complexType>
            
            <!-- Nested type -->
            <complexType name="NestedContainer">
                <sequence>
                    <element name="nested" type="tns:UnusedType"/>
                </sequence>
            </complexType>
        </schema>
    </types>
    <message name="TestMessage">
        <part name="body" element="tns:UsedElement"/>
    </message>
    <portType name="TestPortType">
        <operation name="TestOp">
            <input message="tns:TestMessage"/>
        </operation>
    </portType>
</definitions>`

	// Create a temporary file with the WSDL content
	tempFile, err := os.CreateTemp("", "test-*.wsdl")
	require.NoError(t, err)
	defer os.Remove(tempFile.Name())
	defer tempFile.Close()
	
	_, err = tempFile.Write([]byte(wsdl))
	require.NoError(t, err)
	
	// Create generator with the temp file
	g, err := New(tempFile.Name(), WithPackage("test"))
	require.NoError(t, err)
	
	files, err := g.Generate(context.Background())
	require.NoError(t, err)
	
	typesContent := string(files["types"])
	
	// The wrapper type UsedElement (created from the element) should have TargetNamespace method
	assert.Contains(t, typesContent, "type UsedElement struct")
	assert.Contains(t, typesContent, "func (t UsedElement) TargetNamespace() string", "UsedElement should have TargetNamespace")
	
	// Check that UnusedType does NOT have TargetNamespace method
	unusedTypeIdx := strings.Index(typesContent, "type UnusedType struct")
	require.Greater(t, unusedTypeIdx, -1)
	
	// Find the next type or func definition after UnusedType to limit our search
	afterUnusedType := typesContent[unusedTypeIdx:]
	
	// Look for the closing brace of the struct
	braceIdx := strings.Index(afterUnusedType, "}")
	if braceIdx > 0 && braceIdx < 500 { // Reasonable limit for a struct definition
		// Check the area right after the struct definition
		checkArea := afterUnusedType[:braceIdx+100]
		assert.NotContains(t, checkArea, "func (t UnusedType) TargetNamespace() string", "UnusedType should NOT have TargetNamespace")
	}
}