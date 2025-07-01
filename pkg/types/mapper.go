// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at http://mozilla.org/MPL/2.0/.

package types

import (
	"strings"
)

// TypeMapper handles conversion from XSD types to Go types
type TypeMapper struct {
	xsdToGoTypes map[string]string
	reservedWords map[string]string
}

// NewTypeMapper creates a new type mapper with default mappings
func NewTypeMapper() *TypeMapper {
	return &TypeMapper{
		xsdToGoTypes: initXSDToGoTypes(),
		reservedWords: initReservedWords(),
	}
}

// MapXSDTypeToGoType converts an XSD type to its Go equivalent
func (tm *TypeMapper) MapXSDTypeToGoType(xsdType string, nillable bool) string {
	// Remove namespace prefix if present
	if idx := strings.LastIndex(xsdType, ":"); idx >= 0 {
		xsdType = xsdType[idx+1:]
	}
	
	goType, ok := tm.xsdToGoTypes[xsdType]
	if !ok {
		// For unknown types, assume they are user-defined complex types
		// and use the type name as-is (without namespace prefix)
		goType = xsdType
	}
	
	// Make pointer if nillable
	if nillable && !strings.HasPrefix(goType, "*") && goType != "interface{}" {
		goType = "*" + goType
	}
	
	return goType
}

// IsReservedWord checks if a word is a Go reserved word
func (tm *TypeMapper) IsReservedWord(word string) bool {
	_, reserved := tm.reservedWords[word]
	return reserved
}

// EscapeReservedWord escapes a reserved word by appending an underscore
func (tm *TypeMapper) EscapeReservedWord(word string) string {
	if escaped, ok := tm.reservedWords[word]; ok {
		return escaped
	}
	return word
}

func initXSDToGoTypes() map[string]string {
	return map[string]string{
		// String types
		"string":            "string",
		"normalizedString":  "string",
		"token":             "string",
		"language":          "string",
		"NMTOKEN":           "string",
		"NMTOKENS":          "string",
		"Name":              "string",
		"NCName":            "string",
		"ID":                "string",
		"IDREF":             "string",
		"IDREFS":            "string",
		"ENTITY":            "string",
		"ENTITIES":          "string",
		"NOTATION":          "string",
		"anyURI":            "string",
		"QName":             "string",
		
		// Numeric types
		"decimal":           "float64",
		"float":             "float32",
		"double":            "float64",
		"integer":           "int64",
		"nonPositiveInteger": "int64",
		"negativeInteger":    "int64",
		"long":              "int64",
		"int":               "int32",
		"short":             "int16",
		"byte":              "int8",
		"nonNegativeInteger": "uint64",
		"unsignedLong":      "uint64",
		"unsignedInt":       "uint32",
		"unsignedShort":     "uint16",
		"unsignedByte":      "uint8",
		"positiveInteger":   "uint64",
		
		// Boolean
		"boolean": "bool",
		
		// Binary
		"base64Binary": "[]byte",
		"hexBinary":    "[]byte",
		
		// Date/Time
		"duration":      "string",
		"dateTime":      "soap.XSDDateTime",
		"time":          "soap.XSDTime",
		"date":          "soap.XSDDate",
		"gYearMonth":    "string",
		"gYear":         "string",
		"gMonthDay":     "string",
		"gDay":          "string",
		"gMonth":        "string",
		
		// Other
		"any":           "interface{}",
		"anyType":       "interface{}",
		"anySimpleType": "interface{}",
	}
}

func initReservedWords() map[string]string {
	return map[string]string{
		"break":       "break_",
		"default":     "default_",
		"func":        "func_",
		"interface":   "interface_",
		"select":      "select_",
		"case":        "case_",
		"defer":       "defer_",
		"go":          "go_",
		"map":         "map_",
		"struct":      "struct_",
		"chan":        "chan_",
		"else":        "else_",
		"goto":        "goto_",
		"package":     "package_",
		"switch":      "switch_",
		"const":       "const_",
		"fallthrough": "fallthrough_",
		"if":          "if_",
		"range":       "range_",
		"type":        "type_",
		"continue":    "continue_",
		"for":         "for_",
		"import":      "import_",
		// "return" is commonly used in RPC-style WSDLs for response parts
		// Map to "result" for more idiomatic Go field names
		"return":      "result",
		"var":         "var_",
		// Common standard library conflicts
		"string":      "string_",
		"error":       "error_",
		"byte":        "byte_",
		"rune":        "rune_",
		"int":         "int_",
		"uint":        "uint_",
		"bool":        "bool_",
		"float32":     "float32_",
		"float64":     "float64_",
		"complex64":   "complex64_",
		"complex128":  "complex128_",
		"uintptr":     "uintptr_",
		"any":         "any_",
	}
}