package soap

import (
	"encoding/xml"
)

// SOAPEnvelopeWithNamespaces is an enhanced SOAP envelope that supports target namespace declarations
type SOAPEnvelopeWithNamespaces struct {
	XMLName       xml.Name `xml:"soap:Envelope"`
	XmlNSSoap     string   `xml:"xmlns:soap,attr"`
	XmlNSTns      string   `xml:"xmlns:tns,attr,omitempty"`
	XmlNSXSI      string   `xml:"xmlns:xsi,attr,omitempty"`
	XmlNSXSD      string   `xml:"xmlns:xsd,attr,omitempty"`
	
	Header *SOAPHeader
	Body   SOAPBody
}

// NewSOAPEnvelopeWithNamespaces creates a SOAP envelope with namespace declarations
func NewSOAPEnvelopeWithNamespaces(targetNamespace string) *SOAPEnvelopeWithNamespaces {
	envelope := &SOAPEnvelopeWithNamespaces{
		XmlNSSoap: XmlNsSoapEnv,
	}
	
	if targetNamespace != "" {
		envelope.XmlNSTns = targetNamespace
		// Add standard XML schema namespaces that are often needed
		envelope.XmlNSXSI = "http://www.w3.org/2001/XMLSchema-instance"
		envelope.XmlNSXSD = "http://www.w3.org/2001/XMLSchema"
	}
	
	return envelope
}

// NamespacedRequest wraps a request with namespace prefix support
type NamespacedRequest struct {
	XMLName xml.Name
	Content interface{}
}

// WrapRequestWithNamespace wraps a request to use namespace prefix
func WrapRequestWithNamespace(request interface{}, elementName, namespacePrefix string) *NamespacedRequest {
	return &NamespacedRequest{
		XMLName: xml.Name{Local: namespacePrefix + ":" + elementName},
		Content: request,
	}
}