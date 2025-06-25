// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at http://mozilla.org/MPL/2.0/.

package gowsdl

import (
	"bytes"
	"encoding/xml"
	"fmt"
	"io"
)

// detectWSDLVersion detects whether the WSDL is version 1.1 or 2.0
func detectWSDLVersion(data []byte) (string, error) {
	decoder := xml.NewDecoder(bytes.NewReader(data))
	
	for {
		tok, err := decoder.Token()
		if err == io.EOF {
			break
		}
		if err != nil {
			return "", fmt.Errorf("error detecting WSDL version: %w", err)
		}
		
		if se, ok := tok.(xml.StartElement); ok {
			// Check the namespace of the root element
			if se.Name.Local == "definitions" && se.Name.Space == wsdlNamespace {
				return "1.1", nil
			}
			if se.Name.Local == "description" && se.Name.Space == wsdl2Namespace {
				return "2.0", nil
			}
			
			// Also check default namespace
			for _, attr := range se.Attr {
				if attr.Name.Local == "xmlns" {
					switch attr.Value {
					case wsdlNamespace:
						return "1.1", nil
					case wsdl2Namespace:
						return "2.0", nil
					}
				}
			}
			
			// If we've processed the root element and didn't find a match
			return "", fmt.Errorf("unknown WSDL version or invalid WSDL document")
		}
	}
	
	return "", fmt.Errorf("unable to determine WSDL version")
}