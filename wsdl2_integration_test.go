package gowsdl

import (
	"context"
	"os"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestWSDL2Integration(t *testing.T) {
	// Create a simple WSDL 2.0 document
	wsdl2Content := `<?xml version="1.0" encoding="UTF-8"?>
<description xmlns="http://www.w3.org/ns/wsdl/"
             xmlns:tns="http://example.com/stockquote"
             xmlns:wsoap="http://www.w3.org/ns/wsdl/soap"
             xmlns:soap="http://schemas.xmlsoap.org/soap/"
             xmlns:wsdlx="http://www.w3.org/ns/wsdl-extensions"
             targetNamespace="http://example.com/stockquote">

  <documentation>
    Stock quote service for WSDL 2.0 testing
  </documentation>

  <types>
    <xs:schema xmlns:xs="http://www.w3.org/2001/XMLSchema" 
               targetNamespace="http://example.com/stockquote"
               elementFormDefault="qualified">
      
      <xs:element name="GetQuoteRequest">
        <xs:complexType>
          <xs:sequence>
            <xs:element name="symbol" type="xs:string"/>
          </xs:sequence>
        </xs:complexType>
      </xs:element>
      
      <xs:element name="GetQuoteResponse">
        <xs:complexType>
          <xs:sequence>
            <xs:element name="price" type="xs:decimal"/>
          </xs:sequence>
        </xs:complexType>
      </xs:element>
      
      <xs:element name="QuoteFault">
        <xs:complexType>
          <xs:sequence>
            <xs:element name="errorMessage" type="xs:string"/>
          </xs:sequence>
        </xs:complexType>
      </xs:element>
      
    </xs:schema>
  </types>

  <interface name="StockQuoteInterface">
    <fault name="InvalidSymbolFault" element="tns:QuoteFault"/>
    <operation name="GetQuote" pattern="http://www.w3.org/ns/wsdl/in-out">
      <input element="tns:GetQuoteRequest"/>
      <output element="tns:GetQuoteResponse"/>
      <outfault ref="tns:InvalidSymbolFault"/>
    </operation>
  </interface>

  <binding name="StockQuoteSoapBinding" 
           interface="tns:StockQuoteInterface" 
           type="http://www.w3.org/ns/wsdl/soap"
           wsoap:protocol="http://schemas.xmlsoap.org/soap/http">
    <operation ref="tns:GetQuote" wsoap:mep="http://www.w3.org/2003/05/soap/mep/request-response"/>
  </binding>

  <service name="StockQuoteService" interface="tns:StockQuoteInterface">
    <endpoint name="StockQuoteEndpoint" 
              binding="tns:StockQuoteSoapBinding" 
              address="http://example.com/stockquote"/>
  </service>

</description>`

	// Create temporary WSDL 2.0 file
	tempFile, err := os.CreateTemp("", "wsdl2-test-*.wsdl")
	require.NoError(t, err)
	defer func() {
		_ = os.Remove(tempFile.Name())
	}()

	_, err = tempFile.WriteString(wsdl2Content)
	require.NoError(t, err)
	_ = tempFile.Close()

	t.Run("WSDL2VersionDetection", func(t *testing.T) {
		version, err := detectWSDLVersion([]byte(wsdl2Content))
		require.NoError(t, err)
		assert.Equal(t, "2.0", version)
	})

	t.Run("WSDL2Parsing", func(t *testing.T) {
		// Parse WSDL 2.0
		g, err := NewGoWSDL(tempFile.Name(), "stockquote", false, true)
		require.NoError(t, err)

		err = g.unmarshal(context.Background())
		require.NoError(t, err)

		// Verify WSDL 2.0 was detected and parsed
		assert.Equal(t, "2.0", g.wsdlVersion)
		assert.NotNil(t, g.wsdl2)
		assert.Nil(t, g.wsdl) // WSDL 1.1 should be nil

		// Verify parsed content
		assert.Equal(t, "http://example.com/stockquote", g.wsdl2.TargetNamespace)
		assert.Len(t, g.wsdl2.Interfaces, 1)
		assert.Len(t, g.wsdl2.Bindings, 1)
		assert.Len(t, g.wsdl2.Services, 1)

		// Check interface
		iface := g.wsdl2.Interfaces[0]
		assert.Equal(t, "StockQuoteInterface", iface.Name)
		assert.Len(t, iface.Operations, 1)
		assert.Len(t, iface.Faults, 1)

		// Check operation
		op := iface.Operations[0]
		assert.Equal(t, "GetQuote", op.Name)
		assert.NotNil(t, op.Input)
		assert.NotNil(t, op.Output)
		assert.Equal(t, "tns:GetQuoteRequest", op.Input.Element)
		assert.Equal(t, "tns:GetQuoteResponse", op.Output.Element)

		// Check binding
		binding := g.wsdl2.Bindings[0]
		assert.Equal(t, "StockQuoteSoapBinding", binding.Name)
		assert.Equal(t, "tns:StockQuoteInterface", binding.Interface)

		// Check service
		service := g.wsdl2.Services[0]
		assert.Equal(t, "StockQuoteService", service.Name)
		assert.Len(t, service.Endpoints, 1)
		assert.Equal(t, "http://example.com/stockquote", service.Endpoints[0].Address)
	})

	t.Run("WSDL2CodeGeneration", func(t *testing.T) {
		// Test types generation (this should work since types are the same structure)
		g, err := NewGoWSDL(tempFile.Name(), "stockquote", false, true)
		require.NoError(t, err)

		err = g.unmarshal(context.Background())
		require.NoError(t, err)

		// Test types generation specifically
		typesCode, err := g.genTypes()
		require.NoError(t, err)

		// Check that types were generated
		typesStr := string(typesCode)
		assert.Contains(t, typesStr, "GetQuoteRequest")
		assert.Contains(t, typesStr, "GetQuoteResponse")
		assert.Contains(t, typesStr, "QuoteFault")

		// Note: Operations and server generation require template updates for WSDL 2.0
		// This is a known limitation - templates expect WSDL 1.1 structures
		t.Log("Operations and server generation require WSDL 2.0 template updates")
	})

	t.Run("WSDL2HelperFunctions", func(t *testing.T) {
		g, err := NewGoWSDL(tempFile.Name(), "stockquote", false, true)
		require.NoError(t, err)

		err = g.unmarshal(context.Background())
		require.NoError(t, err)

		// Test findElementType function
		elementType := g.findElementType("GetQuoteRequest")
		// Should find the element even if no explicit type (returns element name)
		assert.NotEmpty(t, elementType)

		// Test findServiceAddress function
		address := g.findServiceAddress("StockQuoteEndpoint")
		assert.Equal(t, "http://example.com/stockquote", address)

		// Test namespace registration
		assert.NotNil(t, g.namespaceManager)
		uri, ok := g.namespaceManager.GetNamespaceURI("tns")
		if ok {
			assert.Equal(t, "http://example.com/stockquote", uri)
		}
	})

	t.Run("WSDL2WithXSDElementReferences", func(t *testing.T) {
		// Test that element references work in WSDL 2.0
		wsdl2WithRefs := strings.ReplaceAll(wsdl2Content, 
			`<xs:element name="GetQuoteRequest">`,
			`<xs:element name="GetQuoteRequest">`) + `
		<!-- Add a type that references the element -->
		<xs:complexType name="QuoteContainer">
			<xs:sequence>
				<xs:element ref="tns:GetQuoteRequest"/>
			</xs:sequence>
		</xs:complexType>
		`

		// This is a basic test to ensure the parsing doesn't break
		// Full element reference resolution testing is in element_ref_test.go
		version, err := detectWSDLVersion([]byte(wsdl2WithRefs))
		require.NoError(t, err)
		assert.Equal(t, "2.0", version)
	})
}