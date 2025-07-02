package generator

import (
	"context"
	"os"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestElementAliasGeneration(t *testing.T) {
	// Create a WSDL that has elements with the same name as complex types
	wsdlContent := `<?xml version="1.0" encoding="UTF-8"?>
<wsdl:definitions xmlns:wsdl="http://schemas.xmlsoap.org/wsdl/" 
	targetNamespace="http://example.com/test" 
	xmlns:tns="http://example.com/test" 
	xmlns:xsd="http://www.w3.org/2001/XMLSchema"
	xmlns:soap="http://schemas.xmlsoap.org/wsdl/soap/">
	
	<wsdl:types>
		<xsd:schema targetNamespace="http://example.com/test">
			<!-- Define elements that reference complex types with the same name -->
			<xsd:element name="checkV1" type="tns:checkV1"/>
			<xsd:element name="checkV1Response" type="tns:checkV1Response"/>
			
			<!-- Define the complex types -->
			<xsd:complexType name="checkV1">
				<xsd:sequence>
					<xsd:element name="input" type="xsd:string"/>
				</xsd:sequence>
			</xsd:complexType>
			
			<xsd:complexType name="checkV1Response">
				<xsd:sequence>
					<xsd:element name="output" type="xsd:string"/>
				</xsd:sequence>
			</xsd:complexType>
			
			<!-- Also test an element with a different name than its type -->
			<xsd:element name="differentName" type="tns:checkV1"/>
		</xsd:schema>
	</wsdl:types>
	
	<wsdl:message name="checkV1Request">
		<wsdl:part name="parameters" element="tns:checkV1"/>
	</wsdl:message>
	<wsdl:message name="checkV1Response">
		<wsdl:part name="parameters" element="tns:checkV1Response"/>
	</wsdl:message>
	
	<wsdl:portType name="TestPortType">
		<wsdl:operation name="checkV1">
			<wsdl:input message="tns:checkV1Request"/>
			<wsdl:output message="tns:checkV1Response"/>
		</wsdl:operation>
	</wsdl:portType>
	
	<wsdl:binding name="TestBinding" type="tns:TestPortType">
		<soap:binding style="document" transport="http://schemas.xmlsoap.org/soap/http"/>
		<wsdl:operation name="checkV1">
			<soap:operation soapAction=""/>
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

	// Write to temp file
	tmpFile, err := os.CreateTemp("", "test-element-alias-*.wsdl")
	require.NoError(t, err)
	defer os.Remove(tmpFile.Name())

	_, err = tmpFile.WriteString(wsdlContent)
	require.NoError(t, err)
	err = tmpFile.Close()
	require.NoError(t, err)

	// Generate code
	gen, err := New(tmpFile.Name(), WithPackage("testpkg"))
	require.NoError(t, err)

	result, err := gen.Generate(context.Background())
	require.NoError(t, err)

	// Check types file
	typesCode, ok := result["types"]
	require.True(t, ok, "types file should be generated")

	typesStr := string(typesCode)

	// Should have struct definitions for CheckV1 and CheckV1Response
	assert.Contains(t, typesStr, "type CheckV1 struct {")
	assert.Contains(t, typesStr, "type CheckV1Response struct {")
	
	// Should NOT have duplicate type aliases like "type CheckV1 checkV1"
	assert.NotContains(t, typesStr, "type CheckV1 checkV1")
	assert.NotContains(t, typesStr, "type CheckV1Response checkV1Response")
	
	// Should have type alias for the element with different name
	assert.Contains(t, typesStr, "type DifferentName CheckV1")
	
	// Count occurrences of type definitions
	checkV1Count := strings.Count(typesStr, "type CheckV1 ")
	checkV1ResponseCount := strings.Count(typesStr, "type CheckV1Response ")
	
	// Each type should be defined exactly once
	assert.Equal(t, 1, checkV1Count, "CheckV1 should be defined exactly once")
	assert.Equal(t, 1, checkV1ResponseCount, "CheckV1Response should be defined exactly once")
}

func TestElementAliasWithPrimitiveTypes(t *testing.T) {
	// Test that elements with primitive types still get type aliases
	wsdlContent := `<?xml version="1.0" encoding="UTF-8"?>
<wsdl:definitions xmlns:wsdl="http://schemas.xmlsoap.org/wsdl/" 
	targetNamespace="http://example.com/test" 
	xmlns:tns="http://example.com/test" 
	xmlns:xsd="http://www.w3.org/2001/XMLSchema"
	xmlns:soap="http://schemas.xmlsoap.org/wsdl/soap/">
	
	<wsdl:types>
		<xsd:schema targetNamespace="http://example.com/test">
			<!-- Elements with primitive types should get type aliases -->
			<xsd:element name="myString" type="xsd:string"/>
			<xsd:element name="myInt" type="xsd:int"/>
			<xsd:element name="myDateTime" type="xsd:dateTime"/>
		</xsd:schema>
	</wsdl:types>
	
	<wsdl:service name="TestService">
		<wsdl:port name="TestPort">
			<soap:address location="http://example.com/test"/>
		</wsdl:port>
	</wsdl:service>
</wsdl:definitions>`

	// Write to temp file
	tmpFile, err := os.CreateTemp("", "test-primitive-alias-*.wsdl")
	require.NoError(t, err)
	defer os.Remove(tmpFile.Name())

	_, err = tmpFile.WriteString(wsdlContent)
	require.NoError(t, err)
	err = tmpFile.Close()
	require.NoError(t, err)

	// Generate code
	gen, err := New(tmpFile.Name(), WithPackage("testpkg"))
	require.NoError(t, err)

	result, err := gen.Generate(context.Background())
	require.NoError(t, err)

	// Check types file
	typesCode, ok := result["types"]
	require.True(t, ok, "types file should be generated")

	typesStr := string(typesCode)

	// Should have type aliases for primitive type elements
	assert.Contains(t, typesStr, "type MyString string")
	assert.Contains(t, typesStr, "type MyInt int32")
	assert.Contains(t, typesStr, "type MyDateTime soap.XSDDateTime")
	
	// Should have marshal/unmarshal functions for XSDDateTime
	assert.Contains(t, typesStr, "func (xdt MyDateTime) MarshalXML")
	assert.Contains(t, typesStr, "func (xdt *MyDateTime) UnmarshalXML")
}