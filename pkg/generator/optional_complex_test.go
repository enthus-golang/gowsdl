package generator

import (
	"context"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestOptionalComplexTypes(t *testing.T) {
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
	tempFile, err := os.CreateTemp("", "test-optional-complex-*.wsdl")
	require.NoError(t, err)
	defer os.Remove(tempFile.Name())

	_, err = tempFile.WriteString(wsdl)
	require.NoError(t, err)
	tempFile.Close()

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

	// Check that optional complex type uses pointer
	assert.Contains(t, typesStr, "ShippingAddress *struct {",
		"Optional complex type should be a pointer")

	// Check that required complex type is not a pointer
	assert.Contains(t, typesStr, "BillingAddress struct {",
		"Required complex type should not be a pointer")

	// Check that optional complex type with simpleContent uses pointer
	assert.Contains(t, typesStr, "Discount *struct {",
		"Optional complex type with simpleContent should be a pointer")

	// Verify the struct fields have proper XML tags
	assert.Contains(t, typesStr, "`xml:\"shippingAddress,omitempty\"")
	assert.Contains(t, typesStr, "`xml:\"billingAddress,omitempty\"")
	assert.Contains(t, typesStr, "`xml:\"discount,omitempty\"")
}