// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at http://mozilla.org/MPL/2.0/.

package utils

import (
	"regexp"
	"strings"
	"unicode"
)

var (
	regexpNumbersOnly = regexp.MustCompile(`^\d+$`)
	regexpSpecialChars = regexp.MustCompile(`[\s-:/\\,\.\(\)\[\]{}#=<>]`)
)

// MakePublic converts the first letter of a string to uppercase
func MakePublic(id string) string {
	if id == "" {
		return ""
	}
	
	runes := []rune(id)
	for i, r := range runes {
		if unicode.IsLetter(r) || r == '_' {
			if i == 0 || (i > 0 && runes[i-1] != '_') {
				runes[i] = unicode.ToUpper(r)
			}
			break
		}
	}
	return string(runes)
}

// ToGoName converts a string to a valid Go identifier
func ToGoName(s string, public bool) string {
	if s == "" {
		return ""
	}

	// Replace special characters with underscores
	s = regexpSpecialChars.ReplaceAllString(s, "_")
	
	// Remove consecutive underscores
	for strings.Contains(s, "__") {
		s = strings.ReplaceAll(s, "__", "_")
	}
	
	// Trim underscores
	s = strings.Trim(s, "_")
	
	// Handle numeric-only names
	if regexpNumbersOnly.MatchString(s) {
		s = "Field" + s
	}
	
	// Ensure first character is valid
	if s != "" && !unicode.IsLetter(rune(s[0])) && s[0] != '_' {
		s = "Field" + s
	}
	
	// Make public if requested
	if public {
		return MakePublic(s)
	}
	
	return s
}

// NormalizeXMLName normalizes XML element/attribute names for Go
func NormalizeXMLName(name string) string {
	// Remove namespace prefix if present
	if idx := strings.LastIndex(name, ":"); idx >= 0 {
		name = name[idx+1:]
	}
	
	return ToGoName(name, true)
}

// ReservedGoWords contains Go reserved words that need to be escaped
var ReservedGoWords = map[string]string{
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
	"return":      "return_",
	"var":         "var_",
	// Common standard library conflicts
	"string":      "string_",
	"error":       "error_",
	"byte":        "byte_",
	"rune":        "rune_",
}