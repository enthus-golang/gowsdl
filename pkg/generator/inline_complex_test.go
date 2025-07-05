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

func TestInlineComplexTypeGeneration(t *testing.T) {
	wsdl := `<?xml version="1.0" encoding="UTF-8"?>
<definitions xmlns="http://schemas.xmlsoap.org/wsdl/"
             xmlns:soap="http://schemas.xmlsoap.org/wsdl/soap/"
             xmlns:tns="http://example.com/order"
             xmlns:xsd="http://www.w3.org/2001/XMLSchema"
             targetNamespace="http://example.com/order">
    <types>
        <schema xmlns="http://www.w3.org/2001/XMLSchema"
                targetNamespace="http://example.com/order">
            
            <!-- Complex type with inline optional complex type -->
            <complexType name="OrderType">
                <sequence>
                    <element name="orderId" type="string"/>
                    <element name="shippingAddress" minOccurs="0">
                        <complexType>
                            <sequence>
                                <element name="street" type="string"/>
                                <element name="city" type="string"/>
                                <element name="zipCode" type="string" minOccurs="0"/>
                            </sequence>
                        </complexType>
                    </element>
                    <element name="billingAddress">
                        <complexType>
                            <sequence>
                                <element name="street" type="string"/>
                                <element name="city" type="string"/>
                            </sequence>
                        </complexType>
                    </element>
                    <element name="items" maxOccurs="unbounded">
                        <complexType>
                            <sequence>
                                <element name="itemId" type="string"/>
                                <element name="quantity" type="int"/>
                            </sequence>
                        </complexType>
                    </element>
                </sequence>
            </complexType>
            
            <!-- Element with inline complex type -->
            <element name="CreateOrderRequest">
                <complexType>
                    <sequence>
                        <element name="order" type="tns:OrderType"/>
                    </sequence>
                </complexType>
            </element>
            
            <element name="CreateOrderResponse">
                <complexType>
                    <sequence>
                        <element name="success" type="boolean"/>
                    </sequence>
                </complexType>
            </element>
        </schema>
    </types>
    
    <message name="CreateOrderRequestMessage">
        <part name="parameters" element="tns:CreateOrderRequest"/>
    </message>
    
    <message name="CreateOrderResponseMessage">
        <part name="parameters" element="tns:CreateOrderResponse"/>
    </message>
    
    <portType name="OrderServicePortType">
        <operation name="CreateOrder">
            <input message="tns:CreateOrderRequestMessage"/>
            <output message="tns:CreateOrderResponseMessage"/>
        </operation>
    </portType>
    
    <binding name="OrderServiceBinding" type="tns:OrderServicePortType">
        <soap:binding style="document" transport="http://schemas.xmlsoap.org/soap/http"/>
        <operation name="CreateOrder">
            <soap:operation soapAction="http://example.com/order/CreateOrder"/>
            <input><soap:body use="literal"/></input>
            <output><soap:body use="literal"/></output>
        </operation>
    </binding>
</definitions>`

	file, err := os.CreateTemp("", "test-wsdl-*.wsdl")
	require.NoError(t, err)
	defer func() { _ = os.Remove(file.Name()) }()
	defer func() { _ = file.Close() }()
	
	_, err = file.WriteString(wsdl)
	require.NoError(t, err)

	g, err := New(file.Name(), WithPackage("order"))
	require.NoError(t, err)

	files, err := g.Generate(context.Background())
	require.NoError(t, err)

	// Check types file
	types, ok := files["types"]
	require.True(t, ok, "types file should be generated")
	
	typesStr := string(types)
	t.Logf("Generated types:\n%s", typesStr)
	
	// Check that named types are generated for inline complex types
	assert.Contains(t, typesStr, "type OrderTypeShippingAddressType struct", "Should generate named type for optional inline complex type")
	assert.Contains(t, typesStr, "type OrderTypeBillingAddressType struct", "Should generate named type for required inline complex type")
	assert.Contains(t, typesStr, "type OrderTypeItemsType struct", "Should generate named type for array inline complex type")
	
	// Check that the OrderType uses pointers to named types
	assert.Contains(t, typesStr, "ShippingAddress *OrderTypeShippingAddressType", "Optional field should be pointer to named type")
	assert.Contains(t, typesStr, "BillingAddress OrderTypeBillingAddressType", "Required field should be value type")
	assert.Contains(t, typesStr, "Items []OrderTypeItemsType", "Array field should use named type")
	
	// Check that inline types for elements are also handled
	assert.Contains(t, typesStr, "type CreateOrderRequestType struct", "Should generate named type for element's inline complex type")
	assert.Contains(t, typesStr, "type CreateOrderResponseType struct", "Should generate named type for element's inline complex type")
}

func TestInlineComplexTypeMarshalingWithNamedTypes(t *testing.T) {
	// This test verifies that the generated code with named types marshals correctly
	xmlData := `<order>
		<orderId>12345</orderId>
		<billingAddress>
			<street>Main St</street>
			<city>NYC</city>
		</billingAddress>
		<items>
			<itemId>item1</itemId>
			<quantity>2</quantity>
		</items>
		<items>
			<itemId>item2</itemId>
			<quantity>1</quantity>
		</items>
	</order>`
	
	// Define the expected structure with named types
	type OrderTypeBillingAddressType struct {
		Street string `xml:"street,omitempty"`
		City   string `xml:"city,omitempty"`
	}
	
	type OrderTypeShippingAddressType struct {
		Street  string `xml:"street,omitempty"`
		City    string `xml:"city,omitempty"`
		ZipCode string `xml:"zipCode,omitempty"`
	}
	
	type OrderTypeItemsType struct {
		ItemId   string `xml:"itemId,omitempty"`
		Quantity int    `xml:"quantity,omitempty"`
	}
	
	type OrderType struct {
		OrderId         string                        `xml:"orderId,omitempty"`
		ShippingAddress *OrderTypeShippingAddressType `xml:"shippingAddress,omitempty"`
		BillingAddress  OrderTypeBillingAddressType  `xml:"billingAddress,omitempty"`
		Items           []OrderTypeItemsType          `xml:"items,omitempty"`
	}
	
	// Test unmarshaling
	var order OrderType
	err := xml.Unmarshal([]byte(xmlData), &order)
	require.NoError(t, err)
	
	assert.Equal(t, "12345", order.OrderId)
	assert.Nil(t, order.ShippingAddress) // Not in XML
	assert.Equal(t, "Main St", order.BillingAddress.Street)
	assert.Equal(t, "NYC", order.BillingAddress.City)
	assert.Len(t, order.Items, 2)
	assert.Equal(t, "item1", order.Items[0].ItemId)
	assert.Equal(t, 2, order.Items[0].Quantity)
	
	// Test marshaling with optional field
	order.ShippingAddress = &OrderTypeShippingAddressType{
		Street:  "Ship St",
		City:    "LA",
		ZipCode: "90210",
	}
	
	data, err := xml.Marshal(order)
	require.NoError(t, err)
	
	xmlStr := string(data)
	assert.Contains(t, xmlStr, "<shippingAddress>")
	assert.Contains(t, xmlStr, "<street>Ship St</street>")
	assert.Contains(t, xmlStr, "<city>LA</city>")
	assert.Contains(t, xmlStr, "<zipCode>90210</zipCode>")
	
	// Test nil pointer is omitted
	order.ShippingAddress = nil
	data, err = xml.Marshal(order)
	require.NoError(t, err)
	
	xmlStr = string(data)
	assert.NotContains(t, xmlStr, "<shippingAddress")
}

func TestNestedInlineComplexTypes(t *testing.T) {
	wsdl := `<?xml version="1.0" encoding="UTF-8"?>
<definitions xmlns="http://schemas.xmlsoap.org/wsdl/"
             xmlns:soap="http://schemas.xmlsoap.org/wsdl/soap/"
             xmlns:tns="http://example.com/nested"
             xmlns:xsd="http://www.w3.org/2001/XMLSchema"
             targetNamespace="http://example.com/nested">
    <types>
        <schema xmlns="http://www.w3.org/2001/XMLSchema"
                targetNamespace="http://example.com/nested">
            
            <complexType name="CompanyType">
                <sequence>
                    <element name="name" type="string"/>
                    <element name="headquarters" minOccurs="0">
                        <complexType>
                            <sequence>
                                <element name="address">
                                    <complexType>
                                        <sequence>
                                            <element name="street" type="string"/>
                                            <element name="city" type="string"/>
                                            <element name="coordinates" minOccurs="0">
                                                <complexType>
                                                    <sequence>
                                                        <element name="lat" type="decimal"/>
                                                        <element name="lon" type="decimal"/>
                                                    </sequence>
                                                </complexType>
                                            </element>
                                        </sequence>
                                    </complexType>
                                </element>
                                <element name="phone" type="string"/>
                            </sequence>
                        </complexType>
                    </element>
                </sequence>
            </complexType>
        </schema>
    </types>
</definitions>`

	file, err := os.CreateTemp("", "test-wsdl-*.wsdl")
	require.NoError(t, err)
	defer func() { _ = os.Remove(file.Name()) }()
	defer func() { _ = file.Close() }()
	
	_, err = file.WriteString(wsdl)
	require.NoError(t, err)

	g, err := New(file.Name(), WithPackage("nested"))
	require.NoError(t, err)

	files, err := g.Generate(context.Background())
	require.NoError(t, err)

	types, ok := files["types"]
	require.True(t, ok, "types file should be generated")
	
	typesStr := string(types)
	t.Logf("Generated types:\n%s", typesStr)
	
	// Check nested type generation
	assert.Contains(t, typesStr, "type CompanyTypeHeadquartersType struct", "Should generate type for headquarters")
	assert.Contains(t, typesStr, "type CompanyTypeHeadquartersTypeAddressType struct", "Should generate type for nested address")
	assert.Contains(t, typesStr, "type CompanyTypeHeadquartersTypeAddressTypeCoordinatesType struct", "Should generate type for deeply nested coordinates")
	
	// Check proper usage in parent types
	assert.Contains(t, typesStr, "Headquarters *CompanyTypeHeadquartersType", "Optional headquarters should use pointer")
	assert.Contains(t, typesStr, "Address CompanyTypeHeadquartersTypeAddressType", "Required address should be value type")
	assert.Contains(t, typesStr, "Coordinates *CompanyTypeHeadquartersTypeAddressTypeCoordinatesType", "Optional coordinates should use pointer")
}