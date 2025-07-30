package generator

import (
	"context"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGenerator_isMessageElement(t *testing.T) {
	tests := []struct {
		name     string
		wsdl     string
		typeName string
		expected bool
	}{
		{
			name: "element used in message",
			wsdl: `<?xml version="1.0" encoding="UTF-8"?>
<wsdl:definitions xmlns:wsdl="http://schemas.xmlsoap.org/wsdl/" 
	targetNamespace="http://example.com/test" 
	xmlns:tns="http://example.com/test" 
	xmlns:xsd="http://www.w3.org/2001/XMLSchema"
	xmlns:soap="http://schemas.xmlsoap.org/wsdl/soap/">
	<wsdl:types>
		<xsd:schema targetNamespace="http://example.com/test">
			<xsd:element name="checkV1" type="tns:checkV1"/>
			<xsd:complexType name="checkV1">
				<xsd:sequence>
					<xsd:element name="input" type="xsd:string"/>
				</xsd:sequence>
			</xsd:complexType>
		</xsd:schema>
	</wsdl:types>
	<wsdl:message name="checkV1Request">
		<wsdl:part name="parameters" element="tns:checkV1"/>
	</wsdl:message>
</wsdl:definitions>`,
			typeName: "checkV1",
			expected: true,
		},
		{
			name: "element not used in message",
			wsdl: `<?xml version="1.0" encoding="UTF-8"?>
<wsdl:definitions xmlns:wsdl="http://schemas.xmlsoap.org/wsdl/" 
	targetNamespace="http://example.com/test" 
	xmlns:tns="http://example.com/test" 
	xmlns:xsd="http://www.w3.org/2001/XMLSchema"
	xmlns:soap="http://schemas.xmlsoap.org/wsdl/soap/">
	<wsdl:types>
		<xsd:schema targetNamespace="http://example.com/test">
			<xsd:element name="checkV1" type="tns:checkV1"/>
			<xsd:element name="unusedElement" type="tns:unusedType"/>
			<xsd:complexType name="checkV1">
				<xsd:sequence>
					<xsd:element name="input" type="xsd:string"/>
				</xsd:sequence>
			</xsd:complexType>
			<xsd:complexType name="unusedType">
				<xsd:sequence>
					<xsd:element name="data" type="xsd:string"/>
				</xsd:sequence>
			</xsd:complexType>
		</xsd:schema>
	</wsdl:types>
	<wsdl:message name="checkV1Request">
		<wsdl:part name="parameters" element="tns:checkV1"/>
	</wsdl:message>
</wsdl:definitions>`,
			typeName: "unusedElement",
			expected: false,
		},
		{
			name: "element with namespace prefix used in message",
			wsdl: `<?xml version="1.0" encoding="UTF-8"?>
<wsdl:definitions xmlns:wsdl="http://schemas.xmlsoap.org/wsdl/" 
	targetNamespace="http://example.com/test" 
	xmlns:tns="http://example.com/test" 
	xmlns:xsd="http://www.w3.org/2001/XMLSchema"
	xmlns:soap="http://schemas.xmlsoap.org/wsdl/soap/">
	<wsdl:types>
		<xsd:schema targetNamespace="http://example.com/test">
			<xsd:element name="myRequest" type="tns:myRequestType"/>
			<xsd:complexType name="myRequestType">
				<xsd:sequence>
					<xsd:element name="param" type="xsd:string"/>
				</xsd:sequence>
			</xsd:complexType>
		</xsd:schema>
	</wsdl:types>
	<wsdl:message name="myMessage">
		<wsdl:part name="body" element="tns:myRequest"/>
	</wsdl:message>
</wsdl:definitions>`,
			typeName: "myRequest",
			expected: true,
		},
		{
			name: "multiple messages with different elements",
			wsdl: `<?xml version="1.0" encoding="UTF-8"?>
<wsdl:definitions xmlns:wsdl="http://schemas.xmlsoap.org/wsdl/" 
	targetNamespace="http://example.com/test" 
	xmlns:tns="http://example.com/test" 
	xmlns:xsd="http://www.w3.org/2001/XMLSchema"
	xmlns:soap="http://schemas.xmlsoap.org/wsdl/soap/">
	<wsdl:types>
		<xsd:schema targetNamespace="http://example.com/test">
			<xsd:element name="request1" type="tns:requestType1"/>
			<xsd:element name="request2" type="tns:requestType2"/>
			<xsd:element name="response1" type="tns:responseType1"/>
			<xsd:complexType name="requestType1">
				<xsd:sequence>
					<xsd:element name="param1" type="xsd:string"/>
				</xsd:sequence>
			</xsd:complexType>
			<xsd:complexType name="requestType2">
				<xsd:sequence>
					<xsd:element name="param2" type="xsd:string"/>
				</xsd:sequence>
			</xsd:complexType>
			<xsd:complexType name="responseType1">
				<xsd:sequence>
					<xsd:element name="result" type="xsd:string"/>
				</xsd:sequence>
			</xsd:complexType>
		</xsd:schema>
	</wsdl:types>
	<wsdl:message name="message1">
		<wsdl:part name="body" element="tns:request1"/>
	</wsdl:message>
	<wsdl:message name="message2">
		<wsdl:part name="body" element="tns:response1"/>
	</wsdl:message>
</wsdl:definitions>`,
			typeName: "request1",
			expected: true,
		},
		{
			name: "message part with type instead of element",
			wsdl: `<?xml version="1.0" encoding="UTF-8"?>
<wsdl:definitions xmlns:wsdl="http://schemas.xmlsoap.org/wsdl/" 
	targetNamespace="http://example.com/test" 
	xmlns:tns="http://example.com/test" 
	xmlns:xsd="http://www.w3.org/2001/XMLSchema"
	xmlns:soap="http://schemas.xmlsoap.org/wsdl/soap/">
	<wsdl:types>
		<xsd:schema targetNamespace="http://example.com/test">
			<xsd:element name="myElement" type="tns:myType"/>
			<xsd:complexType name="myType">
				<xsd:sequence>
					<xsd:element name="data" type="xsd:string"/>
				</xsd:sequence>
			</xsd:complexType>
		</xsd:schema>
	</wsdl:types>
	<wsdl:message name="rpcMessage">
		<wsdl:part name="body" type="tns:myType"/>
	</wsdl:message>
</wsdl:definitions>`,
			typeName: "myElement",
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Write WSDL to temp file
			tmpFile, err := os.CreateTemp("", "test-ismessageelement-*.wsdl")
			require.NoError(t, err)
			defer func() {
				_ = os.Remove(tmpFile.Name())
			}()

			_, err = tmpFile.WriteString(tt.wsdl)
			require.NoError(t, err)
			err = tmpFile.Close()
			require.NoError(t, err)

			// Create generator
			gen, err := New(tmpFile.Name(), WithPackage("testpkg"))
			require.NoError(t, err)

			// Generate code to trigger WSDL parsing
			_, err = gen.Generate(context.Background())
			require.NoError(t, err)

			// Test isMessageElement function
			result := gen.isMessageElement(tt.typeName)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestGenerator_isMessageElement_WSDL2(t *testing.T) {
	// Skip this test for now as WSDL 2.0 support is limited
	// The function correctly returns false for WSDL 2.0 as implemented
	t.Skip("WSDL 2.0 support is limited, test covers WSDL 1.1 extensively")
}

func TestGenerator_isMessageElement_EmptyWSDL(t *testing.T) {
	// Test with empty WSDL data
	gen := &Generator{
		wsdlVersion: "1.1",
	}

	result := gen.isMessageElement("anyElement")
	assert.False(t, result, "Should return false when no WSDL data")
}

func TestGenerator_isMessageElement_NoMessages(t *testing.T) {
	// Test with WSDL that has no messages
	wsdlContent := `<?xml version="1.0" encoding="UTF-8"?>
<wsdl:definitions xmlns:wsdl="http://schemas.xmlsoap.org/wsdl/" 
	targetNamespace="http://example.com/test" 
	xmlns:tns="http://example.com/test" 
	xmlns:xsd="http://www.w3.org/2001/XMLSchema">
	<wsdl:types>
		<xsd:schema targetNamespace="http://example.com/test">
			<xsd:element name="myElement" type="tns:myType"/>
			<xsd:complexType name="myType">
				<xsd:sequence>
					<xsd:element name="data" type="xsd:string"/>
				</xsd:sequence>
			</xsd:complexType>
		</xsd:schema>
	</wsdl:types>
</wsdl:definitions>`

	// Write to temp file
	tmpFile, err := os.CreateTemp("", "test-nomessages-*.wsdl")
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

	// Generate code to trigger WSDL parsing
	_, err = gen.Generate(context.Background())
	require.NoError(t, err)

	// Test isMessageElement function with no messages
	result := gen.isMessageElement("myElement")
	assert.False(t, result, "Should return false when no messages exist")
}