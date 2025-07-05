package soap

import (
	"encoding/xml"
)

// EnhancedSOAPEnvelope is a SOAP envelope that supports additional namespace declarations
// for proper handling of elementFormDefault="unqualified" schemas
type EnhancedSOAPEnvelope struct {
	XMLName xml.Name `xml:"soap:Envelope"`
	
	// Standard SOAP namespace
	XmlNSSoap string `xml:"xmlns:soap,attr"`
	
	// Target namespace for the service (optional)
	XmlNSTns string `xml:"xmlns:tns,attr,omitempty"`
	
	// Common XML schema namespaces (optional)
	XmlNSXSI string `xml:"xmlns:xsi,attr,omitempty"`
	XmlNSXSD string `xml:"xmlns:xsd,attr,omitempty"`

	Header *SOAPHeader
	Body   SOAPBody
}

// NewEnhancedSOAPEnvelope creates a SOAP envelope with namespace support
func NewEnhancedSOAPEnvelope(targetNamespace string) *EnhancedSOAPEnvelope {
	env := &EnhancedSOAPEnvelope{
		XmlNSSoap: XmlNsSoapEnv,
	}
	
	if targetNamespace != "" {
		env.XmlNSTns = targetNamespace
		// Add common XML schema namespaces
		env.XmlNSXSI = "http://www.w3.org/2001/XMLSchema-instance"
		env.XmlNSXSD = "http://www.w3.org/2001/XMLSchema"
	}
	
	return env
}