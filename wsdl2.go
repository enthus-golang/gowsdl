// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at http://mozilla.org/MPL/2.0/.

package gowsdl

import "encoding/xml"

const wsdl2Namespace = "http://www.w3.org/ns/wsdl/"

// WSDL2 represents the structure of a WSDL 2.0 document
type WSDL2 struct {
	Xmlns           map[string]string `xml:"-"`
	Name            string            `xml:"name,attr"`
	TargetNamespace string            `xml:"targetNamespace,attr"`
	Types           WSDLType          `xml:"http://www.w3.org/ns/wsdl/ types"`
	Interfaces      []*WSDL2Interface `xml:"http://www.w3.org/ns/wsdl/ interface"`
	Bindings        []*WSDL2Binding   `xml:"http://www.w3.org/ns/wsdl/ binding"`
	Services        []*WSDL2Service   `xml:"http://www.w3.org/ns/wsdl/ service"`
	Imports         []*WSDLImport     `xml:"import"`
	Includes        []*WSDL2Include   `xml:"include"`
	Doc             string            `xml:"documentation"`
}

// WSDL2Interface represents an interface in WSDL 2.0 (equivalent to portType in WSDL 1.1)
type WSDL2Interface struct {
	Name       string               `xml:"name,attr"`
	Extends    string               `xml:"extends,attr"`
	Operations []*WSDL2Operation    `xml:"operation"`
	Faults     []*WSDL2InterfaceFault `xml:"fault"`
	Doc        string               `xml:"documentation"`
}

// WSDL2Operation represents an operation in WSDL 2.0
type WSDL2Operation struct {
	Name    string              `xml:"name,attr"`
	Pattern string              `xml:"pattern,attr"`
	Style   string              `xml:"style,attr"`
	Safe    bool                `xml:"safe,attr"`
	Input   *WSDL2OperationMsg  `xml:"input"`
	Output  *WSDL2OperationMsg  `xml:"output"`
	InFault []*WSDL2OperationFaultRef `xml:"infault"`
	OutFault []*WSDL2OperationFaultRef `xml:"outfault"`
	Doc     string              `xml:"documentation"`
}

// WSDL2OperationMsg represents input/output messages in WSDL 2.0
type WSDL2OperationMsg struct {
	MessageLabel string `xml:"messageLabel,attr"`
	Element      string `xml:"element,attr"`
}

// WSDL2OperationFaultRef represents fault references in operations
type WSDL2OperationFaultRef struct {
	Ref          string `xml:"ref,attr"`
	MessageLabel string `xml:"messageLabel,attr"`
}

// WSDL2InterfaceFault represents a fault definition in an interface
type WSDL2InterfaceFault struct {
	Name    string `xml:"name,attr"`
	Element string `xml:"element,attr"`
}

// WSDL2Binding represents a binding in WSDL 2.0
type WSDL2Binding struct {
	Name            string                    `xml:"name,attr"`
	Interface       string                    `xml:"interface,attr"`
	Type            string                    `xml:"type,attr"`
	Operations      []*WSDL2BindingOperation  `xml:"operation"`
	Faults          []*WSDL2BindingFault      `xml:"fault"`
	Doc             string                    `xml:"documentation"`
}

// WSDL2BindingOperation represents binding details for an operation
type WSDL2BindingOperation struct {
	Ref              string                         `xml:"ref,attr"`
	SOAPAction       string                         `xml:"http://www.w3.org/ns/wsdl/soap soapAction,attr"`
	SOAPMEPDefault   string                         `xml:"http://www.w3.org/ns/wsdl/soap mepDefault,attr"`
	Input            *WSDL2BindingOperationMessage  `xml:"input"`
	Output           *WSDL2BindingOperationMessage  `xml:"output"`
	InFault          []*WSDL2BindingOperationFault `xml:"infault"`
	OutFault         []*WSDL2BindingOperationFault `xml:"outfault"`
}

// WSDL2BindingOperationMessage represents input/output binding details
type WSDL2BindingOperationMessage struct {
	ContentType string `xml:"contentType,attr"`
}

// WSDL2BindingOperationFault represents fault binding details
type WSDL2BindingOperationFault struct {
	Ref         string `xml:"ref,attr"`
	ContentType string `xml:"contentType,attr"`
}

// WSDL2BindingFault represents binding-level fault details
type WSDL2BindingFault struct {
	Ref         string `xml:"ref,attr"`
	ContentType string `xml:"contentType,attr"`
}

// WSDL2Service represents a service in WSDL 2.0
type WSDL2Service struct {
	Name       string            `xml:"name,attr"`
	Interface  string            `xml:"interface,attr"`
	Endpoints  []*WSDL2Endpoint  `xml:"endpoint"`
	Doc        string            `xml:"documentation"`
}

// WSDL2Endpoint represents an endpoint in WSDL 2.0 (equivalent to port in WSDL 1.1)
type WSDL2Endpoint struct {
	Name    string `xml:"name,attr"`
	Binding string `xml:"binding,attr"`
	Address string `xml:"address,attr"`
}

// WSDL2Include represents an include element in WSDL 2.0
type WSDL2Include struct {
	Location string `xml:"location,attr"`
}

// UnmarshalXML implements interface xml.Unmarshaler for WSDL2
func (w *WSDL2) UnmarshalXML(d *xml.Decoder, start xml.StartElement) error {
	w.Xmlns = make(map[string]string)
	for _, attr := range start.Attr {
		if attr.Name.Space == "xmlns" {
			w.Xmlns[attr.Name.Local] = attr.Value
			continue
		}

		switch attr.Name.Local {
		case "name":
			w.Name = attr.Value
		case "targetNamespace":
			w.TargetNamespace = attr.Value
		}
	}

Loop:
	for {
		tok, err := d.Token()
		if err != nil {
			return err
		}

		switch t := tok.(type) {
		case xml.StartElement:
			switch {
			case t.Name.Local == "import":
				x := new(WSDLImport)
				if err := d.DecodeElement(x, &t); err != nil {
					return err
				}
				w.Imports = append(w.Imports, x)
			case t.Name.Local == "include":
				x := new(WSDL2Include)
				if err := d.DecodeElement(x, &t); err != nil {
					return err
				}
				w.Includes = append(w.Includes, x)
			case t.Name.Local == "documentation":
				if err := d.DecodeElement(&w.Doc, &t); err != nil {
					return err
				}
			case t.Name.Space == wsdl2Namespace:
				switch t.Name.Local {
				case "types":
					if err := d.DecodeElement(&w.Types, &t); err != nil {
						return err
					}
					for prefix, namespace := range w.Xmlns {
						for _, s := range w.Types.Schemas {
							if _, ok := s.Xmlns[prefix]; !ok {
								s.Xmlns[prefix] = namespace
							}
						}
					}
				case "interface":
					x := new(WSDL2Interface)
					if err := d.DecodeElement(x, &t); err != nil {
						return err
					}
					w.Interfaces = append(w.Interfaces, x)
				case "binding":
					x := new(WSDL2Binding)
					if err := d.DecodeElement(x, &t); err != nil {
						return err
					}
					w.Bindings = append(w.Bindings, x)
				case "service":
					x := new(WSDL2Service)
					if err := d.DecodeElement(x, &t); err != nil {
						return err
					}
					w.Services = append(w.Services, x)
				default:
					if err := d.Skip(); err != nil {
						return err
					}
				}
			default:
				if err := d.Skip(); err != nil {
					return err
				}
			}
		case xml.EndElement:
			break Loop
		}
	}

	return nil
}