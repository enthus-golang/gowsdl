package generator

import (
	"fmt"
	"strings"

	"github.com/enthus-golang/gowsdl/pkg/parser"
	"github.com/enthus-golang/gowsdl/pkg/utils"
)

// OptionalComplexType represents an optional complex type that needs a named type with MarshalXML
type OptionalComplexType struct {
	ParentType string
	Element    *parser.XSDElement
	TypeName   string
}

// collectOptionalComplexTypes collects all optional inline complex types
func (g *Generator) collectOptionalComplexTypes(data struct {
	SimpleType      []*parser.XSDSimpleType
	ComplexTypes    []*parser.XSDComplexType
	Elements        []*parser.XSDElement
	Schemas         []*parser.XSDSchema
	Messages        []*parser.WSDLMessage
	TargetNamespace string
	OptionalComplexTypes []OptionalComplexType
}) []OptionalComplexType {
	var result []OptionalComplexType
	
	// Check complex types
	for _, ct := range data.ComplexTypes {
		result = append(result, g.collectFromComplexType(ct.Name, ct)...)
	}
	
	// Check global elements
	for _, elem := range data.Elements {
		if elem.ComplexType != nil {
			result = append(result, g.collectFromComplexType(elem.Name, elem.ComplexType)...)
		}
	}
	
	return result
}

// collectFromComplexType recursively collects optional complex types
func (g *Generator) collectFromComplexType(parentTypeName string, ct *parser.XSDComplexType) []OptionalComplexType {
	var result []OptionalComplexType
	
	// Process all elements (sequence, choice, all)
	allElements := append(append(ct.Sequence, ct.Choice...), ct.All...)
	
	for _, elem := range allElements {
		// Check if this is an optional complex type
		if elem.MinOccurs == "0" && elem.ComplexType != nil && elem.MaxOccurs != "unbounded" {
			// Generate type name
			parentName := utils.MakePublic(g.typeMapper.EscapeReservedWord(parentTypeName))
			elemName := utils.MakePublic(g.typeMapper.EscapeReservedWord(elem.Name))
			typeName := fmt.Sprintf("%s%sType", parentName, elemName)
			
			result = append(result, OptionalComplexType{
				ParentType: parentTypeName,
				Element:    elem,
				TypeName:   typeName,
			})
			
			// Recursively check nested elements
			result = append(result, g.collectFromComplexType(typeName, elem.ComplexType)...)
		} else if elem.ComplexType != nil {
			// For non-optional complex types, still check their nested elements
			nestedName := parentTypeName
			if elem.Name != "" {
				nestedName = fmt.Sprintf("%s_%s", parentTypeName, elem.Name)
			}
			result = append(result, g.collectFromComplexType(nestedName, elem.ComplexType)...)
		}
	}
	
	return result
}

// generateEmptyCheck generates the condition to check if all fields are empty
func generateEmptyCheck(elem *parser.XSDElement) string {
	var checks []string
	
	// Check SimpleContent
	if elem.ComplexType.SimpleContent.Extension.Base != "" {
		baseType := strings.ToLower(elem.ComplexType.SimpleContent.Extension.Base)
		if strings.Contains(baseType, "string") {
			checks = append(checks, "t.Value == \"\"")
		} else if strings.Contains(baseType, "decimal") || strings.Contains(baseType, "float") || 
			strings.Contains(baseType, "double") || strings.Contains(baseType, "int") {
			checks = append(checks, "t.Value == 0")
		} else {
			// Default to zero value check for unknown types
			checks = append(checks, "t.Value == 0")
		}
	}
	
	// Check all elements
	allElements := append(append(elem.ComplexType.Sequence, elem.ComplexType.Choice...), elem.ComplexType.All...)
	for _, e := range allElements {
		// Use simple field name transformation for template
		fieldName := strings.Title(e.Name)
		
		// For string types, check if empty
		if e.Type != "" && strings.Contains(strings.ToLower(e.Type), "string") {
			checks = append(checks, fmt.Sprintf("t.%s == \"\"", fieldName))
		} else if e.Type != "" {
			// For numeric types, check if zero
			// This is a simplification - in practice we'd need type-specific checks
			checks = append(checks, fmt.Sprintf("t.%s == 0", fieldName))
		} else if e.ComplexType != nil {
			// For complex types, we'd need to check if they're at zero value
			// For now, skip this check as it's complex
			continue
		} else {
			// Default to string check
			checks = append(checks, fmt.Sprintf("t.%s == \"\"", fieldName))
		}
	}
	
	// Check attributes
	for _, attr := range elem.ComplexType.Attributes {
		attrName := strings.Title(attr.Name)
		checks = append(checks, fmt.Sprintf("t.%s == \"\"", attrName))
	}
	
	if len(checks) == 0 {
		return "false"
	}
	
	return strings.Join(checks, " && ")
}