// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at http://mozilla.org/MPL/2.0/.

package utils

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMakePublic(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "lowercase_to_uppercase",
			input:    "testName",
			expected: "TestName",
		},
		{
			name:     "already_uppercase",
			input:    "TestName",
			expected: "TestName",
		},
		{
			name:     "single_character",
			input:    "a",
			expected: "A",
		},
		{
			name:     "empty_string",
			input:    "",
			expected: "",
		},
		{
			name:     "underscore_start",
			input:    "_private",
			expected: "_private", // Underscore names stay private
		},
		{
			name:     "numbers_start",
			input:    "123test",
			expected: "123Test", // Numbers at start with capitalization
		},
		{
			name:     "mixed_case",
			input:    "xmlParser",
			expected: "XmlParser",
		},
		{
			name:     "all_caps",
			input:    "XML",
			expected: "XML",
		},
		{
			name:     "camelCase",
			input:    "getElementById",
			expected: "GetElementById",
		},
		{
			name:     "with_numbers",
			input:    "element2Name",
			expected: "Element2Name",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := MakePublic(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestMakePublicSpecialCases(t *testing.T) {
	// Test some real-world scenarios
	tests := []struct {
		input    string
		expected string
	}{
		// Common WSDL/XSD names
		{"elementName", "ElementName"},
		{"complexType", "ComplexType"},
		{"simpleType", "SimpleType"},
		{"targetNamespace", "TargetNamespace"},
		
		// Common programming terms
		{"httpClient", "HttpClient"},
		{"xmlDocument", "XmlDocument"},
		{"jsonParser", "JsonParser"},
		{"soapEnvelope", "SoapEnvelope"},
		
		// Edge cases from WSDL generation
		{"id", "Id"},
		{"url", "Url"},
		{"uri", "Uri"},
		{"uuid", "Uuid"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := MakePublic(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestMakePublicWithSpecialCharacters(t *testing.T) {
	// Test handling of special characters
	tests := []struct {
		input    string
		expected string
	}{
		{"test-name", "Test-name"},    // Hyphen preserved
		{"test_name", "Test_name"},    // Underscore preserved
		{"test.name", "Test.name"},    // Dot preserved
		{"test name", "Test name"},    // Space preserved
		{"test123", "Test123"},        // Numbers preserved
		{"123test", "123Test"},        // Leading numbers with capitalization
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := MakePublic(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestMakePublicUnicode(t *testing.T) {
	// Test Unicode handling
	tests := []struct {
		input    string
		expected string
	}{
		{"élément", "Élément"},       // French accented character
		{"测试", "测试"},                // Chinese characters (no change expected)
		{"тест", "Тест"},              // Cyrillic characters with capitalization
		{"münchen", "München"},       // German umlaut
		{"naïve", "Naïve"},          // Character with diaeresis
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := MakePublic(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestMakePublicPerformance(t *testing.T) {
	// Test with a longer string to ensure performance is reasonable
	longString := "thisIsAVeryLongStringWithManyWordsToTestThePerformanceOfTheMakePublicFunction"
	expected := "ThisIsAVeryLongStringWithManyWordsToTestThePerformanceOfTheMakePublicFunction"
	
	result := MakePublic(longString)
	assert.Equal(t, expected, result)
}

func BenchmarkMakePublic(b *testing.B) {
	testCases := []string{
		"test",
		"testName",
		"veryLongTestNameWithManyWords",
		"element",
		"xmlParser",
	}

	for _, tc := range testCases {
		b.Run(tc, func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				_ = MakePublic(tc)
			}
		})
	}
}