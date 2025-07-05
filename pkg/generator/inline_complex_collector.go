// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at http://mozilla.org/MPL/2.0/.

package generator

import (
	"fmt"
	
	"github.com/enthus-golang/gowsdl/pkg/parser"
	"github.com/enthus-golang/gowsdl/pkg/utils"
)

// InlineComplexType represents an inline complex type that needs to be extracted
type InlineComplexType struct {
	Name          string
	GeneratedName string
	ComplexType   *parser.XSDComplexType
	MinOccurs     string
	MaxOccurs     string
}

// CollectInlineComplexTypes finds all inline complex types in the schema and generates names for them
func CollectInlineComplexTypes(schemas []*parser.XSDSchema) map[string]*InlineComplexType {
	inlineTypes := make(map[string]*InlineComplexType)
	
	for _, schema := range schemas {
		// Process complex types
		for _, ct := range schema.ComplexTypes {
			collectFromComplexType(ct.Name, ct, inlineTypes)
		}
		
		// Process elements
		for _, el := range schema.Elements {
			if el.ComplexType != nil {
				// Global element with inline complex type
				typeName := utils.MakePublic(el.Name) + "Type"
				key := fmt.Sprintf("element_%s", el.Name)
				inlineTypes[key] = &InlineComplexType{
					Name:          el.Name,
					GeneratedName: typeName,
					ComplexType:   el.ComplexType,
					MinOccurs:     el.MinOccurs,
					MaxOccurs:     el.MaxOccurs,
				}
				// Also collect nested inline types within this element
				collectFromComplexType(el.Name, el.ComplexType, inlineTypes)
			}
		}
	}
	
	return inlineTypes
}

// collectFromComplexType recursively collects inline complex types from a complex type
func collectFromComplexType(parentTypeName string, ct *parser.XSDComplexType, inlineTypes map[string]*InlineComplexType) {
	// Process sequences
	for _, seq := range ct.Sequence {
		if seq.ComplexType != nil {
			// Generate a name based on parent type and element name
			typeName := utils.MakePublic(parentTypeName) + utils.MakePublic(seq.Name) + "Type"
			key := fmt.Sprintf("%s_%s", parentTypeName, seq.Name)
			
			inlineTypes[key] = &InlineComplexType{
				Name:          seq.Name,
				GeneratedName: typeName,
				ComplexType:   seq.ComplexType,
				MinOccurs:     seq.MinOccurs,
				MaxOccurs:     seq.MaxOccurs,
			}
			
			// Recursively process nested complex types
			collectFromComplexType(typeName, seq.ComplexType, inlineTypes)
		}
	}
	
	// Process choices
	for _, choice := range ct.Choice {
		if choice.ComplexType != nil {
			typeName := utils.MakePublic(parentTypeName) + utils.MakePublic(choice.Name) + "Type"
			key := fmt.Sprintf("%s_%s", parentTypeName, choice.Name)
			
			inlineTypes[key] = &InlineComplexType{
				Name:          choice.Name,
				GeneratedName: typeName,
				ComplexType:   choice.ComplexType,
				MinOccurs:     choice.MinOccurs,
				MaxOccurs:     choice.MaxOccurs,
			}
			
			collectFromComplexType(typeName, choice.ComplexType, inlineTypes)
		}
	}
	
	// Process all
	for _, all := range ct.All {
		if all.ComplexType != nil {
			typeName := utils.MakePublic(parentTypeName) + utils.MakePublic(all.Name) + "Type"
			key := fmt.Sprintf("%s_%s", parentTypeName, all.Name)
			
			inlineTypes[key] = &InlineComplexType{
				Name:          all.Name,
				GeneratedName: typeName,
				ComplexType:   all.ComplexType,
				MinOccurs:     all.MinOccurs,
				MaxOccurs:     all.MaxOccurs,
			}
			
			collectFromComplexType(typeName, all.ComplexType, inlineTypes)
		}
	}
	
	// Process complex content extensions
	if ct.ComplexContent.Extension.Base != "" {
		for _, seq := range ct.ComplexContent.Extension.Sequence {
			if seq.ComplexType != nil {
				typeName := utils.MakePublic(parentTypeName) + utils.MakePublic(seq.Name) + "Type"
				key := fmt.Sprintf("%s_%s", parentTypeName, seq.Name)
				
				inlineTypes[key] = &InlineComplexType{
					Name:          seq.Name,
					GeneratedName: typeName,
					ComplexType:   seq.ComplexType,
					MinOccurs:     seq.MinOccurs,
					MaxOccurs:     seq.MaxOccurs,
				}
				
				collectFromComplexType(typeName, seq.ComplexType, inlineTypes)
			}
		}
	}
}

// GetInlineTypeKey generates a unique key for an inline type based on its context
func GetInlineTypeKey(parentName, elementName string) string {
	if parentName == "" {
		return fmt.Sprintf("element_%s", elementName)
	}
	return fmt.Sprintf("%s_%s", parentName, elementName)
}

// FindInlineType finds the generated type name for an inline complex type
func FindInlineType(inlineTypes map[string]*InlineComplexType, parentName, elementName string) (string, bool) {
	key := GetInlineTypeKey(parentName, elementName)
	if it, ok := inlineTypes[key]; ok {
		return it.GeneratedName, true
	}
	return "", false
}

// IsOptionalType checks if the type should use a pointer
func IsOptionalType(minOccurs string, maxOccurs string) bool {
	return minOccurs == "0" && maxOccurs != "unbounded"
}