package generator

import (
	"context"
	"os"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestXMLNameGeneration_MessageElements(t *testing.T) {
	// Test that XMLName is properly generated for message elements
	wsdlContent := `<?xml version="1.0" encoding="UTF-8"?>
<wsdl:definitions xmlns:wsdl="http://schemas.xmlsoap.org/wsdl/" 
	targetNamespace="http://b2b.also.com/ActWebServices.PriceAvailability" 
	xmlns:tns="http://b2b.also.com/ActWebServices.PriceAvailability" 
	xmlns:xsd="http://www.w3.org/2001/XMLSchema"
	xmlns:soap="http://schemas.xmlsoap.org/wsdl/soap/">
	
	<wsdl:types>
		<xsd:schema targetNamespace="http://b2b.also.com/ActWebServices.PriceAvailability">
			<xsd:element name="checkV1" type="tns:checkV1"/>
			<xsd:element name="checkV1Response" type="tns:checkV1Response"/>
			<xsd:element name="nonMessageElement" type="tns:nonMessageType"/>
			
			<xsd:complexType name="checkV1">
				<xsd:sequence>
					<xsd:element name="PartnerCountryCode" type="xsd:string"/>
					<xsd:element name="StockType" type="xsd:string"/>
				</xsd:sequence>
			</xsd:complexType>
			
			<xsd:complexType name="checkV1Response">
				<xsd:sequence>
					<xsd:element name="ErrorMessage" nillable="true" type="xsd:string"/>
				</xsd:sequence>
			</xsd:complexType>
			
			<xsd:complexType name="nonMessageType">
				<xsd:sequence>
					<xsd:element name="data" type="xsd:string"/>
				</xsd:sequence>
			</xsd:complexType>
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
			<soap:operation soapAction="test_checkV1"/>
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
	tmpFile, err := os.CreateTemp("", "test-xmlname-*.wsdl")
	require.NoError(t, err)
	defer func() {
		_ = os.Remove(tmpFile.Name())
	}()

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

	// Test that message elements have XMLName with proper namespaces
	
	// Request element should have tns: prefix
	assert.Contains(t, typesStr, "type CheckV1 struct {")
	assert.Contains(t, typesStr, `XMLName xml.Name `+"`"+`xml:"tns:checkV1"`+"`")
	
	// Response element should use local name only
	assert.Contains(t, typesStr, "type CheckV1Response struct {")
	assert.Contains(t, typesStr, `XMLName xml.Name `+"`"+`xml:"checkV1Response"`+"`")
	
	// Non-message element should NOT have XMLName
	assert.Contains(t, typesStr, "type NonMessageType struct {")
	
	// Count XMLName occurrences to ensure only message elements have them
	checkV1XMLNameCount := strings.Count(typesStr, `XMLName xml.Name `+"`"+`xml:"tns:checkV1"`+"`")
	checkV1ResponseXMLNameCount := strings.Count(typesStr, `XMLName xml.Name `+"`"+`xml:"checkV1Response"`+"`")
	
	assert.Equal(t, 1, checkV1XMLNameCount, "CheckV1 should have exactly one XMLName")
	assert.Equal(t, 1, checkV1ResponseXMLNameCount, "CheckV1Response should have exactly one XMLName")
	
	// NonMessageType should not have XMLName
	nonMessageTypeStart := strings.Index(typesStr, "type NonMessageType struct {")
	require.NotEqual(t, -1, nonMessageTypeStart, "NonMessageType should be defined")
	
	// Find the end of the NonMessageType struct
	nonMessageTypeEnd := strings.Index(typesStr[nonMessageTypeStart:], "}")
	require.NotEqual(t, -1, nonMessageTypeEnd, "NonMessageType struct should be closed")
	
	nonMessageTypeStr := typesStr[nonMessageTypeStart : nonMessageTypeStart+nonMessageTypeEnd+1]
	assert.NotContains(t, nonMessageTypeStr, "XMLName", "NonMessageType should not have XMLName")
}

func TestXMLNameGeneration_ElementFormQualified(t *testing.T) {
	// Test XMLName generation with elementFormDefault="qualified"
	wsdlContent := `<?xml version="1.0" encoding="UTF-8"?>
<wsdl:definitions xmlns:wsdl="http://schemas.xmlsoap.org/wsdl/" 
	targetNamespace="http://example.com/qualified" 
	xmlns:tns="http://example.com/qualified" 
	xmlns:xsd="http://www.w3.org/2001/XMLSchema"
	xmlns:soap="http://schemas.xmlsoap.org/wsdl/soap/">
	
	<wsdl:types>
		<xsd:schema targetNamespace="http://example.com/qualified" elementFormDefault="qualified">
			<xsd:element name="qualifiedRequest" type="tns:qualifiedRequestType"/>
			
			<xsd:complexType name="qualifiedRequestType">
				<xsd:sequence>
					<xsd:element name="param" type="xsd:string"/>
				</xsd:sequence>
			</xsd:complexType>
		</xsd:schema>
	</wsdl:types>
	
	<wsdl:message name="qualifiedMessage">
		<wsdl:part name="parameters" element="tns:qualifiedRequest"/>
	</wsdl:message>
	
	<wsdl:portType name="QualifiedPortType">
		<wsdl:operation name="qualifiedOp">
			<wsdl:input message="tns:qualifiedMessage"/>
		</wsdl:operation>
	</wsdl:portType>
</wsdl:definitions>`

	// Write to temp file
	tmpFile, err := os.CreateTemp("", "test-qualified-*.wsdl")
	require.NoError(t, err)
	defer func() {
		_ = os.Remove(tmpFile.Name())
	}()

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

	// With elementFormDefault="qualified", XMLName should use full namespace
	assert.Contains(t, typesStr, "type QualifiedRequestType struct {")
	assert.Contains(t, typesStr, `XMLName xml.Name `+"`"+`xml:"http://example.com/qualified qualifiedRequest"`+"`")
}

func TestXMLNameGeneration_NoMessages(t *testing.T) {
	// Test that complex types without messages don't get XMLName
	wsdlContent := `<?xml version="1.0" encoding="UTF-8"?>
<wsdl:definitions xmlns:wsdl="http://schemas.xmlsoap.org/wsdl/" 
	targetNamespace="http://example.com/test" 
	xmlns:tns="http://example.com/test" 
	xmlns:xsd="http://www.w3.org/2001/XMLSchema">
	
	<wsdl:types>
		<xsd:schema targetNamespace="http://example.com/test">
			<xsd:element name="standaloneElement" type="tns:standaloneType"/>
			
			<xsd:complexType name="standaloneType">
				<xsd:sequence>
					<xsd:element name="data" type="xsd:string"/>
				</xsd:sequence>
			</xsd:complexType>
		</xsd:schema>
	</wsdl:types>
</wsdl:definitions>`

	// Write to temp file
	tmpFile, err := os.CreateTemp("", "test-nomsg-*.wsdl")
	require.NoError(t, err)
	defer func() {
		_ = os.Remove(tmpFile.Name())
	}()

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

	// StandaloneType should not have XMLName since it's not used in messages
	assert.Contains(t, typesStr, "type StandaloneType struct {")
	
	// Find the StandaloneType struct and verify it doesn't have XMLName
	standaloneTypeStart := strings.Index(typesStr, "type StandaloneType struct {")
	require.NotEqual(t, -1, standaloneTypeStart, "StandaloneType should be defined")
	
	standaloneTypeEnd := strings.Index(typesStr[standaloneTypeStart:], "}")
	require.NotEqual(t, -1, standaloneTypeEnd, "StandaloneType struct should be closed")
	
	standaloneTypeStr := typesStr[standaloneTypeStart : standaloneTypeStart+standaloneTypeEnd+1]
	assert.NotContains(t, standaloneTypeStr, "XMLName", "StandaloneType should not have XMLName when not used in messages")
}