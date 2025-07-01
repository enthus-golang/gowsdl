// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at http://mozilla.org/MPL/2.0/.

package types

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewTypeMapper(t *testing.T) {
	tm := NewTypeMapper()
	assert.NotNil(t, tm)
	assert.NotNil(t, tm.xsdToGoTypes)
	assert.NotNil(t, tm.reservedWords)
}

func TestMapXSDTypeToGoType(t *testing.T) {
	tm := NewTypeMapper()

	tests := []struct {
		name     string
		xsdType  string
		nillable bool
		expected string
	}{
		// String types
		{"string", "string", false, "string"},
		{"normalizedString", "normalizedString", false, "string"},
		{"token", "token", false, "string"},
		
		// Numeric types
		{"decimal", "decimal", false, "float64"},
		{"float", "float", false, "float32"},
		{"double", "double", false, "float64"},
		{"integer", "integer", false, "int64"},
		{"int", "int", false, "int32"},
		{"short", "short", false, "int16"},
		{"byte", "byte", false, "int8"},
		{"long", "long", false, "int64"},
		
		// Boolean
		{"boolean", "boolean", false, "bool"},
		
		// Binary
		{"base64Binary", "base64Binary", false, "[]byte"},
		{"hexBinary", "hexBinary", false, "[]byte"},
		
		// Date/Time
		{"dateTime", "dateTime", false, "soap.XSDDateTime"},
		{"date", "date", false, "soap.XSDDate"},
		{"time", "time", false, "soap.XSDTime"},
		
		// Other
		{"any", "any", false, "interface{}"},
		{"anyType", "anyType", false, "interface{}"},
		
		// With namespace prefix
		{"xsd:string", "xsd:string", false, "string"},
		{"ns:string", "ns:string", false, "string"},
		
		// Unknown types (user-defined)
		{"CustomType", "CustomType", false, "CustomType"},
		{"MyComplexType", "MyComplexType", false, "MyComplexType"},
		
		// Nillable types
		{"string", "string", true, "*string"},
		{"int", "int", true, "*int32"},
		{"CustomType", "CustomType", true, "*CustomType"},
		
		// Interface types should not be pointers even when nillable
		{"any", "any", true, "interface{}"},
		{"anyType", "anyType", true, "interface{}"},
	}

	for _, tt := range tests {
		nillableStr := "false"
		if tt.nillable {
			nillableStr = "true"
		}
		t.Run(tt.name+"_nillable_"+nillableStr, func(t *testing.T) {
			result := tm.MapXSDTypeToGoType(tt.xsdType, tt.nillable)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestIsReservedWord(t *testing.T) {
	tm := NewTypeMapper()

	tests := []struct {
		word     string
		expected bool
	}{
		// Go keywords
		{"break", true},
		{"case", true},
		{"chan", true},
		{"const", true},
		{"continue", true},
		{"default", true},
		{"defer", true},
		{"else", true},
		{"fallthrough", true},
		{"for", true},
		{"func", true},
		{"go", true},
		{"goto", true},
		{"if", true},
		{"import", true},
		{"interface", true},
		{"map", true},
		{"package", true},
		{"range", true},
		{"return", true},
		{"select", true},
		{"struct", true},
		{"switch", true},
		{"type", true},
		{"var", true},
		
		// Built-in types
		{"string", true},
		{"int", true},
		{"float32", true},
		{"bool", true},
		{"byte", true},
		{"rune", true},
		{"error", true},
		{"any", true},
		
		// Non-reserved words
		{"hello", false},
		{"CustomType", false},
		{"MyStruct", false},
		{"", false},
	}

	for _, tt := range tests {
		t.Run(tt.word, func(t *testing.T) {
			result := tm.IsReservedWord(tt.word)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestEscapeReservedWord(t *testing.T) {
	tm := NewTypeMapper()

	tests := []struct {
		word     string
		expected string
	}{
		// Reserved words get escaped
		{"break", "break_"},
		{"case", "case_"},
		{"string", "string_"},
		{"int", "int_"},
		{"interface", "interface_"},
		
		// Special case: "return" maps to "result" for RPC-style WSDLs
		{"return", "result"},
		
		// Non-reserved words stay the same
		{"hello", "hello"},
		{"CustomType", "CustomType"},
		{"MyStruct", "MyStruct"},
		{"", ""},
	}

	for _, tt := range tests {
		t.Run(tt.word, func(t *testing.T) {
			result := tm.EscapeReservedWord(tt.word)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestInitXSDToGoTypes(t *testing.T) {
	mapping := initXSDToGoTypes()
	
	// Test some key mappings
	assert.Equal(t, "string", mapping["string"])
	assert.Equal(t, "int32", mapping["int"])
	assert.Equal(t, "float64", mapping["double"])
	assert.Equal(t, "bool", mapping["boolean"])
	assert.Equal(t, "[]byte", mapping["base64Binary"])
	assert.Equal(t, "soap.XSDDateTime", mapping["dateTime"])
	assert.Equal(t, "interface{}", mapping["any"])
	
	// Ensure we have a reasonable number of mappings
	assert.Greater(t, len(mapping), 20)
}

func TestInitReservedWords(t *testing.T) {
	reserved := initReservedWords()
	
	// Test some key reserved words
	assert.Equal(t, "break_", reserved["break"])
	assert.Equal(t, "string_", reserved["string"])
	assert.Equal(t, "interface_", reserved["interface"])
	assert.Equal(t, "type_", reserved["type"])
	
	// Ensure we have a reasonable number of reserved words
	assert.Greater(t, len(reserved), 20)
}

func TestTypeMapperEdgeCases(t *testing.T) {
	tm := NewTypeMapper()

	t.Run("empty_string_type", func(t *testing.T) {
		result := tm.MapXSDTypeToGoType("", false)
		assert.Equal(t, "", result)
	})

	t.Run("complex_namespace_prefix", func(t *testing.T) {
		result := tm.MapXSDTypeToGoType("ns1:ns2:localName", false)
		assert.Equal(t, "localName", result)
	})

	t.Run("already_pointer_type", func(t *testing.T) {
		// The function checks if the type already has a * prefix to avoid double pointers
		result := tm.MapXSDTypeToGoType("*CustomType", true)
		assert.Equal(t, "*CustomType", result) // Doesn't add another * if already present
	})
}