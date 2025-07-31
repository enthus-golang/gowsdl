// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at http://mozilla.org/MPL/2.0/.

package testing

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestXMLComparatorBasic(t *testing.T) {
	comparator := NewXMLComparator()
	assert.NotNil(t, comparator)
}

func TestXMLComparatorCompare(t *testing.T) {
	comparator := NewXMLComparator()
	
	tests := []struct {
		name         string
		expected     string
		actual       string
		shouldMatch  bool
		expectedDiffs int
	}{
		{
			name:         "Identical XML",
			expected:     `<root><child>value</child></root>`,
			actual:       `<root><child>value</child></root>`,
			shouldMatch:  true,
			expectedDiffs: 0,
		},
		{
			name:         "Different values",
			expected:     `<root><child>value1</child></root>`,
			actual:       `<root><child>value2</child></root>`,
			shouldMatch:  false,
			expectedDiffs: 1,
		},
		{
			name: "Whitespace differences (normalized)",
			expected: `<root><child>value</child></root>`,
			actual: `<root>
				<child>value</child>
			</root>`,
			shouldMatch:  true,
			expectedDiffs: 0,
		},
		{
			name:         "Different structure",
			expected:     `<root><child>value</child></root>`,
			actual:       `<root><other>value</other></root>`,
			shouldMatch:  false,
			expectedDiffs: 1,
		},
		{
			name:         "Missing element",
			expected:     `<root><child1>value</child1><child2>value</child2></root>`,
			actual:       `<root><child1>value</child1></root>`,
			shouldMatch:  false,
			expectedDiffs: 1,
		},
		{
			name:         "Extra element",
			expected:     `<root><child1>value</child1></root>`,
			actual:       `<root><child1>value</child1><child2>value</child2></root>`,
			shouldMatch:  false,
			expectedDiffs: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := comparator.Compare(tt.expected, tt.actual)
			require.NoError(t, err)
			
			assert.Equal(t, tt.shouldMatch, result.Equal, "Match result should be %v", tt.shouldMatch)
			assert.Len(t, result.Differences, tt.expectedDiffs, "Expected %d differences", tt.expectedDiffs)
			
			if !tt.shouldMatch {
				assert.NotEmpty(t, result.Differences, "Should have differences when not matching")
				for _, diff := range result.Differences {
					assert.NotEmpty(t, diff.Type, "Difference type should be set")
					assert.NotEmpty(t, diff.Path, "Difference path should be set")
				}
			}
		})
	}
}

func TestXMLComparatorCompareSOAPRequests(t *testing.T) {
	comparator := NewXMLComparator()
	
	tests := []struct {
		name        string
		expected    string
		actual      string
		shouldMatch bool
	}{
		{
			name: "Identical SOAP requests",
			expected: `<?xml version="1.0"?>
<soap:Envelope xmlns:soap="http://schemas.xmlsoap.org/soap/envelope/">
  <soap:Body>
    <getUser xmlns="http://example.com/test">
      <userId>123</userId>
    </getUser>
  </soap:Body>
</soap:Envelope>`,
			actual: `<?xml version="1.0"?>
<soap:Envelope xmlns:soap="http://schemas.xmlsoap.org/soap/envelope/">
  <soap:Body>
    <getUser xmlns="http://example.com/test">
      <userId>123</userId>
    </getUser>
  </soap:Body>
</soap:Envelope>`,
			shouldMatch: true,
		},
		{
			name: "Different operation names",
			expected: `<?xml version="1.0"?>
<soap:Envelope xmlns:soap="http://schemas.xmlsoap.org/soap/envelope/">
  <soap:Body>
    <getUser xmlns="http://example.com/test">
      <userId>123</userId>
    </getUser>
  </soap:Body>
</soap:Envelope>`,
			actual: `<?xml version="1.0"?>
<soap:Envelope xmlns:soap="http://schemas.xmlsoap.org/soap/envelope/">
  <soap:Body>
    <getCustomer xmlns="http://example.com/test">
      <userId>123</userId>
    </getCustomer>
  </soap:Body>
</soap:Envelope>`,
			shouldMatch: false,
		},
		{
			name: "Different parameter values",
			expected: `<?xml version="1.0"?>
<soap:Envelope xmlns:soap="http://schemas.xmlsoap.org/soap/envelope/">
  <soap:Body>
    <getUser xmlns="http://example.com/test">
      <userId>123</userId>
    </getUser>
  </soap:Body>
</soap:Envelope>`,
			actual: `<?xml version="1.0"?>
<soap:Envelope xmlns:soap="http://schemas.xmlsoap.org/soap/envelope/">
  <soap:Body>
    <getUser xmlns="http://example.com/test">
      <userId>456</userId>
    </getUser>
  </soap:Body>
</soap:Envelope>`,
			shouldMatch: false,
		},
		{
			name: "Different namespaces",
			expected: `<?xml version="1.0"?>
<soap:Envelope xmlns:soap="http://schemas.xmlsoap.org/soap/envelope/">
  <soap:Body>
    <getUser xmlns="http://example.com/test">
      <userId>123</userId>
    </getUser>
  </soap:Body>
</soap:Envelope>`,
			actual: `<?xml version="1.0"?>
<soap:Envelope xmlns:soap="http://schemas.xmlsoap.org/soap/envelope/">
  <soap:Body>
    <getUser xmlns="http://different.example.com/test">
      <userId>123</userId>
    </getUser>
  </soap:Body>
</soap:Envelope>`,
			shouldMatch: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := comparator.CompareSOAPRequests(tt.expected, tt.actual)
			require.NoError(t, err)
			
			assert.Equal(t, tt.shouldMatch, result.Equal, "SOAP requests match result should be %v", tt.shouldMatch)
			
			if !tt.shouldMatch {
				assert.NotEmpty(t, result.Differences, "Should have differences when SOAP requests don't match")
			}
		})
	}
}

func TestXMLComparatorInvalidXML(t *testing.T) {
	comparator := NewXMLComparator()
	
	tests := []struct {
		name     string
		expected string
		actual   string
	}{
		{
			name:     "Invalid expected XML",
			expected: "<invalid xml>",
			actual:   "<root>value</root>",
		},
		{
			name:     "Invalid actual XML",
			expected: "<root>value</root>",
			actual:   "<invalid xml>",
		},
		{
			name:     "Both invalid",
			expected: "<invalid xml>",
			actual:   "<also invalid>",
		},
		{
			name:     "Empty expected",
			expected: "",
			actual:   "<root>value</root>",
		},
		{
			name:     "Empty actual",
			expected: "<root>value</root>",
			actual:   "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := comparator.Compare(tt.expected, tt.actual)
			
			// Should handle invalid XML gracefully
			if err != nil {
				assert.Contains(t, err.Error(), "parse", "Error should mention parsing issue")
			} else {
				// If no error, should at least detect they're different
				assert.False(t, result.Equal, "Invalid XML should not match")
			}
		})
	}
}

func TestXMLComparatorAttributeOrder(t *testing.T) {
	comparator := NewXMLComparator()
	
	// Attributes in different order should still match
	expected := `<root attr1="value1" attr2="value2">content</root>`
	actual := `<root attr2="value2" attr1="value1">content</root>`
	
	result, err := comparator.Compare(expected, actual)
	require.NoError(t, err)
	
	assert.True(t, result.Equal, "XML with attributes in different order should match")
	assert.Empty(t, result.Differences, "Should have no differences")
}

func TestXMLComparatorNamespaceHandling(t *testing.T) {
	comparator := NewXMLComparator()
	
	tests := []struct {
		name        string
		expected    string
		actual      string
		shouldMatch bool
	}{
		{
			name: "Same namespace, different prefixes",
			expected: `<soap:Envelope xmlns:soap="http://schemas.xmlsoap.org/soap/envelope/">
				<soap:Body><test:element xmlns:test="http://example.com/test">value</test:element></soap:Body>
			</soap:Envelope>`,
			actual: `<s:Envelope xmlns:s="http://schemas.xmlsoap.org/soap/envelope/">
				<s:Body><t:element xmlns:t="http://example.com/test">value</t:element></s:Body>
			</s:Envelope>`,
			shouldMatch: true, // Same semantic meaning
		},
		{
			name: "Different namespaces",
			expected: `<root xmlns="http://example.com/ns1">value</root>`,
			actual:   `<root xmlns="http://example.com/ns2">value</root>`,
			shouldMatch: false,
		},
		{
			name: "Namespace vs no namespace",
			expected: `<root>value</root>`,
			actual:   `<root xmlns="http://example.com/ns">value</root>`,
			shouldMatch: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := comparator.Compare(tt.expected, tt.actual)
			require.NoError(t, err)
			
			assert.Equal(t, tt.shouldMatch, result.Equal, "Namespace handling should match expected result")
		})
	}
}

func TestXMLComparatorComplexStructure(t *testing.T) {
	comparator := NewXMLComparator()
	
	expected := `<?xml version="1.0"?>
<soap:Envelope xmlns:soap="http://schemas.xmlsoap.org/soap/envelope/">
  <soap:Header>
    <auth:Authentication xmlns:auth="http://example.com/auth">
      <auth:username>user</auth:username>
      <auth:password>pass</auth:password>
    </auth:Authentication>
  </soap:Header>
  <soap:Body>
    <getUser xmlns="http://example.com/test">
      <userId>123</userId>
      <options>
        <includeDetails>true</includeDetails>
        <format>xml</format>
      </options>
    </getUser>
  </soap:Body>
</soap:Envelope>`

	actual := `<?xml version="1.0"?>
<soap:Envelope xmlns:soap="http://schemas.xmlsoap.org/soap/envelope/">
  <soap:Header>
    <auth:Authentication xmlns:auth="http://example.com/auth">
      <auth:username>user</auth:username>
      <auth:password>pass</auth:password>
    </auth:Authentication>
  </soap:Header>
  <soap:Body>
    <getUser xmlns="http://example.com/test">
      <userId>123</userId>
      <options>
        <includeDetails>true</includeDetails>
        <format>xml</format>
      </options>
    </getUser>
  </soap:Body>
</soap:Envelope>`

	result, err := comparator.Compare(expected, actual)
	require.NoError(t, err)
	
	assert.True(t, result.Equal, "Complex identical structures should match")
	assert.Empty(t, result.Differences, "Should have no differences")
}

func TestXMLComparatorDifferenceTypes(t *testing.T) {
	comparator := NewXMLComparator()
	
	tests := []struct {
		name         string
		expected     string
		actual       string
		expectedType string
	}{
		{
			name:         "Value difference",
			expected:     `<root>value1</root>`,
			actual:       `<root>value2</root>`,
			expectedType: "value",
		},
		{
			name:         "Element name difference",
			expected:     `<root><child1>value</child1></root>`,
			actual:       `<root><child2>value</child2></root>`,
			expectedType: "element",
		},
		{
			name:         "Attribute difference",
			expected:     `<root attr="value1">content</root>`,
			actual:       `<root attr="value2">content</root>`,
			expectedType: "attribute",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := comparator.Compare(tt.expected, tt.actual)
			require.NoError(t, err)
			
			assert.False(t, result.Equal, "Should not match")
			assert.NotEmpty(t, result.Differences, "Should have differences")
			
			// Check that appropriate difference type is detected
			found := false
			for _, diff := range result.Differences {
				if diff.Type == tt.expectedType {
					found = true
					break
				}
			}
			assert.True(t, found, "Should detect difference type: %s", tt.expectedType)
		})
	}
}