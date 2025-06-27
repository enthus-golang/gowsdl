package parser

import (
	"testing"
)

func TestElementReferenceResolution(t *testing.T) {
	// Create test schemas
	schema1 := &XSDSchema{
		TargetNamespace: "http://example.com/schema1",
		Xmlns: map[string]string{
			"tns": "http://example.com/schema1",
			"xsd": "http://www.w3.org/2001/XMLSchema",
		},
		Elements: []*XSDElement{
			{
				Name:     "GlobalElement",
				Type:     "xsd:string",
				Nillable: true,
				MinOccurs: "0",
				MaxOccurs: "1",
			},
		},
		ComplexTypes: []*XSDComplexType{
			{
				Name: "TestType",
				Sequence: []*XSDElement{
					{
						Ref: "tns:GlobalElement",
					},
				},
			},
		},
	}

	schema2 := &XSDSchema{
		TargetNamespace: "http://example.com/schema2",
		Xmlns: map[string]string{
			"s1": "http://example.com/schema1",
			"xsd": "http://www.w3.org/2001/XMLSchema",
		},
		ComplexTypes: []*XSDComplexType{
			{
				Name: "CrossSchemaType",
				Sequence: []*XSDElement{
					{
						Ref: "s1:GlobalElement",
					},
				},
			},
		},
	}

	schemas := []*XSDSchema{schema1, schema2}
	
	// Test element reference resolution
	t.Run("ResolveElementReference", func(t *testing.T) {
		traverser := NewTraverser(schema1, schemas)
		traverser.Traverse()
		
		// Check if the reference was resolved
		complexType := schema1.ComplexTypes[0]
		element := complexType.Sequence[0]
		
		if element.Name != "GlobalElement" {
			t.Errorf("Expected element name to be 'GlobalElement', got '%s'", element.Name)
		}
		if element.Type != "xsd:string" {
			t.Errorf("Expected element type to be 'xsd:string', got '%s'", element.Type)
		}
		if !element.Nillable {
			t.Error("Expected element to be nillable")
		}
	})

	t.Run("CrossSchemaElementReference", func(t *testing.T) {
		traverser := NewTraverser(schema2, schemas)
		traverser.Traverse()
		
		// Check if the cross-schema reference was resolved
		complexType := schema2.ComplexTypes[0]
		element := complexType.Sequence[0]
		
		if element.Name != "GlobalElement" {
			t.Errorf("Expected element name to be 'GlobalElement', got '%s'", element.Name)
		}
		if element.Type != "xsd:string" {
			t.Errorf("Expected element type to be 'xsd:string', got '%s'", element.Type)
		}
	})

	t.Run("ElementWithComplexType", func(t *testing.T) {
		// Test element reference with complex type
		schema3 := &XSDSchema{
			TargetNamespace: "http://example.com/schema3",
			Xmlns: map[string]string{
				"tns": "http://example.com/schema3",
			},
			Elements: []*XSDElement{
				{
					Name: "ComplexElement",
					ComplexType: &XSDComplexType{
						Sequence: []*XSDElement{
							{Name: "Field1", Type: "xsd:string"},
							{Name: "Field2", Type: "xsd:int"},
						},
					},
				},
			},
			ComplexTypes: []*XSDComplexType{
				{
					Name: "ReferenceType",
					Sequence: []*XSDElement{
						{
							Ref: "tns:ComplexElement",
						},
					},
				},
			},
		}
		
		traverser := NewTraverser(schema3, []*XSDSchema{schema3})
		traverser.Traverse()
		
		// Check if complex type was copied
		complexType := schema3.ComplexTypes[0]
		element := complexType.Sequence[0]
		
		if element.ComplexType == nil {
			t.Error("Expected element to have ComplexType")
		}
		if len(element.ComplexType.Sequence) != 2 {
			t.Errorf("Expected 2 sequence elements, got %d", len(element.ComplexType.Sequence))
		}
	})
}