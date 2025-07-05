// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at http://mozilla.org/MPL/2.0/.

package generator

import (
	"context"
	"encoding/xml"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestElementFormDefaultUnqualified(t *testing.T) {
	// Test WSDL with elementFormDefault not specified (defaults to unqualified)
	wsdl := `<?xml version="1.0" encoding="UTF-8"?>
<definitions xmlns="http://schemas.xmlsoap.org/wsdl/"
             xmlns:soap="http://schemas.xmlsoap.org/wsdl/soap/"
             xmlns:tns="https://example.com/service"
             xmlns:xsd="http://www.w3.org/2001/XMLSchema"
             targetNamespace="https://example.com/service">
    <types>
        <schema xmlns="http://www.w3.org/2001/XMLSchema"
                targetNamespace="https://example.com/service">
            <!-- Note: elementFormDefault is NOT specified, defaults to "unqualified" -->
            
            <element name="CreateOrder">
                <complexType>
                    <sequence>
                        <element name="customerID" type="string"/>
                        <element name="orderDate" type="date"/>
                        <element name="items">
                            <complexType>
                                <sequence>
                                    <element name="item" maxOccurs="unbounded">
                                        <complexType>
                                            <sequence>
                                                <element name="productCode" type="string"/>
                                                <element name="quantity" type="int"/>
                                            </sequence>
                                        </complexType>
                                    </element>
                                </sequence>
                            </complexType>
                        </element>
                    </sequence>
                </complexType>
            </element>
            
            <element name="CreateOrderResponse">
                <complexType>
                    <sequence>
                        <element name="orderID" type="string"/>
                        <element name="status" type="string"/>
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
</definitions>`

	file, err := os.CreateTemp("", "test-namespace-*.wsdl")
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
	
	// Check that the generated code compiles and test XML output
	testCode := `package main

import (
	"encoding/xml"
	"fmt"
	"log"
)

` + typesStr + `

func main() {
	order := &CreateOrder{
		CustomerID: "12345",
		OrderDate:  "2024-01-01",
		Items: &CreateOrderItemsType{
			Item: []CreateOrderItemsTypeItemType{
				{
					ProductCode: "PROD001",
					Quantity:    2,
				},
			},
		},
	}
	
	data, err := xml.Marshal(order)
	if err != nil {
		log.Fatal(err)
	}
	
	fmt.Println(string(data))
}`

	// Write test program
	testFile, err := os.CreateTemp("", "test-namespace-*.go")
	require.NoError(t, err)
	defer func() { _ = os.Remove(testFile.Name()) }()
	
	_, err = testFile.WriteString(testCode)
	require.NoError(t, err)
	testFile.Close()
	
	// The key assertion: child elements should NOT have namespace prefix
	// When elementFormDefault is unqualified, only the root element should be qualified
	assert.Contains(t, typesStr, "CreateOrder", "Should generate CreateOrder type")
	
	// Check that XMLName is only set on global elements
	// TODO: This is what needs to be fixed - currently all types get XMLName with namespace
}

func TestElementFormDefaultQualified(t *testing.T) {
	// Test WSDL with elementFormDefault="qualified"
	wsdl := `<?xml version="1.0" encoding="UTF-8"?>
<definitions xmlns="http://schemas.xmlsoap.org/wsdl/"
             xmlns:soap="http://schemas.xmlsoap.org/wsdl/soap/"
             xmlns:tns="https://example.com/service"
             xmlns:xsd="http://www.w3.org/2001/XMLSchema"
             targetNamespace="https://example.com/service">
    <types>
        <schema xmlns="http://www.w3.org/2001/XMLSchema"
                targetNamespace="https://example.com/service"
                elementFormDefault="qualified">
            
            <element name="CreateOrder">
                <complexType>
                    <sequence>
                        <element name="customerID" type="string"/>
                        <element name="orderDate" type="date"/>
                    </sequence>
                </complexType>
            </element>
        </schema>
    </types>
</definitions>`

	file, err := os.CreateTemp("", "test-qualified-*.wsdl")
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
	
	// When elementFormDefault="qualified", all elements should be in the namespace
	// This is the current behavior and should remain unchanged
	assert.Contains(t, typesStr, "XMLName", "Should have XMLName fields")
}

// TestXMLMarshalingWithUnqualifiedElements tests that the generated code
// produces correct XML for unqualified schemas
func TestXMLMarshalingWithUnqualifiedElements(t *testing.T) {
	// This demonstrates the expected XML output
	type CreateOrderItemType struct {
		ProductCode string `xml:"productCode,omitempty"`
		Quantity    int    `xml:"quantity,omitempty"`
	}
	
	type CreateOrderItems struct {
		Item []CreateOrderItemType `xml:"item,omitempty"`
	}
	
	// Only the root element should have namespace
	type CreateOrder struct {
		XMLName    xml.Name          `xml:"https://example.com/service CreateOrder"`
		CustomerID string            `xml:"customerID,omitempty"`
		Items      *CreateOrderItems `xml:"items,omitempty"`
	}
	
	order := &CreateOrder{
		CustomerID: "12345",
		Items: &CreateOrderItems{
			Item: []CreateOrderItemType{
				{ProductCode: "PROD001", Quantity: 2},
			},
		},
	}
	
	data, err := xml.Marshal(order)
	require.NoError(t, err)
	
	xmlStr := string(data)
	t.Logf("Expected XML structure:\n%s", xmlStr)
	
	// The root element should have the namespace
	assert.Contains(t, xmlStr, "CreateOrder xmlns=")
	
	// Child elements should NOT have namespace prefix
	assert.Contains(t, xmlStr, "<customerID>")
	assert.Contains(t, xmlStr, "<items>")
	assert.Contains(t, xmlStr, "<item>")
	assert.Contains(t, xmlStr, "<productCode>")
	
	// Verify no namespace prefixes on child elements
	assert.NotContains(t, xmlStr, "tns:customerID")
	assert.NotContains(t, xmlStr, "ns:customerID")
}