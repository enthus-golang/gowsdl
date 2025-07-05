package generator

import (
	"context"
	"encoding/xml"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestOptionalComplexPointerTypes(t *testing.T) {
	wsdl := `<?xml version="1.0" encoding="UTF-8"?>
<definitions xmlns="http://schemas.xmlsoap.org/wsdl/" 
    xmlns:soap="http://schemas.xmlsoap.org/wsdl/soap/"
    xmlns:tns="http://example.com/test"
    targetNamespace="http://example.com/test">
    <types>
        <schema xmlns="http://www.w3.org/2001/XMLSchema" targetNamespace="http://example.com/test">
            <complexType name="OrderType">
                <sequence>
                    <element name="orderId" type="string" minOccurs="1"/>
                    <element name="shippingAddress" minOccurs="0">
                        <complexType>
                            <sequence>
                                <element name="street" type="string"/>
                                <element name="city" type="string"/>
                                <element name="zipCode" type="string"/>
                            </sequence>
                        </complexType>
                    </element>
                    <element name="billingAddress" minOccurs="1">
                        <complexType>
                            <sequence>
                                <element name="street" type="string"/>
                                <element name="city" type="string"/>
                            </sequence>
                        </complexType>
                    </element>
                    <element name="discount" minOccurs="0">
                        <complexType>
                            <simpleContent>
                                <extension base="decimal">
                                    <attribute name="currency" type="string"/>
                                </extension>
                            </simpleContent>
                        </complexType>
                    </element>
                </sequence>
            </complexType>
            
            <element name="createOrder" type="tns:OrderType"/>
            <element name="createOrderResponse" type="string"/>
        </schema>
    </types>
</definitions>`

	// Create temp file
	tmpFile, err := os.CreateTemp("", "test-wsdl-*.wsdl")
	require.NoError(t, err)
	defer func() { _ = os.Remove(tmpFile.Name()) }()

	_, err = tmpFile.Write([]byte(wsdl))
	require.NoError(t, err)
	require.NoError(t, tmpFile.Close())

	g, err := New(tmpFile.Name())
	require.NoError(t, err)

	code, err := g.Generate(context.TODO())
	require.NoError(t, err)
	
	// Get the generated code - concatenate all files
	var codeStr string
	for filename, content := range code {
		t.Logf("Generated file: %s", filename)
		codeStr += string(content) + "\n"
	}
	
	// Check that optional complex types use pointers to named types
	assert.Contains(t, codeStr, "ShippingAddress *OrderTypeShippingAddressType", "Optional complex type should be a pointer to named type")
	assert.Contains(t, codeStr, "BillingAddress OrderTypeBillingAddressType", "Required complex type should not be a pointer")
	assert.Contains(t, codeStr, "Discount *OrderTypeDiscountType", "Optional simple content should be a pointer")
}

func TestOptionalComplexPointerMarshaling(t *testing.T) {
	// Define types that would be generated
	type AddressType struct {
		Street  string `xml:"street,omitempty"`
		City    string `xml:"city,omitempty"`
		ZipCode string `xml:"zipCode,omitempty"`
	}

	type OrderType struct {
		XMLName         xml.Name     `xml:"order"`
		OrderId         string       `xml:"orderId,omitempty"`
		ShippingAddress *AddressType `xml:"shippingAddress,omitempty"`
		BillingAddress  AddressType  `xml:"billingAddress,omitempty"`
	}

	// Test 1: With nil optional field
	order1 := OrderType{
		OrderId: "12345",
		BillingAddress: AddressType{
			Street: "Main St",
			City:   "NYC",
		},
		// ShippingAddress is nil
	}

	data1, err := xml.Marshal(order1)
	require.NoError(t, err)
	
	// Should NOT contain empty shippingAddress element
	xmlStr1 := string(data1)
	assert.NotContains(t, xmlStr1, "<shippingAddress/>")
	assert.NotContains(t, xmlStr1, "<shippingAddress></shippingAddress>")
	assert.Contains(t, xmlStr1, "<billingAddress>")

	// Test 2: With non-nil optional field
	order2 := OrderType{
		OrderId: "12345",
		ShippingAddress: &AddressType{
			Street:  "Ship St",
			City:    "LA",
			ZipCode: "90210",
		},
		BillingAddress: AddressType{
			Street: "Bill St",
			City:   "NYC",
		},
	}

	data2, err := xml.Marshal(order2)
	require.NoError(t, err)
	
	// Should contain shippingAddress element with data
	xmlStr2 := string(data2)
	assert.Contains(t, xmlStr2, "<shippingAddress>")
	assert.Contains(t, xmlStr2, "<street>Ship St</street>")
	assert.Contains(t, xmlStr2, "<city>LA</city>")
	assert.Contains(t, xmlStr2, "<zipCode>90210</zipCode>")

	// Test 3: Unmarshal works correctly
	var order3 OrderType
	err = xml.Unmarshal(data2, &order3)
	require.NoError(t, err)
	require.NotNil(t, order3.ShippingAddress)
	assert.Equal(t, "Ship St", order3.ShippingAddress.Street)
	assert.Equal(t, "LA", order3.ShippingAddress.City)

	// Test 4: Unmarshal without optional field
	xmlWithoutOptional := `<order>
		<orderId>999</orderId>
		<billingAddress>
			<street>Test St</street>
			<city>Test City</city>
		</billingAddress>
	</order>`
	
	var order4 OrderType
	err = xml.Unmarshal([]byte(xmlWithoutOptional), &order4)
	require.NoError(t, err)
	assert.Nil(t, order4.ShippingAddress, "Optional field should be nil when not present")
	assert.Equal(t, "Test St", order4.BillingAddress.Street)
}

func TestPointerSafetyDocumentation(t *testing.T) {
	// This test serves as documentation for the breaking change
	type Order struct {
		ShippingAddress *struct {
			Street string `xml:"street,omitempty"`
			City   string `xml:"city,omitempty"`
		} `xml:"shippingAddress,omitempty"`
	}

	var order Order
	
	// Old code would panic:
	// _ = order.ShippingAddress.Street // panic: nil pointer dereference
	
	// New code must check for nil:
	if order.ShippingAddress != nil {
		_ = order.ShippingAddress.Street // Safe
	}
	
	// Or use a helper method (if we add them later):
	// street := order.GetShippingAddressStreet() // Returns "" if nil
}