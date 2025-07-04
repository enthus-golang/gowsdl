package generator

import (
	"context"
	"encoding/xml"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestOptionalComplexNamedTypes(t *testing.T) {
	wsdl := `<?xml version="1.0" encoding="UTF-8"?>
<definitions xmlns="http://schemas.xmlsoap.org/wsdl/"
    xmlns:soap="http://schemas.xmlsoap.org/wsdl/soap/"
    xmlns:tns="http://example.com/test"
    xmlns:xsd="http://www.w3.org/2001/XMLSchema"
    targetNamespace="http://example.com/test">
    
    <types>
        <xsd:schema targetNamespace="http://example.com/test">
            <!-- Test type with optional complex elements -->
            <xsd:complexType name="OrderType">
                <xsd:sequence>
                    <xsd:element name="orderId" type="xsd:string"/>
                    <!-- Optional complex type (minOccurs="0") -->
                    <xsd:element name="shippingAddress" minOccurs="0">
                        <xsd:complexType>
                            <xsd:sequence>
                                <xsd:element name="street" type="xsd:string"/>
                                <xsd:element name="city" type="xsd:string"/>
                                <xsd:element name="zipCode" type="xsd:string" minOccurs="0"/>
                            </xsd:sequence>
                        </xsd:complexType>
                    </xsd:element>
                    <!-- Required complex type (minOccurs="1" or not specified) -->
                    <xsd:element name="billingAddress">
                        <xsd:complexType>
                            <xsd:sequence>
                                <xsd:element name="street" type="xsd:string"/>
                                <xsd:element name="city" type="xsd:string"/>
                            </xsd:sequence>
                        </xsd:complexType>
                    </xsd:element>
                    <!-- Optional complex type with attributes -->
                    <xsd:element name="discount" minOccurs="0">
                        <xsd:complexType>
                            <xsd:simpleContent>
                                <xsd:extension base="xsd:decimal">
                                    <xsd:attribute name="currency" type="xsd:string"/>
                                </xsd:extension>
                            </xsd:simpleContent>
                        </xsd:complexType>
                    </xsd:element>
                </xsd:sequence>
            </xsd:complexType>
            
            <xsd:element name="createOrder" type="tns:OrderType"/>
            <xsd:element name="createOrderResponse" type="xsd:string"/>
        </xsd:schema>
    </types>
    
    <message name="createOrderRequest">
        <part name="parameters" element="tns:createOrder"/>
    </message>
    <message name="createOrderResponse">
        <part name="parameters" element="tns:createOrderResponse"/>
    </message>
    
    <portType name="OrderServicePortType">
        <operation name="createOrder">
            <input message="tns:createOrderRequest"/>
            <output message="tns:createOrderResponse"/>
        </operation>
    </portType>
    
    <binding name="OrderServiceBinding" type="tns:OrderServicePortType">
        <soap:binding style="document" transport="http://schemas.xmlsoap.org/soap/http"/>
        <operation name="createOrder">
            <soap:operation soapAction="createOrder"/>
            <input><soap:body use="literal"/></input>
            <output><soap:body use="literal"/></output>
        </operation>
    </binding>
    
    <service name="OrderService">
        <port name="OrderServicePort" binding="tns:OrderServiceBinding">
            <soap:address location="http://example.com/order"/>
        </port>
    </service>
</definitions>`

	// Write WSDL to temp file
	tempFile, err := os.CreateTemp("", "test-optional-complex-named-*.wsdl")
	require.NoError(t, err)
	defer func() {
		_ = os.Remove(tempFile.Name())
	}()

	_, err = tempFile.WriteString(wsdl)
	require.NoError(t, err)
	err = tempFile.Close()
	require.NoError(t, err)

	// Create generator
	g, err := New(tempFile.Name(), WithPackage("test"))
	require.NoError(t, err)

	// Generate code
	ctx := context.Background()
	files, err := g.Generate(ctx)
	require.NoError(t, err)

	// Find the types content
	typesContent, ok := files["types"]
	require.True(t, ok, "types file not found")
	
	typesStr := string(typesContent)

	// Check that optional complex types have named types generated
	assert.Contains(t, typesStr, "type OrderTypeShippingAddressType struct {",
		"Optional complex type should have a named type")
	
	assert.Contains(t, typesStr, "type OrderTypeDiscountType struct {",
		"Optional complex type with simpleContent should have a named type")

	// Check that the MarshalXML method is generated for optional types
	assert.Contains(t, typesStr, "func (t OrderTypeShippingAddressType) MarshalXML",
		"Optional complex type should have MarshalXML method")
	
	assert.Contains(t, typesStr, "func (t OrderTypeDiscountType) MarshalXML",
		"Optional complex type with simpleContent should have MarshalXML method")

	// Check that required complex type doesn't have a named type (stays inline)
	assert.NotContains(t, typesStr, "type OrderTypeBillingAddressType",
		"Required complex type should not have a named type")

	// Check that fields use the named types
	assert.Contains(t, typesStr, "ShippingAddress OrderTypeShippingAddressType",
		"Field should use the named type")
}

// Define types outside the test function
type testOrderTypeShippingAddressType struct {
	Street  string `xml:"street,omitempty"`
	City    string `xml:"city,omitempty"`
	ZipCode string `xml:"zipCode,omitempty"`
}

// MarshalXML implements xml.Marshaler
func (t testOrderTypeShippingAddressType) MarshalXML(e *xml.Encoder, start xml.StartElement) error {
	// Check if all fields are empty
	if t.Street == "" && t.City == "" && t.ZipCode == "" {
		return nil // Skip marshaling
	}
	// Otherwise use default marshaling
	type alias testOrderTypeShippingAddressType
	return e.EncodeElement(alias(t), start)
}

func TestOptionalComplexEmptyMarshaling(t *testing.T) {
	// Test that empty optional complex types are omitted from XML
	type testStruct struct {
		XMLName          xml.Name                           `xml:"order"`
		OrderId          string                             `xml:"orderId"`
		ShippingAddress  testOrderTypeShippingAddressType   `xml:"shippingAddress,omitempty"`
	}

	// Test with empty shipping address
	order1 := testStruct{
		OrderId: "12345",
		// ShippingAddress is zero value (all fields empty)
	}

	data1, err := xml.Marshal(order1)
	require.NoError(t, err)
	xmlStr1 := string(data1)
	
	// Should NOT contain empty shippingAddress element
	assert.NotContains(t, xmlStr1, "<shippingAddress>")
	assert.NotContains(t, xmlStr1, "<shippingAddress/>")
	assert.Contains(t, xmlStr1, "<orderId>12345</orderId>")

	// Test with populated shipping address
	order2 := testStruct{
		OrderId: "12345",
		ShippingAddress: testOrderTypeShippingAddressType{
			Street: "123 Main St",
			City:   "Anytown",
		},
	}

	data2, err := xml.Marshal(order2)
	require.NoError(t, err)
	xmlStr2 := string(data2)
	
	// Should contain shippingAddress element with data
	assert.Contains(t, xmlStr2, "<shippingAddress>")
	assert.Contains(t, xmlStr2, "<street>123 Main St</street>")
	assert.Contains(t, xmlStr2, "<city>Anytown</city>")
}