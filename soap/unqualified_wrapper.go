package soap

import (
	"encoding/xml"
	"reflect"
)

// UnqualifiedRequestWrapper wraps a request to ensure proper namespace handling
// for schemas with elementFormDefault="unqualified"
type UnqualifiedRequestWrapper struct {
	Request         interface{}
	TargetNamespace string
	OperationName   string
}

// MarshalXML implements xml.Marshaler to generate XML with namespace prefix
func (w UnqualifiedRequestWrapper) MarshalXML(e *xml.Encoder, start xml.StartElement) error {
	// Create the start element with namespace prefix
	start.Name = xml.Name{
		Space: w.TargetNamespace,
		Local: w.OperationName,
	}
	
	// Marshal the inner request directly without the XMLName field
	// This ensures child elements are unqualified
	v := reflect.ValueOf(w.Request)
	if v.Kind() == reflect.Ptr {
		v = v.Elem()
	}
	
	// Start the element
	if err := e.EncodeToken(start); err != nil {
		return err
	}
	
	// Encode fields, skipping XMLName
	t := v.Type()
	for i := 0; i < v.NumField(); i++ {
		field := t.Field(i)
		if field.Name == "XMLName" {
			continue
		}
		
		value := v.Field(i)
		if err := e.EncodeElement(value.Interface(), xml.StartElement{
			Name: xml.Name{Local: field.Tag.Get("xml")},
		}); err != nil {
			return err
		}
	}
	
	// End the element
	return e.EncodeToken(start.End())
}