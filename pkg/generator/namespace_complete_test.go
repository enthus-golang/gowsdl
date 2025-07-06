// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at http://mozilla.org/MPL/2.0/.

package generator

import (
	"bytes"
	"context"
	"encoding/xml"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestCompleteNamespaceHandling tests the complete flow with a WSDL similar to the user's
func TestCompleteNamespaceHandling(t *testing.T) {
	// WSDL similar to user's case - without elementFormDefault (defaults to unqualified)
	wsdl := `<?xml version="1.0" encoding="UTF-8"?>
<definitions xmlns="http://schemas.xmlsoap.org/wsdl/"
             xmlns:soap="http://schemas.xmlsoap.org/wsdl/soap/"
             xmlns:tns="https://wsdemo.ax4.com/ws/GeneralOvernight"
             xmlns:xsd="http://www.w3.org/2001/XMLSchema"
             targetNamespace="https://wsdemo.ax4.com/ws/GeneralOvernight">
    <types>
        <xsd:schema targetNamespace="https://wsdemo.ax4.com/ws/GeneralOvernight">
            <xsd:element name="GOWebService_SendungsErstellung">
                <xsd:complexType>
                    <xsd:sequence>
                        <xsd:element name="Versender" type="xsd:string" minOccurs="0"/>
                        <xsd:element name="Benutzername" type="xsd:string" minOccurs="0"/>
                        <xsd:element name="Status" type="xsd:string" minOccurs="0"/>
                        <xsd:element name="Kundenreferenz" type="xsd:string" minOccurs="0"/>
                        <xsd:element name="Ansprechpartner" minOccurs="0" maxOccurs="1">
                            <xsd:complexType>
                                <xsd:sequence>
                                    <xsd:element name="Telefon" minOccurs="1" maxOccurs="1">
                                        <xsd:complexType>
                                            <xsd:sequence>
                                                <xsd:element name="LaenderPrefix" type="xsd:string" minOccurs="1" maxOccurs="1"/>
                                                <xsd:element name="Vorwahl" type="xsd:string" minOccurs="1" maxOccurs="1"/>
                                                <xsd:element name="Telefonnummer" type="xsd:string" minOccurs="1" maxOccurs="1"/>
                                            </xsd:sequence>
                                        </xsd:complexType>
                                    </xsd:element>
                                </xsd:sequence>
                            </xsd:complexType>
                        </xsd:element>
                    </xsd:sequence>
                </xsd:complexType>
            </xsd:element>
            
            <xsd:element name="GOWebService_SendungsErstellungResponse">
                <xsd:complexType>
                    <xsd:sequence>
                        <xsd:element name="status" type="xsd:string"/>
                    </xsd:sequence>
                </xsd:complexType>
            </xsd:element>
        </xsd:schema>
    </types>
    
    <message name="SendungsErstellungRequest">
        <part name="parameters" element="tns:GOWebService_SendungsErstellung"/>
    </message>
    
    <message name="SendungsErstellungResponse">
        <part name="parameters" element="tns:GOWebService_SendungsErstellungResponse"/>
    </message>
    
    <portType name="GeneralOvernightService">
        <operation name="SendungsErstellung">
            <input message="tns:SendungsErstellungRequest"/>
            <output message="tns:SendungsErstellungResponse"/>
        </operation>
    </portType>
    
    <binding name="GeneralOvernightServiceBinding" type="tns:GeneralOvernightService">
        <soap:binding transport="http://schemas.xmlsoap.org/soap/http"/>
        <operation name="SendungsErstellung">
            <soap:operation soapAction="SendungsErstellung"/>
            <input>
                <soap:body use="literal"/>
            </input>
            <output>
                <soap:body use="literal"/>
            </output>
        </operation>
    </binding>
    
    <service name="GeneralOvernightService">
        <port name="GeneralOvernightServicePort" binding="tns:GeneralOvernightServiceBinding">
            <soap:address location="https://example.com/service"/>
        </port>
    </service>
</definitions>`

	file, err := os.CreateTemp("", "test-namespace-complete-*.wsdl")
	require.NoError(t, err)
	defer func() { _ = os.Remove(file.Name()) }()
	defer func() { _ = file.Close() }()
	
	_, err = file.WriteString(wsdl)
	require.NoError(t, err)

	// Generate code
	g, err := New(file.Name(), WithPackage("service"))
	require.NoError(t, err)

	files, err := g.Generate(context.Background())
	require.NoError(t, err)

	types, ok := files["types"]
	require.True(t, ok, "types file should be generated")
	
	typesStr := string(types)
	
	// Check that XMLName uses prefix
	assert.Contains(t, typesStr, `xml:"tns:GOWebService_SendungsErstellung"`)
	
	// Check that TargetNamespace method is generated
	assert.Contains(t, typesStr, "func (t GOWebService_SendungsErstellung) TargetNamespace() string")
	assert.Contains(t, typesStr, `return "https://wsdemo.ax4.com/ws/GeneralOvernight"`)
	
	// Test actual marshaling with our enhanced SOAP envelope
	testMarshalingWithEnhancedEnvelope(t)
}

func testMarshalingWithEnhancedEnvelope(t *testing.T) {
	// Enhanced body type
	type EnhancedBody struct {
		XMLName xml.Name    `xml:"soap:Body"`
		Content interface{} `xml:",any"`
	}
	
	// Enhanced envelope that declares namespaces
	type EnhancedEnvelope struct {
		XMLName  xml.Name `xml:"soap:Envelope"`
		XmlNS    string   `xml:"xmlns:soap,attr"`
		XmlNSTns string   `xml:"xmlns:tns,attr,omitempty"`
		XmlNSXSI string   `xml:"xmlns:xsi,attr,omitempty"`
		XmlNSXSD string   `xml:"xmlns:xsd,attr,omitempty"`
		Body     EnhancedBody
	}
	
	// Generated inline types
	type GOWebService_SendungsErstellungAnsprechpartnerTypeTelefonType struct {
		LaenderPrefix  string `xml:"LaenderPrefix,omitempty" json:"LaenderPrefix,omitempty"`
		Vorwahl        string `xml:"Vorwahl,omitempty" json:"Vorwahl,omitempty"`
		Telefonnummer  string `xml:"Telefonnummer,omitempty" json:"Telefonnummer,omitempty"`
	}
	
	type GOWebService_SendungsErstellungAnsprechpartnerType struct {
		Telefon GOWebService_SendungsErstellungAnsprechpartnerTypeTelefonType `xml:"Telefon,omitempty" json:"Telefon,omitempty"`
	}
	
	// Operation type with namespace prefix
	type GOWebService_SendungsErstellung struct {
		XMLName         xml.Name `xml:"tns:GOWebService_SendungsErstellung"`
		Versender       string   `xml:"Versender,omitempty"`
		Benutzername    string   `xml:"Benutzername,omitempty"`
		Status          string   `xml:"Status,omitempty"`
		Kundenreferenz  string   `xml:"Kundenreferenz,omitempty"`
		Ansprechpartner *GOWebService_SendungsErstellungAnsprechpartnerType `xml:"Ansprechpartner,omitempty"`
	}
	
	// Create request with nested structure
	request := &GOWebService_SendungsErstellung{
		Versender:      "2611570",
		Benutzername:   "21480MCL",
		Status:         "1",
		Kundenreferenz: "AU201061_LS300999",
		Ansprechpartner: &GOWebService_SendungsErstellungAnsprechpartnerType{
			Telefon: GOWebService_SendungsErstellungAnsprechpartnerTypeTelefonType{
				LaenderPrefix:  "49",
				Vorwahl:        "89",
				Telefonnummer:  "12345678",
			},
		},
	}
	
	envelope := EnhancedEnvelope{
		XmlNS:    "http://schemas.xmlsoap.org/soap/envelope/",
		XmlNSTns: "https://wsdemo.ax4.com/ws/GeneralOvernight",
		XmlNSXSI: "http://www.w3.org/2001/XMLSchema-instance",
		XmlNSXSD: "http://www.w3.org/2001/XMLSchema",
		Body: EnhancedBody{
			Content: request,
		},
	}
	
	var buf bytes.Buffer
	encoder := xml.NewEncoder(&buf)
	encoder.Indent("", "  ")
	
	err := encoder.Encode(envelope)
	require.NoError(t, err)
	
	xmlStr := buf.String()
	t.Logf("Generated SOAP envelope:\n%s", xmlStr)
	
	// Verify correct structure
	assert.Contains(t, xmlStr, `xmlns:tns="https://wsdemo.ax4.com/ws/GeneralOvernight"`)
	assert.Contains(t, xmlStr, `<tns:GOWebService_SendungsErstellung>`)
	assert.Contains(t, xmlStr, `<Versender>2611570</Versender>`)
	assert.Contains(t, xmlStr, `<Ansprechpartner>`)
	assert.Contains(t, xmlStr, `<Telefon>`)
	assert.Contains(t, xmlStr, `<LaenderPrefix>49</LaenderPrefix>`)
	
	// Verify unqualified child elements
	assert.NotContains(t, xmlStr, `<tns:Versender>`)
	assert.NotContains(t, xmlStr, `<tns:Ansprechpartner>`)
	assert.NotContains(t, xmlStr, `<tns:Telefon>`)
	assert.NotContains(t, xmlStr, `<tns:LaenderPrefix>`)
	assert.NotContains(t, xmlStr, `xmlns=`)  // No default namespace declarations
}