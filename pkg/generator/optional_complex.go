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
func (g *Generator) collectOptionalComplexTypes(complexTypes []*parser.XSDComplexType, elements []*parser.XSDElement) []OptionalComplexType {
	var result []OptionalComplexType
	
	// Check complex types
	for _, ct := range complexTypes {
		result = append(result, g.collectFromComplexType(ct.Name, ct)...)
	}
	
	// Check global elements
	for _, elem := range elements {
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
// Note: This function has known limitations:
// - Nested complex types are skipped (would require recursive checking)
// - Element references (ref) are skipped (type information not available)
// - Default values are not considered
// For a more complete solution, consider generating IsEmpty methods on the types themselves
func generateEmptyCheck(elem *parser.XSDElement) string {
	var checks []string
	
	// Check if element or complex type is nil
	if elem == nil || elem.ComplexType == nil {
		return "false"
	}
	
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
		
		// Also check attributes in SimpleContent extension
		for _, attr := range elem.ComplexType.SimpleContent.Extension.Attributes {
			attrName := utils.MakePublic(attr.Name)
			typeLower := strings.ToLower(attr.Type)
			
			if strings.Contains(typeLower, "string") {
				checks = append(checks, fmt.Sprintf("t.%s == \"\"", attrName))
			} else if strings.Contains(typeLower, "decimal") || strings.Contains(typeLower, "float") || 
				strings.Contains(typeLower, "double") {
				checks = append(checks, fmt.Sprintf("t.%s == 0", attrName))
			} else if strings.Contains(typeLower, "int") || strings.Contains(typeLower, "long") || 
				strings.Contains(typeLower, "short") || strings.Contains(typeLower, "byte") {
				checks = append(checks, fmt.Sprintf("t.%s == 0", attrName))
			} else if strings.Contains(typeLower, "bool") {
				checks = append(checks, fmt.Sprintf("!t.%s", attrName))
			} else {
				// Default to string check
				checks = append(checks, fmt.Sprintf("t.%s == \"\"", attrName))
			}
		}
	}
	
	// Check all elements
	allElements := append(append(elem.ComplexType.Sequence, elem.ComplexType.Choice...), elem.ComplexType.All...)
	for _, e := range allElements {
		// Use simple field name transformation for template
		fieldName := utils.MakePublic(e.Name)
		
		// For string types, check if empty
		if e.Type != "" {
			typeLower := strings.ToLower(e.Type)
			if strings.Contains(typeLower, "string") {
				checks = append(checks, fmt.Sprintf("t.%s == \"\"", fieldName))
			} else if strings.Contains(typeLower, "decimal") || strings.Contains(typeLower, "float") || 
				strings.Contains(typeLower, "double") {
				checks = append(checks, fmt.Sprintf("t.%s == 0", fieldName))
			} else if strings.Contains(typeLower, "int") || strings.Contains(typeLower, "long") || 
				strings.Contains(typeLower, "short") || strings.Contains(typeLower, "byte") {
				checks = append(checks, fmt.Sprintf("t.%s == 0", fieldName))
			} else if strings.Contains(typeLower, "bool") {
				checks = append(checks, fmt.Sprintf("!t.%s", fieldName))
			} else if strings.Contains(typeLower, "date") || strings.Contains(typeLower, "time") {
				// For date/time types, check if it's the zero value
				checks = append(checks, fmt.Sprintf("t.%s.IsZero()", fieldName))
			} else {
				// Default to checking against zero value
				checks = append(checks, fmt.Sprintf("t.%s == \"\"", fieldName))
			}
		} else if e.ComplexType != nil {
			// For nested complex types, we need to check all their fields
			// For now, we'll skip this as it would require recursive checking
			// The template will handle this by generating proper types
			continue
		} else {
			// For elements without explicit type, check if it's a named type
			// In WSDL, elements can reference other types by name
			if e.Ref != "" {
				// Reference to another element - skip for now as we don't know its type
				continue
			} else {
				// Default to string check for unknown types
				checks = append(checks, fmt.Sprintf("t.%s == \"\"", fieldName))
			}
		}
	}
	
	// Check attributes (only if not in SimpleContent - already handled above)
	if elem.ComplexType.SimpleContent.Extension.Base == "" {
		for _, attr := range elem.ComplexType.Attributes {
			attrName := utils.MakePublic(attr.Name)
			typeLower := strings.ToLower(attr.Type)
			
			if strings.Contains(typeLower, "string") {
				checks = append(checks, fmt.Sprintf("t.%s == \"\"", attrName))
			} else if strings.Contains(typeLower, "decimal") || strings.Contains(typeLower, "float") || 
				strings.Contains(typeLower, "double") {
				checks = append(checks, fmt.Sprintf("t.%s == 0", attrName))
			} else if strings.Contains(typeLower, "int") || strings.Contains(typeLower, "long") || 
				strings.Contains(typeLower, "short") || strings.Contains(typeLower, "byte") {
				checks = append(checks, fmt.Sprintf("t.%s == 0", attrName))
			} else if strings.Contains(typeLower, "bool") {
				checks = append(checks, fmt.Sprintf("!t.%s", attrName))
			} else {
				// Default to string check
				checks = append(checks, fmt.Sprintf("t.%s == \"\"", attrName))
			}
		}
	}
	
	if len(checks) == 0 {
		return "false"
	}
	
	return strings.Join(checks, " && ")
}