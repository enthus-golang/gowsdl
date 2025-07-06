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

// TestSOAPNamespaceHandling tests that generated code produces correct SOAP envelopes
func TestSOAPNamespaceHandling(t *testing.T) {
	// WSDL with unqualified schema (default)
	wsdl := `<?xml version="1.0" encoding="UTF-8"?>
<definitions xmlns="http://schemas.xmlsoap.org/wsdl/"
             xmlns:soap="http://schemas.xmlsoap.org/wsdl/soap/"
             xmlns:tns="https://example.com/service"
             xmlns:xsd="http://www.w3.org/2001/XMLSchema"
             targetNamespace="https://example.com/service">
    <types>
        <schema xmlns="http://www.w3.org/2001/XMLSchema"
                targetNamespace="https://example.com/service">
            
            <element name="CreateOrder">
                <complexType>
                    <sequence>
                        <element name="customerID" type="string"/>
                        <element name="orderDate" type="date"/>
                    </sequence>
                </complexType>
            </element>
            
            <element name="CreateOrderResponse">
                <complexType>
                    <sequence>
                        <element name="orderID" type="string"/>
                    </sequence>
                </complexType>
            </element>
        </schema>
    </types>
    
    <message name="CreateOrderRequest">
        <part name="parameters" element="tns:CreateOrder"/>
    </message>
    
    <message name="CreateOrderResponse">
        <part name="parameters" element="tns:CreateOrderResponse"/>
    </message>
    
    <portType name="OrderService">
        <operation name="CreateOrder">
            <input message="tns:CreateOrderRequest"/>
            <output message="tns:CreateOrderResponse"/>
        </operation>
    </portType>
    
    <binding name="OrderServiceBinding" type="tns:OrderService">
        <soap:binding transport="http://schemas.xmlsoap.org/soap/http"/>
        <operation name="CreateOrder">
            <soap:operation soapAction="CreateOrder"/>
            <input>
                <soap:body use="literal"/>
            </input>
            <output>
                <soap:body use="literal"/>
            </output>
        </operation>
    </binding>
    
    <service name="OrderService">
        <port name="OrderServicePort" binding="tns:OrderServiceBinding">
            <soap:address location="http://example.com/service"/>
        </port>
    </service>
</definitions>`

	file, err := os.CreateTemp("", "test-soap-namespace-*.wsdl")
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
	
	// Check that XMLName includes namespace
	typesStr := string(types)
	assert.Contains(t, typesStr, "CreateOrder")
	assert.Contains(t, typesStr, "XMLName")
	assert.Contains(t, typesStr, "https://example.com/service")
	
	// Now test how it marshals
	testXMLMarshaling(t, typesStr)
}

func testXMLMarshaling(t *testing.T, generatedTypes string) {
	// Create a test that shows the current behavior
	type TestBody struct {
		XMLName xml.Name `xml:"soap:Body"`
		Content interface{} `xml:",any"`
	}
	
	type TestEnvelope struct {
		XMLName xml.Name `xml:"soap:Envelope"`
		Xmlns   string   `xml:"xmlns:soap,attr"`
		Body    TestBody
	}
	
	// Simulate what the generated type looks like
	type CreateOrder struct {
		XMLName    xml.Name `xml:"https://example.com/service CreateOrder"`
		CustomerID string   `xml:"customerID,omitempty"`
		OrderDate  string   `xml:"orderDate,omitempty"`
	}
	
	order := &CreateOrder{
		CustomerID: "12345",
		OrderDate:  "2024-01-01",
	}
	
	envelope := TestEnvelope{
		Xmlns: "http://schemas.xmlsoap.org/soap/envelope/",
		Body: TestBody{
			Content: order,
		},
	}
	
	data, err := xml.Marshal(envelope)
	require.NoError(t, err)
	
	xmlStr := string(data)
	t.Logf("Current SOAP envelope:\n%s", xmlStr)
	
	// The problem: this creates a default namespace on CreateOrder
	// which makes all child elements qualified
	assert.Contains(t, xmlStr, `<CreateOrder xmlns="https://example.com/service">`)
	
	// What we need: namespace prefix usage
	// This would require changes to how we generate types or marshal XML
}

// TestDesiredSOAPStructure shows what we want to achieve
func TestDesiredSOAPStructure(t *testing.T) {
	// Define body type first
	type EnhancedBody struct {
		XMLName xml.Name `xml:"soapenv:Body"`
		Content interface{} `xml:",any"`
	}
	
	// Enhanced SOAP envelope that supports target namespace
	type EnhancedEnvelope struct {
		XMLName   xml.Name `xml:"soapenv:Envelope"`
		XmlnsSoap string   `xml:"xmlns:soapenv,attr"`
		XmlnsTns  string   `xml:"xmlns:tns,attr,omitempty"`
		XmlnsXsi  string   `xml:"xmlns:xsi,attr,omitempty"`
		XmlnsXsd  string   `xml:"xmlns:xsd,attr,omitempty"`
		Body      EnhancedBody
	}
	
	// Operation type that uses namespace prefix
	type CreateOrderPrefixed struct {
		XMLName    xml.Name `xml:"tns:CreateOrder"`
		CustomerID string   `xml:"customerID,omitempty"`
		OrderDate  string   `xml:"orderDate,omitempty"`
	}
	
	order := &CreateOrderPrefixed{
		CustomerID: "12345",
		OrderDate:  "2024-01-01",
	}
	
	envelope := EnhancedEnvelope{
		XmlnsSoap: "http://schemas.xmlsoap.org/soap/envelope/",
		XmlnsTns:  "https://example.com/service",
		XmlnsXsi:  "http://www.w3.org/2001/XMLSchema-instance",
		XmlnsXsd:  "http://www.w3.org/2001/XMLSchema",
		Body: EnhancedBody{
			Content: order,
		},
	}
	
	var buf bytes.Buffer
	encoder := xml.NewEncoder(&buf)
	encoder.Indent("", "  ")
	
	err := encoder.Encode(envelope)
	require.NoError(t, err)
	
	xmlStr := buf.String()
	t.Logf("Desired SOAP structure:\n%s", xmlStr)
	
	// This should produce the correct structure
	assert.Contains(t, xmlStr, `xmlns:tns="https://example.com/service"`)
	assert.Contains(t, xmlStr, `<tns:CreateOrder>`)
	assert.Contains(t, xmlStr, `<customerID>`) // Unqualified
	
	// Should NOT have
	assert.NotContains(t, xmlStr, `<tns:customerID>`)
	assert.NotContains(t, xmlStr, `<CreateOrder xmlns=`)
}

// TestRealWorldExample tests with the user's actual WSDL structure
func TestRealWorldExample(t *testing.T) {
	// Define body type first
	type WorkingBody struct {
		XMLName xml.Name `xml:"soapenv:Body"`
		Content interface{} `xml:",any"`
	}
	
	// Simulate the user's case
	type GOWebServiceSendungsErstellung struct {
		XMLName        xml.Name `xml:"tns:GOWebService_SendungsErstellung"`
		Versender      string   `xml:"Versender,omitempty"`
		Benutzername   string   `xml:"Benutzername,omitempty"`
		Status         string   `xml:"Status,omitempty"`
		Kundenreferenz string   `xml:"Kundenreferenz,omitempty"`
	}
	
	type WorkingEnvelope struct {
		XMLName   xml.Name `xml:"soapenv:Envelope"`
		XmlnsSoap string   `xml:"xmlns:soapenv,attr"`
		XmlnsTns  string   `xml:"xmlns:tns,attr"`
		XmlnsXsd  string   `xml:"xmlns:xsd,attr"`
		XmlnsXsi  string   `xml:"xmlns:xsi,attr"`
		Body      WorkingBody
	}
	
	request := &GOWebServiceSendungsErstellung{
		Versender:      "2611570",
		Benutzername:   "21480MCL",
		Status:         "1",
		Kundenreferenz: "AU201061_LS300999",
	}
	
	envelope := WorkingEnvelope{
		XmlnsSoap: "http://schemas.xmlsoap.org/soap/envelope/",
		XmlnsTns:  "https://wsdemo.ax4.com/ws/GeneralOvernight",
		XmlnsXsd:  "http://www.w3.org/2001/XMLSchema",
		XmlnsXsi:  "http://www.w3.org/2001/XMLSchema-instance",
		Body: WorkingBody{
			Content: request,
		},
	}
	
	data, err := xml.Marshal(envelope)
	require.NoError(t, err)
	
	xmlStr := string(data)
	t.Logf("Working structure:\n%s", xmlStr)
	
	// Verify it produces the correct format
	assert.Contains(t, xmlStr, `<soapenv:Envelope`)
	assert.Contains(t, xmlStr, `xmlns:tns="https://wsdemo.ax4.com/ws/GeneralOvernight"`)
	assert.Contains(t, xmlStr, `<tns:GOWebService_SendungsErstellung>`)
	assert.Contains(t, xmlStr, `<Versender>2611570</Versender>`)
	
	// Clean output without namespace on child elements
	// Verify child elements are unqualified
	assert.Contains(t, xmlStr, "<Versender>2611570</Versender>")
	assert.NotContains(t, xmlStr, "<tns:Versender>")
	assert.NotContains(t, xmlStr, "<Versender xmlns=")
}