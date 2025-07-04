package generator

import (
	"strings"
	"testing"

	"github.com/enthus-golang/gowsdl/pkg/parser"
	"github.com/enthus-golang/gowsdl/pkg/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCollectOptionalComplexTypes(t *testing.T) {
	g := &Generator{
		typeMapper: types.NewTypeMapper(), // Initialize the type mapper
	}

	// Test data with various scenarios
	complexTypes := []*parser.XSDComplexType{
		{
			Name: "OrderType",
			Sequence: []*parser.XSDElement{
				{
					Name:      "optionalAddress",
					MinOccurs: "0",
					ComplexType: &parser.XSDComplexType{
						Sequence: []*parser.XSDElement{
							{Name: "street", Type: "xsd:string"},
						},
					},
				},
				{
					Name:      "requiredAddress", 
					MinOccurs: "1",
					ComplexType: &parser.XSDComplexType{
						Sequence: []*parser.XSDElement{
							{Name: "street", Type: "xsd:string"},
						},
					},
				},
				{
					Name:      "unboundedItems",
					MinOccurs: "0",
					MaxOccurs: "unbounded",
					ComplexType: &parser.XSDComplexType{
						Sequence: []*parser.XSDElement{
							{Name: "item", Type: "xsd:string"},
						},
					},
				},
			},
		},
	}

	elements := []*parser.XSDElement{
		{
			Name: "globalElement",
			ComplexType: &parser.XSDComplexType{
				Sequence: []*parser.XSDElement{
					{
						Name:      "optionalField",
						MinOccurs: "0",
						ComplexType: &parser.XSDComplexType{
							Sequence: []*parser.XSDElement{
								{Name: "value", Type: "xsd:string"},
							},
						},
					},
				},
			},
		},
	}

	result := g.collectOptionalComplexTypes(complexTypes, elements)

	// Should find 2 optional complex types:
	// 1. OrderType.optionalAddress (minOccurs="0", not unbounded)
	// 2. globalElement.optionalField (minOccurs="0", not unbounded)
	require.Len(t, result, 2)

	// Check first optional type
	assert.Equal(t, "OrderType", result[0].ParentType)
	assert.Equal(t, "optionalAddress", result[0].Element.Name)
	assert.Equal(t, "OrderTypeOptionalAddressType", result[0].TypeName)

	// Check second optional type
	assert.Equal(t, "globalElement", result[1].ParentType)
	assert.Equal(t, "optionalField", result[1].Element.Name)
	assert.Equal(t, "GlobalElementOptionalFieldType", result[1].TypeName)
}

func TestGenerateEmptyCheck(t *testing.T) {
	tests := []struct {
		name     string
		element  *parser.XSDElement
		expected string
	}{
		{
			name: "simple_content_string",
			element: &parser.XSDElement{
				ComplexType: &parser.XSDComplexType{
					SimpleContent: parser.XSDSimpleContent{
						Extension: parser.XSDExtension{
							Base: "xsd:string",
						},
					},
				},
			},
			expected: "t.Value == \"\"",
		},
		{
			name: "simple_content_decimal",
			element: &parser.XSDElement{
				ComplexType: &parser.XSDComplexType{
					SimpleContent: parser.XSDSimpleContent{
						Extension: parser.XSDExtension{
							Base: "xsd:decimal",
						},
					},
				},
			},
			expected: "t.Value == 0",
		},
		{
			name: "simple_content_float",
			element: &parser.XSDElement{
				ComplexType: &parser.XSDComplexType{
					SimpleContent: parser.XSDSimpleContent{
						Extension: parser.XSDExtension{
							Base: "xsd:float",
						},
					},
				},
			},
			expected: "t.Value == 0",
		},
		{
			name: "simple_content_double", 
			element: &parser.XSDElement{
				ComplexType: &parser.XSDComplexType{
					SimpleContent: parser.XSDSimpleContent{
						Extension: parser.XSDExtension{
							Base: "xsd:double",
						},
					},
				},
			},
			expected: "t.Value == 0",
		},
		{
			name: "simple_content_int",
			element: &parser.XSDElement{
				ComplexType: &parser.XSDComplexType{
					SimpleContent: parser.XSDSimpleContent{
						Extension: parser.XSDExtension{
							Base: "xsd:int",
						},
					},
				},
			},
			expected: "t.Value == 0",
		},
		{
			name: "string_elements",
			element: &parser.XSDElement{
				ComplexType: &parser.XSDComplexType{
					Sequence: []*parser.XSDElement{
						{Name: "name", Type: "xsd:string"},
						{Name: "description", Type: "string"},
					},
				},
			},
			expected: "t.Name == \"\" && t.Description == \"\"",
		},
		{
			name: "numeric_elements",
			element: &parser.XSDElement{
				ComplexType: &parser.XSDComplexType{
					Sequence: []*parser.XSDElement{
						{Name: "price", Type: "xsd:decimal"},
						{Name: "quantity", Type: "xsd:int"},
						{Name: "weight", Type: "xsd:float"},
						{Name: "height", Type: "xsd:double"},
						{Name: "length", Type: "xsd:long"},
						{Name: "width", Type: "xsd:short"},
						{Name: "depth", Type: "xsd:byte"},
					},
				},
			},
			expected: "t.Price == 0 && t.Quantity == 0 && t.Weight == 0 && t.Height == 0 && t.Length == 0 && t.Width == 0 && t.Depth == 0",
		},
		{
			name: "boolean_element",
			element: &parser.XSDElement{
				ComplexType: &parser.XSDComplexType{
					Sequence: []*parser.XSDElement{
						{Name: "active", Type: "xsd:boolean"},
						{Name: "enabled", Type: "bool"},
					},
				},
			},
			expected: "!t.Active && !t.Enabled",
		},
		{
			name: "date_time_elements",
			element: &parser.XSDElement{
				ComplexType: &parser.XSDComplexType{
					Sequence: []*parser.XSDElement{
						{Name: "createdAt", Type: "xsd:dateTime"},
						{Name: "updatedAt", Type: "xsd:date"},
						{Name: "scheduledTime", Type: "xsd:time"},
					},
				},
			},
			expected: "t.CreatedAt.IsZero() && t.UpdatedAt.IsZero() && t.ScheduledTime.IsZero()",
		},
		{
			name: "mixed_attributes",
			element: &parser.XSDElement{
				ComplexType: &parser.XSDComplexType{
					Attributes: []*parser.XSDAttribute{
						{Name: "id", Type: "xsd:string"},
						{Name: "version", Type: "xsd:int"},
						{Name: "enabled", Type: "xsd:boolean"},
						{Name: "rate", Type: "xsd:decimal"},
					},
				},
			},
			expected: "t.Id == \"\" && t.Version == 0 && !t.Enabled && t.Rate == 0",
		},
		{
			name: "complex_with_nested_complex_type",
			element: &parser.XSDElement{
				ComplexType: &parser.XSDComplexType{
					Sequence: []*parser.XSDElement{
						{Name: "name", Type: "xsd:string"},
						{
							Name: "address",
							ComplexType: &parser.XSDComplexType{
								Sequence: []*parser.XSDElement{
									{Name: "street", Type: "xsd:string"},
								},
							},
						},
					},
				},
			},
			expected: "t.Name == \"\"", // nested complex type should be skipped
		},
		{
			name: "element_with_ref",
			element: &parser.XSDElement{
				ComplexType: &parser.XSDComplexType{
					Sequence: []*parser.XSDElement{
						{Name: "name", Type: "xsd:string"},
						{Name: "refElement", Ref: "tns:SomeElement"},
					},
				},
			},
			expected: "t.Name == \"\"", // ref element should be skipped
		},
		{
			name: "element_without_type",
			element: &parser.XSDElement{
				ComplexType: &parser.XSDComplexType{
					Sequence: []*parser.XSDElement{
						{Name: "unknownElement"}, // no Type or ComplexType or Ref
					},
				},
			},
			expected: "t.UnknownElement == \"\"", // default to string check
		},
		{
			name: "unknown_type",
			element: &parser.XSDElement{
				ComplexType: &parser.XSDComplexType{
					Sequence: []*parser.XSDElement{
						{Name: "customField", Type: "custom:CustomType"},
					},
				},
			},
			expected: "t.CustomField == \"\"", // unknown type defaults to string check
		},
		{
			name: "choice_elements",
			element: &parser.XSDElement{
				ComplexType: &parser.XSDComplexType{
					Choice: []*parser.XSDElement{
						{Name: "option1", Type: "xsd:string"},
						{Name: "option2", Type: "xsd:int"},
					},
				},
			},
			expected: "t.Option1 == \"\" && t.Option2 == 0",
		},
		{
			name: "all_elements",
			element: &parser.XSDElement{
				ComplexType: &parser.XSDComplexType{
					All: []*parser.XSDElement{
						{Name: "field1", Type: "xsd:string"},
						{Name: "field2", Type: "xsd:boolean"},
					},
				},
			},
			expected: "t.Field1 == \"\" && !t.Field2",
		},
		{
			name: "empty_complex_type",
			element: &parser.XSDElement{
				ComplexType: &parser.XSDComplexType{},
			},
			expected: "false", // no elements or attributes to check
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := generateEmptyCheck(tt.element)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestGenerateEmptyCheckWithSimpleContentAndAttributes(t *testing.T) {
	element := &parser.XSDElement{
		ComplexType: &parser.XSDComplexType{
			SimpleContent: parser.XSDSimpleContent{
				Extension: parser.XSDExtension{
					Base: "xsd:decimal",
					Attributes: []*parser.XSDAttribute{
						{Name: "currency", Type: "xsd:string"},
						{Name: "precision", Type: "xsd:int"},
					},
				},
			},
		},
	}

	result := generateEmptyCheck(element)
	expected := "t.Value == 0 && t.Currency == \"\" && t.Precision == 0"
	assert.Equal(t, expected, result)
}

func TestCollectFromComplexTypeRecursive(t *testing.T) {
	g := &Generator{
		typeMapper: types.NewTypeMapper(),
	}

	// Create a complex type with nested optional complex types
	complexType := &parser.XSDComplexType{
		Sequence: []*parser.XSDElement{
			{
				Name:      "level1Optional",
				MinOccurs: "0",
				ComplexType: &parser.XSDComplexType{
					Sequence: []*parser.XSDElement{
						{
							Name:      "level2Optional",
							MinOccurs: "0",
							ComplexType: &parser.XSDComplexType{
								Sequence: []*parser.XSDElement{
									{Name: "value", Type: "xsd:string"},
								},
							},
						},
						{Name: "level2Required", Type: "xsd:string"},
					},
				},
			},
			{Name: "level1Required", Type: "xsd:string"},
		},
	}

	result := g.collectFromComplexType("TestType", complexType)

	// Should collect both level1Optional and level2Optional
	require.Len(t, result, 2)

	// First should be level1Optional
	assert.Equal(t, "TestType", result[0].ParentType)
	assert.Equal(t, "level1Optional", result[0].Element.Name)
	assert.Equal(t, "TestTypeLevel1OptionalType", result[0].TypeName)

	// Second should be level2Optional (nested)
	assert.Equal(t, "TestTypeLevel1OptionalType", result[1].ParentType)
	assert.Equal(t, "level2Optional", result[1].Element.Name)
	assert.Equal(t, "TestTypeLevel1OptionalTypeLevel2OptionalType", result[1].TypeName)
}

func TestCollectFromComplexTypeChoiceAndAll(t *testing.T) {
	g := &Generator{
		typeMapper: types.NewTypeMapper(),
	}

	complexType := &parser.XSDComplexType{
		Choice: []*parser.XSDElement{
			{
				Name:      "choiceOptional",
				MinOccurs: "0",
				ComplexType: &parser.XSDComplexType{
					Sequence: []*parser.XSDElement{
						{Name: "value", Type: "xsd:string"},
					},
				},
			},
		},
		All: []*parser.XSDElement{
			{
				Name:      "allOptional",
				MinOccurs: "0",
				ComplexType: &parser.XSDComplexType{
					Sequence: []*parser.XSDElement{
						{Name: "value", Type: "xsd:string"},
					},
				},
			},
		},
	}

	result := g.collectFromComplexType("TestType", complexType)

	// Should collect both choiceOptional and allOptional
	require.Len(t, result, 2)

	names := []string{result[0].Element.Name, result[1].Element.Name}
	assert.Contains(t, names, "choiceOptional")
	assert.Contains(t, names, "allOptional")
}

func TestGenerateEmptyCheckEdgeCases(t *testing.T) {
	tests := []struct {
		name        string
		element     *parser.XSDElement
		expectedLen int // expected number of conditions
	}{
		{
			name: "nil_complex_type",
			element: &parser.XSDElement{
				ComplexType: nil,
			},
			expectedLen: 0,
		},
		{
			name: "nil_simple_content", 
			element: &parser.XSDElement{
				ComplexType: &parser.XSDComplexType{
					// SimpleContent is a value type, so can't be nil
					// Instead test with empty SimpleContent
					SimpleContent: parser.XSDSimpleContent{},
				},
			},
			expectedLen: 0,
		},
		{
			name: "empty_base_in_simple_content",
			element: &parser.XSDElement{
				ComplexType: &parser.XSDComplexType{
					SimpleContent: parser.XSDSimpleContent{
						Extension: parser.XSDExtension{
							Base: "",
						},
					},
				},
			},
			expectedLen: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := generateEmptyCheck(tt.element)
			if tt.expectedLen == 0 {
				assert.Equal(t, "false", result)
			} else {
				conditions := strings.Split(result, " && ")
				assert.Len(t, conditions, tt.expectedLen)
			}
		})
	}
}