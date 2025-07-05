// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at http://mozilla.org/MPL/2.0/.

package generator

import (
	"testing"

	"github.com/enthus-golang/gowsdl/pkg/parser"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCollectFromComplexType(t *testing.T) {
	tests := []struct {
		name             string
		parentTypeName   string
		complexType      *parser.XSDComplexType
		expectedTypes    map[string]string // key -> generated name
		expectedOptional map[string]bool   // key -> is optional
	}{
		{
			name:           "complex type with choice elements",
			parentTypeName: "OrderType",
			complexType: &parser.XSDComplexType{
				Choice: []*parser.XSDElement{
					{
						Name:      "primaryAddress",
						MinOccurs: "0",
						ComplexType: &parser.XSDComplexType{
							Sequence: []*parser.XSDElement{
								{Name: "street", Type: "string"},
							},
						},
					},
					{
						Name: "secondaryAddress",
						ComplexType: &parser.XSDComplexType{
							Sequence: []*parser.XSDElement{
								{Name: "street", Type: "string"},
							},
						},
					},
				},
			},
			expectedTypes: map[string]string{
				"OrderType_primaryAddress":   "OrderTypePrimaryAddressType",
				"OrderType_secondaryAddress": "OrderTypeSecondaryAddressType",
			},
			expectedOptional: map[string]bool{
				"OrderType_primaryAddress":   true,
				"OrderType_secondaryAddress": false,
			},
		},
		{
			name:           "complex type with all elements",
			parentTypeName: "CompanyType",
			complexType: &parser.XSDComplexType{
				All: []*parser.XSDElement{
					{
						Name:      "headquarters",
						MinOccurs: "0",
						ComplexType: &parser.XSDComplexType{
							Sequence: []*parser.XSDElement{
								{Name: "city", Type: "string"},
							},
						},
					},
				},
			},
			expectedTypes: map[string]string{
				"CompanyType_headquarters": "CompanyTypeHeadquartersType",
			},
			expectedOptional: map[string]bool{
				"CompanyType_headquarters": true,
			},
		},
		{
			name:           "complex type with extension",
			parentTypeName: "ExtendedType",
			complexType: &parser.XSDComplexType{
				ComplexContent: parser.XSDComplexContent{
					Extension: parser.XSDExtension{
						Base: "BaseType",
						Sequence: []*parser.XSDElement{
							{
								Name: "extendedField",
								ComplexType: &parser.XSDComplexType{
									Sequence: []*parser.XSDElement{
										{Name: "value", Type: "string"},
									},
								},
							},
						},
					},
				},
			},
			expectedTypes: map[string]string{
				"ExtendedType_extendedField": "ExtendedTypeExtendedFieldType",
			},
			expectedOptional: map[string]bool{
				"ExtendedType_extendedField": false,
			},
		},
		{
			name:           "nested inline complex types",
			parentTypeName: "RootType",
			complexType: &parser.XSDComplexType{
				Sequence: []*parser.XSDElement{
					{
						Name: "level1",
						ComplexType: &parser.XSDComplexType{
							Sequence: []*parser.XSDElement{
								{
									Name:      "level2",
									MinOccurs: "0",
									ComplexType: &parser.XSDComplexType{
										Sequence: []*parser.XSDElement{
											{Name: "value", Type: "string"},
										},
									},
								},
							},
						},
					},
				},
			},
			expectedTypes: map[string]string{
				"RootType_level1":            "RootTypeLevel1Type",
				"RootTypeLevel1Type_level2": "RootTypeLevel1TypeLevel2Type",
			},
			expectedOptional: map[string]bool{
				"RootType_level1":            false,
				"RootTypeLevel1Type_level2": true,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			inlineTypes := make(map[string]*InlineComplexType)
			collectFromComplexType(tt.parentTypeName, tt.complexType, inlineTypes)

			// Check that all expected types were collected
			for key, expectedName := range tt.expectedTypes {
				it, ok := inlineTypes[key]
				require.True(t, ok, "Expected inline type with key %s", key)
				assert.Equal(t, expectedName, it.GeneratedName)
				
				// Check optional status
				if expectedOptional, hasOptional := tt.expectedOptional[key]; hasOptional {
					isOptional := IsOptionalType(it.MinOccurs, it.MaxOccurs)
					assert.Equal(t, expectedOptional, isOptional, 
						"Type %s optional status mismatch", key)
				}
			}

			// Check no extra types were collected
			assert.Equal(t, len(tt.expectedTypes), len(inlineTypes))
		})
	}
}

func TestGetInlineTypeKey(t *testing.T) {
	tests := []struct {
		parentName  string
		elementName string
		expected    string
	}{
		{
			parentName:  "OrderType",
			elementName: "shippingAddress",
			expected:    "OrderType_shippingAddress",
		},
		{
			parentName:  "",
			elementName: "globalElement",
			expected:    "element_globalElement",
		},
		{
			parentName:  "CompanyType",
			elementName: "headquarters",
			expected:    "CompanyType_headquarters",
		},
	}

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			result := GetInlineTypeKey(tt.parentName, tt.elementName)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestFindInlineType(t *testing.T) {
	inlineTypes := map[string]*InlineComplexType{
		"OrderType_shipping": {
			GeneratedName: "OrderTypeShippingType",
		},
		"element_global": {
			GeneratedName: "GlobalType",
		},
	}

	tests := []struct {
		parentName  string
		elementName string
		expected    string
		found       bool
	}{
		{
			parentName:  "OrderType",
			elementName: "shipping",
			expected:    "OrderTypeShippingType",
			found:       true,
		},
		{
			parentName:  "",
			elementName: "global",
			expected:    "GlobalType",
			found:       true,
		},
		{
			parentName:  "OrderType",
			elementName: "billing",
			expected:    "",
			found:       false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.elementName, func(t *testing.T) {
			result, found := FindInlineType(inlineTypes, tt.parentName, tt.elementName)
			assert.Equal(t, tt.found, found)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestIsOptionalType(t *testing.T) {
	tests := []struct {
		name      string
		minOccurs string
		maxOccurs string
		expected  bool
	}{
		{
			name:      "minOccurs=0",
			minOccurs: "0",
			maxOccurs: "1",
			expected:  true,
		},
		{
			name:      "minOccurs=1",
			minOccurs: "1",
			maxOccurs: "1",
			expected:  false,
		},
		{
			name:      "minOccurs empty (defaults to 1)",
			minOccurs: "",
			maxOccurs: "1",
			expected:  false,
		},
		{
			name:      "unbounded array is not optional",
			minOccurs: "0",
			maxOccurs: "unbounded",
			expected:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := IsOptionalType(tt.minOccurs, tt.maxOccurs)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestCollectInlineComplexTypesWithSimpleContent(t *testing.T) {
	// Test for complex types with simple content
	schemas := []*parser.XSDSchema{
		{
			ComplexTypes: []*parser.XSDComplexType{
				{
					Name: "PriceType",
					Sequence: []*parser.XSDElement{
						{
							Name:      "amount",
							MinOccurs: "0",
							ComplexType: &parser.XSDComplexType{
								SimpleContent: parser.XSDSimpleContent{
									Extension: parser.XSDExtension{
										Base: "decimal",
										Attributes: []*parser.XSDAttribute{
											{Name: "currency", Type: "string"},
										},
									},
								},
							},
						},
					},
				},
			},
		},
	}

	inlineTypes := CollectInlineComplexTypes(schemas)
	
	// Should find the inline simple content type
	key := "PriceType_amount"
	it, ok := inlineTypes[key]
	require.True(t, ok, "Should find inline simple content type")
	assert.Equal(t, "PriceTypeAmountType", it.GeneratedName)
	assert.True(t, IsOptionalType(it.MinOccurs, it.MaxOccurs))
	assert.NotNil(t, it.ComplexType.SimpleContent)
}

func TestCollectInlineComplexTypesFromElements(t *testing.T) {
	// Test for global elements with inline complex types
	schemas := []*parser.XSDSchema{
		{
			Elements: []*parser.XSDElement{
				{
					Name: "CreateOrderRequest",
					ComplexType: &parser.XSDComplexType{
						Sequence: []*parser.XSDElement{
							{
								Name: "orderDetails",
								ComplexType: &parser.XSDComplexType{
									Sequence: []*parser.XSDElement{
										{Name: "total", Type: "decimal"},
									},
								},
							},
						},
					},
				},
			},
		},
	}

	inlineTypes := CollectInlineComplexTypes(schemas)
	
	// Should find the inline type within the element
	key := "CreateOrderRequest_orderDetails"
	it, ok := inlineTypes[key]
	require.True(t, ok, "Should find inline type within element")
	assert.Equal(t, "CreateOrderRequestOrderDetailsType", it.GeneratedName)
}