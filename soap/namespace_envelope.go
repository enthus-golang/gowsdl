// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at http://mozilla.org/MPL/2.0/.

package soap

import (
	"encoding/xml"
)

// EnhancedSOAPEnvelope is a SOAP envelope that supports additional namespace declarations
// for proper handling of elementFormDefault="unqualified" schemas
type EnhancedSOAPEnvelope struct {
	XMLName xml.Name `xml:"soap:Envelope"`
	
	// Standard SOAP namespace
	XMLNSSoap string `xml:"xmlns:soap,attr"`
	
	// Target namespace for the service (optional)
	XMLNSTns string `xml:"xmlns:tns,attr,omitempty"`
	
	// Common XML schema namespaces (optional)
	XMLNSXSI string `xml:"xmlns:xsi,attr,omitempty"`
	XMLNSXSD string `xml:"xmlns:xsd,attr,omitempty"`

	Header *SOAPHeader
	Body   SOAPBody
}

// NewEnhancedSOAPEnvelope creates a SOAP envelope with namespace support
func NewEnhancedSOAPEnvelope(targetNamespace string) *EnhancedSOAPEnvelope {
	env := &EnhancedSOAPEnvelope{
		XMLNSSoap: XmlNsSoapEnv,
	}
	
	if targetNamespace != "" {
		env.XMLNSTns = targetNamespace
		// Add common XML schema namespaces
		env.XMLNSXSI = "http://www.w3.org/2001/XMLSchema-instance"
		env.XMLNSXSD = "http://www.w3.org/2001/XMLSchema"
	}
	
	return env
}