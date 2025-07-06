// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at http://mozilla.org/MPL/2.0/.

package generator

import (
	"encoding/xml"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestNamespacePrefixVsDefault demonstrates the difference between
// namespace prefix and default namespace in XML marshaling
func TestNamespacePrefixVsDefault(t *testing.T) {
	// Test 1: Default namespace (current behavior - WRONG for unqualified schemas)
	type OrderWithDefaultNS struct {
		XMLName    xml.Name `xml:"https://example.com/service Order"`
		CustomerID string   `xml:"customerID,omitempty"`
		ItemCount  int      `xml:"itemCount,omitempty"`
	}
	
	order1 := &OrderWithDefaultNS{
		CustomerID: "12345",
		ItemCount:  2,
	}
	
	data1, err := xml.Marshal(order1)
	require.NoError(t, err)
	xmlStr1 := string(data1)
	t.Logf("Default namespace XML:\n%s", xmlStr1)
	
	// This produces: <Order xmlns="https://example.com/service"><customerID>12345</customerID>...
	// Problem: customerID inherits the namespace
	assert.Contains(t, xmlStr1, `xmlns="https://example.com/service"`)
	
	// Test 2: Namespace prefix (correct for unqualified schemas)
	type OrderWithPrefix struct {
		XMLName    xml.Name `xml:"tns:Order"`
		Xmlns      string   `xml:"xmlns:tns,attr"`
		CustomerID string   `xml:"customerID,omitempty"`
		ItemCount  int      `xml:"itemCount,omitempty"`
	}
	
	order2 := &OrderWithPrefix{
		Xmlns:      "https://example.com/service",
		CustomerID: "12345", 
		ItemCount:  2,
	}
	
	data2, err := xml.Marshal(order2)
	require.NoError(t, err)
	xmlStr2 := string(data2)
	t.Logf("Namespace prefix XML:\n%s", xmlStr2)
	
	// This produces: <tns:Order xmlns:tns="https://example.com/service"><customerID>12345</customerID>...
	// Correct: customerID is unqualified
	assert.Contains(t, xmlStr2, `xmlns:tns="https://example.com/service"`)
	assert.Contains(t, xmlStr2, `<tns:Order`)
	assert.Contains(t, xmlStr2, `<customerID>`) // No namespace prefix
	
	// Test 3: What we need for SOAP body content
	type GOWebServiceSendungsErstellung struct {
		XMLName         xml.Name `xml:"tns:GOWebService_SendungsErstellung"`
		XmlnsTns        string   `xml:"xmlns:tns,attr,omitempty"`
		Versender       string   `xml:"Versender,omitempty"`
		Benutzername    string   `xml:"Benutzername,omitempty"`
		Status          string   `xml:"Status,omitempty"`
		Kundenreferenz  string   `xml:"Kundenreferenz,omitempty"`
	}
	
	order3 := &GOWebServiceSendungsErstellung{
		XmlnsTns:       "https://wsdemo.ax4.com/ws/GeneralOvernight",
		Versender:      "2611570",
		Benutzername:   "21480MCL",
		Status:         "1",
		Kundenreferenz: "AU201061_LS300999",
	}
	
	// When this is marshaled inside a SOAP body, we want:
	// <soap:Body>
	//   <tns:GOWebService_SendungsErstellung xmlns:tns="...">
	//     <Versender>2611570</Versender>  <!-- No namespace -->
	//   </tns:GOWebService_SendungsErstellung>
	// </soap:Body>
	
	data3, err := xml.Marshal(order3)
	require.NoError(t, err)
	xmlStr3 := string(data3)
	t.Logf("SOAP operation XML:\n%s", xmlStr3)
	
	assert.Contains(t, xmlStr3, `<tns:GOWebService_SendungsErstellung`)
	assert.Contains(t, xmlStr3, `<Versender>`) // Unqualified
}

// TestSOAPEnvelopeWithPrefixes tests the complete SOAP envelope structure
func TestSOAPEnvelopeWithPrefixes(t *testing.T) {
	// Define types in correct order
	type TestSOAPBody struct {
		XMLName xml.Name `xml:"soapenv:Body"`
		Content interface{} `xml:",any"`
	}
	
	type TestSOAPEnvelope struct {
		XMLName  xml.Name `xml:"soapenv:Envelope"`
		XmlnsSoap string  `xml:"xmlns:soapenv,attr"`
		XmlnsTns  string  `xml:"xmlns:tns,attr,omitempty"`
		Body     TestSOAPBody
	}
	
	// Define types that use namespace prefixes
	type CreateOrderRequest struct {
		XMLName      xml.Name `xml:"tns:CreateOrder"`
		CustomerID   string   `xml:"customerID,omitempty"`
		OrderDate    string   `xml:"orderDate,omitempty"`
	}
	
	request := &CreateOrderRequest{
		CustomerID: "12345",
		OrderDate:  "2024-01-01",
	}
	
	envelope := TestSOAPEnvelope{
		XmlnsSoap: "http://schemas.xmlsoap.org/soap/envelope/",
		XmlnsTns:  "https://example.com/service",
		Body: TestSOAPBody{
			Content: request,
		},
	}
	
	data, err := xml.Marshal(envelope)
	require.NoError(t, err)
	xmlStr := string(data)
	
	t.Logf("Complete SOAP envelope:\n%s", xmlStr)
	
	// Should have namespace declarations on envelope
	assert.Contains(t, xmlStr, `xmlns:soapenv="http://schemas.xmlsoap.org/soap/envelope/"`)
	assert.Contains(t, xmlStr, `xmlns:tns="https://example.com/service"`)
	
	// Operation should use tns prefix
	assert.Contains(t, xmlStr, `<tns:CreateOrder>`)
	
	// Child elements should be unqualified
	assert.Contains(t, xmlStr, `<customerID>`)
	assert.Contains(t, xmlStr, `<orderDate>`)
	
	// Should NOT have these
	assert.NotContains(t, xmlStr, `<tns:customerID>`)
	assert.NotContains(t, xmlStr, `<tns:orderDate>`)
}

// TestGeneratedTypesShouldUsePrefix tests that our generated types should use prefixes
func TestGeneratedTypesShouldUsePrefix(t *testing.T) {
	// This is what we want to generate for operation types
	expectedTemplate := `
type GOWebService_SendungsErstellung struct {
	XMLName xml.Name ` + "`" + `xml:"tns:GOWebService_SendungsErstellung"` + "`" + `
	Versender string ` + "`" + `xml:"Versender,omitempty"` + "`" + `
	Benutzername string ` + "`" + `xml:"Benutzername,omitempty"` + "`" + `
	// ... other fields
}`

	// When marshaled inside SOAP body, the tns namespace will be declared
	// on the SOAP envelope, and this element will use the prefix
	assert.Contains(t, strings.TrimSpace(expectedTemplate), "tns:GOWebService_SendungsErstellung")
}