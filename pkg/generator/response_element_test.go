// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at http://mozilla.org/MPL/2.0/.

package generator

import (
	"context"
	"encoding/xml"
	"os"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestResponseElementNamespaceHandling(t *testing.T) {
	// Test WSDL with properly defined operations and messages
	wsdl := `<?xml version="1.0" encoding="UTF-8"?>
<definitions xmlns="http://schemas.xmlsoap.org/wsdl/"
             xmlns:soap="http://schemas.xmlsoap.org/wsdl/soap/"
             xmlns:tns="https://example.com/service"
             xmlns:xsd="http://www.w3.org/2001/XMLSchema"
             targetNamespace="https://example.com/service">
    <types>
        <schema xmlns="http://www.w3.org/2001/XMLSchema"
                targetNamespace="https://example.com/service">
            
            <element name="SendungsErstellung">
                <complexType>
                    <sequence>
                        <element name="Versender" type="string"/>
                        <element name="Kundenreferenz" type="string"/>
                    </sequence>
                </complexType>
            </element>
            
            <element name="SendungsRueckmeldung">
                <complexType>
                    <sequence>
                        <element name="SendungsnummerAX4" type="string"/>
                        <element name="Frachtbriefnummer" type="string"/>
                    </sequence>
                </complexType>
            </element>
            
            <element name="GetOrder">
                <complexType>
                    <sequence>
                        <element name="OrderID" type="string"/>
                    </sequence>
                </complexType>
            </element>
            
            <element name="GetOrderResponse">
                <complexType>
                    <sequence>
                        <element name="OrderID" type="string"/>
                        <element name="Status" type="string"/>
                    </sequence>
                </complexType>
            </element>
        </schema>
    </types>
    
    <message name="SendungsErstellungRequest">
        <part name="parameters" element="tns:SendungsErstellung"/>
    </message>
    
    <message name="SendungsRueckmeldungResponse">
        <part name="parameters" element="tns:SendungsRueckmeldung"/>
    </message>
    
    <message name="GetOrderRequest">
        <part name="parameters" element="tns:GetOrder"/>
    </message>
    
    <message name="GetOrderResponse">
        <part name="parameters" element="tns:GetOrderResponse"/>
    </message>
    
    <portType name="ServicePortType">
        <operation name="SendungsErstellung">
            <input message="tns:SendungsErstellungRequest"/>
            <output message="tns:SendungsRueckmeldungResponse"/>
        </operation>
        <operation name="GetOrder">
            <input message="tns:GetOrderRequest"/>
            <output message="tns:GetOrderResponse"/>
        </operation>
    </portType>
    
    <binding name="ServiceBinding" type="tns:ServicePortType">
        <soap:binding transport="http://schemas.xmlsoap.org/soap/http"/>
        <operation name="SendungsErstellung">
            <soap:operation soapAction="SendungsErstellung"/>
            <input><soap:body use="literal"/></input>
            <output><soap:body use="literal"/></output>
        </operation>
        <operation name="GetOrder">
            <soap:operation soapAction="GetOrder"/>
            <input><soap:body use="literal"/></input>
            <output><soap:body use="literal"/></output>
        </operation>
    </binding>
    
    <service name="Service">
        <port name="ServicePort" binding="tns:ServiceBinding">
            <soap:address location="http://example.com/service"/>
        </port>
    </service>
</definitions>`

	file, err := os.CreateTemp("", "test-response-element-*.wsdl")
	require.NoError(t, err)
	defer func() { _ = os.Remove(file.Name()) }()
	defer func() { _ = file.Close() }()
	
	_, err = file.WriteString(wsdl)
	require.NoError(t, err)

	g, err := New(file.Name(), WithPackage("service"))
	require.NoError(t, err)

	files, err := g.Generate(context.Background())
	require.NoError(t, err)

	types, ok := files["types"]
	require.True(t, ok, "types file should be generated")
	
	typesStr := string(types)
	t.Logf("Generated types:\n%s", typesStr)
	
	// Response elements should use local name only (no namespace prefix)
	assert.True(t, strings.Contains(typesStr, `XMLName xml.Name `+"`"+`xml:"SendungsRueckmeldung"`+"`"), 
		"SendungsRueckmeldung should use local name only")
	assert.True(t, strings.Contains(typesStr, `XMLName xml.Name `+"`"+`xml:"GetOrderResponse"`+"`"), 
		"GetOrderResponse should use local name only")
	
	// Request elements should still use namespace prefix
	assert.True(t, strings.Contains(typesStr, `XMLName xml.Name `+"`"+`xml:"tns:SendungsErstellung"`+"`"), 
		"SendungsErstellung should use tns: prefix")
	assert.True(t, strings.Contains(typesStr, `XMLName xml.Name `+"`"+`xml:"tns:GetOrder"`+"`"), 
		"GetOrder should use tns: prefix")
}

// TestResponseUnmarshalWithDifferentPrefixes verifies that response types can be unmarshaled
// regardless of the namespace prefix used by the server
func TestResponseUnmarshalWithDifferentPrefixes(t *testing.T) {
	// Define a response type with local name only
	type GetOrderResponse struct {
		XMLName xml.Name `xml:"GetOrderResponse"`
		OrderID string   `xml:"OrderID,omitempty"`
		Status  string   `xml:"Status,omitempty"`
	}
	
	// Test with different namespace prefixes
	testCases := []struct {
		name      string
		xmlData   string
		wantError bool
	}{
		{
			name: "ns1 prefix",
			xmlData: `<ns1:GetOrderResponse xmlns:ns1="https://example.com/service">
				<OrderID>12345</OrderID>
				<Status>Delivered</Status>
			</ns1:GetOrderResponse>`,
			wantError: false,
		},
		{
			name: "ns2 prefix",
			xmlData: `<ns2:GetOrderResponse xmlns:ns2="https://example.com/service">
				<OrderID>12345</OrderID>
				<Status>Delivered</Status>
			</ns2:GetOrderResponse>`,
			wantError: false,
		},
		{
			name: "tns prefix",
			xmlData: `<tns:GetOrderResponse xmlns:tns="https://example.com/service">
				<OrderID>12345</OrderID>
				<Status>Delivered</Status>
			</tns:GetOrderResponse>`,
			wantError: false,
		},
		{
			name: "no prefix",
			xmlData: `<GetOrderResponse xmlns="https://example.com/service">
				<OrderID>12345</OrderID>
				<Status>Delivered</Status>
			</GetOrderResponse>`,
			wantError: false,
		},
	}
	
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			var resp GetOrderResponse
			err := xml.Unmarshal([]byte(tc.xmlData), &resp)
			
			if tc.wantError {
				assert.Error(t, err)
			} else {
				require.NoError(t, err)
				assert.Equal(t, "12345", resp.OrderID)
				assert.Equal(t, "Delivered", resp.Status)
			}
		})
	}
}

// TestRequestTypesRetainNamespacePrefix verifies that request types still use namespace prefix
func TestRequestTypesRetainNamespacePrefix(t *testing.T) {
	// Define a request type with tns: prefix (as currently generated)
	type GetOrder struct {
		XMLName xml.Name `xml:"tns:GetOrder"`
		OrderID string   `xml:"OrderID,omitempty"`
	}
	
	// This type is used for marshaling requests, not unmarshaling responses
	// So the tns: prefix is correct and should be retained
	order := GetOrder{OrderID: "12345"}
	
	data, err := xml.Marshal(order)
	require.NoError(t, err)
	
	xmlStr := string(data)
	// Should produce <tns:GetOrder><OrderID>12345</OrderID></tns:GetOrder>
	assert.Contains(t, xmlStr, "<tns:GetOrder>")
	assert.Contains(t, xmlStr, "<OrderID>12345</OrderID>")
	assert.Contains(t, xmlStr, "</tns:GetOrder>")
}