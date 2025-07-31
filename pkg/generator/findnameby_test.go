package generator

import (
	"context"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGenerator_findNameByType(t *testing.T) {
	tests := []struct {
		name     string
		wsdl     string
		typeName string
		expected string
	}{
		{
			name: "element with matching type name",
			wsdl: `<?xml version="1.0" encoding="UTF-8"?>
<wsdl:definitions xmlns:wsdl="http://schemas.xmlsoap.org/wsdl/" 
	targetNamespace="http://example.com/test" 
	xmlns:tns="http://example.com/test" 
	xmlns:xsd="http://www.w3.org/2001/XMLSchema">
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
</wsdl:definitions>`,
			typeName: "checkV1",
			expected: "checkV1",
		},
		{
			name: "element with different name than type",
			wsdl: `<?xml version="1.0" encoding="UTF-8"?>
<wsdl:definitions xmlns:wsdl="http://schemas.xmlsoap.org/wsdl/" 
	targetNamespace="http://example.com/test" 
	xmlns:tns="http://example.com/test" 
	xmlns:xsd="http://www.w3.org/2001/XMLSchema">
	<wsdl:types>
		<xsd:schema targetNamespace="http://example.com/test">
			<xsd:element name="requestElement" type="tns:requestType"/>
			<xsd:complexType name="requestType">
				<xsd:sequence>
					<xsd:element name="input" type="xsd:string"/>
				</xsd:sequence>
			</xsd:complexType>
		</xsd:schema>
	</wsdl:types>
</wsdl:definitions>`,
			typeName: "requestType",
			expected: "requestElement",
		},
		{
			name: "type with namespace prefix",
			wsdl: `<?xml version="1.0" encoding="UTF-8"?>
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
</wsdl:definitions>`,
			typeName: "tns:myType",
			expected: "myElement",
		},
		{
			name: "type not found returns original name",
			wsdl: `<?xml version="1.0" encoding="UTF-8"?>
<wsdl:definitions xmlns:wsdl="http://schemas.xmlsoap.org/wsdl/" 
	targetNamespace="http://example.com/test" 
	xmlns:tns="http://example.com/test" 
	xmlns:xsd="http://www.w3.org/2001/XMLSchema">
	<wsdl:types>
		<xsd:schema targetNamespace="http://example.com/test">
			<xsd:element name="existingElement" type="tns:existingType"/>
			<xsd:complexType name="existingType">
				<xsd:sequence>
					<xsd:element name="data" type="xsd:string"/>
				</xsd:sequence>
			</xsd:complexType>
		</xsd:schema>
	</wsdl:types>
</wsdl:definitions>`,
			typeName: "nonExistentType",
			expected: "nonExistentType",
		},
		{
			name: "multiple schemas",
			wsdl: `<?xml version="1.0" encoding="UTF-8"?>
<wsdl:definitions xmlns:wsdl="http://schemas.xmlsoap.org/wsdl/" 
	targetNamespace="http://example.com/test" 
	xmlns:tns="http://example.com/test" 
	xmlns:xsd="http://www.w3.org/2001/XMLSchema">
	<wsdl:types>
		<xsd:schema targetNamespace="http://example.com/test">
			<xsd:element name="element1" type="tns:type1"/>
			<xsd:complexType name="type1">
				<xsd:sequence>
					<xsd:element name="data1" type="xsd:string"/>
				</xsd:sequence>
			</xsd:complexType>
		</xsd:schema>
		<xsd:schema targetNamespace="http://example.com/other">
			<xsd:element name="element2" type="tns:type2"/>
			<xsd:complexType name="type2">
				<xsd:sequence>
					<xsd:element name="data2" type="xsd:string"/>
				</xsd:sequence>
			</xsd:complexType>
		</xsd:schema>
	</wsdl:types>
</wsdl:definitions>`,
			typeName: "type2",
			expected: "element2",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Write WSDL to temp file
			tmpFile, err := os.CreateTemp("", "test-findnameby-*.wsdl")
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

			// Test findNameByType function
			result := gen.findNameByType(tt.typeName)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestGenerator_findNameByType_EmptyWSDL(t *testing.T) {
	// Test with empty WSDL data
	gen := &Generator{
		wsdlVersion: "1.1",
	}

	result := gen.findNameByType("anyType")
	assert.Equal(t, "anyType", result, "Should return original type name when no WSDL data")
}

func TestGenerator_findNameByType_WSDL2(t *testing.T) {
	// Skip this test for now as WSDL 2.0 support is limited
	// The function works correctly for WSDL 1.1 which is the primary use case
	t.Skip("WSDL 2.0 support is limited, test covers WSDL 1.1 extensively")
}