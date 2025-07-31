// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at http://mozilla.org/MPL/2.0/.

package testing

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestXMLComparatorSemanticComparison(t *testing.T) {
	comparator := NewXMLComparator()

	tests := []struct {
		name         string
		expected     string
		actual       string
		shouldEqual  bool
		expectedDiffs int
		description  string
	}{
		{
			name: "Identical XML",
			expected: `<?xml version="1.0"?>
<root xmlns="http://example.com">
  <child attr="value">content</child>
</root>`,
			actual: `<?xml version="1.0"?>
<root xmlns="http://example.com">
  <child attr="value">content</child>
</root>`,
			shouldEqual:  true,
			expectedDiffs: 0,
			description:  "Identical XML should be equal",
		},
		{
			name: "Different whitespace (ignored by default)",
			expected: `<root><child>content</child></root>`,
			actual: `<root>
  <child>content</child>
</root>`,
			shouldEqual:  true,
			expectedDiffs: 0,
			description:  "Whitespace differences should be ignored",
		},
		{
			name: "Different element names",
			expected: `<root><child>content</child></root>`,
			actual:   `<root><other>content</other></root>`,
			shouldEqual:  false,
			expectedDiffs: 1,
			description:  "Different element names should be detected",
		},
		{
			name: "Different text content",
			expected: `<root><child>expected</child></root>`,
			actual:   `<root><child>actual</child></root>`,
			shouldEqual:  false,
			expectedDiffs: 1,
			description:  "Different text content should be detected",
		},
		{
			name: "Missing attribute",
			expected: `<root><child attr="value">content</child></root>`,
			actual:   `<root><child>content</child></root>`,
			shouldEqual:  false,
			expectedDiffs: 1,
			description:  "Missing attributes should be detected",
		},
		{
			name: "Extra attribute",
			expected: `<root><child>content</child></root>`,
			actual:   `<root><child extra="value">content</child></root>`,
			shouldEqual:  false,
			expectedDiffs: 1,
			description:  "Extra attributes should be detected",
		},
		{
			name: "Different attribute values",
			expected: `<root><child attr="expected">content</child></root>`,
			actual:   `<root><child attr="actual">content</child></root>`,
			shouldEqual:  false,
			expectedDiffs: 1,
			description:  "Different attribute values should be detected",
		},
		{
			name: "Missing child element",
			expected: `<root><child1/><child2/></root>`,
			actual:   `<root><child1/></root>`,
			shouldEqual:  false,
			expectedDiffs: 1,
			description:  "Missing child elements should be detected",
		},
		{
			name: "Extra child element",
			expected: `<root><child1/></root>`,
			actual:   `<root><child1/><child2/></root>`,
			shouldEqual:  false,
			expectedDiffs: 1,
			description:  "Extra child elements should be detected",
		},
		{
			name: "Namespace differences",
			expected: `<root xmlns="http://example.com"><child>content</child></root>`,
			actual:   `<root xmlns="http://other.com"><child>content</child></root>`,
			shouldEqual:  false,
			expectedDiffs: 3, // Root element namespace, xmlns attribute, and child element namespace
			description:  "Namespace differences should be detected",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := comparator.Compare(tt.expected, tt.actual)
			require.NoError(t, err, "Comparison should not error")
			
			assert.Equal(t, tt.shouldEqual, result.Equal, tt.description)
			assert.Len(t, result.Differences, tt.expectedDiffs, "Expected %d differences, got %d", tt.expectedDiffs, len(result.Differences))
			
			// Log differences for debugging
			if len(result.Differences) > 0 {
				t.Logf("Differences found:")
				for i, diff := range result.Differences {
					t.Logf("  %d. Type: %s, Path: %s, Expected: %q, Actual: %q, Description: %s", 
						i+1, diff.Type, diff.Path, diff.Expected, diff.Actual, diff.Description)
				}
			}
		})
	}
}

func TestXMLComparatorWithIgnoreOptions(t *testing.T) {
	tests := []struct {
		name                 string
		ignoreWhitespace     bool
		ignoreOrder          bool
		ignoreNamespaces     bool
		expected             string
		actual               string
		shouldEqual          bool
		description          string
	}{
		{
			name:             "Ignore namespaces",
			ignoreNamespaces: true,
			expected:         `<root xmlns="http://example.com"><child>content</child></root>`,
			actual:           `<root xmlns="http://other.com"><child>content</child></root>`,
			shouldEqual:      true,
			description:      "Should ignore namespace differences when configured",
		},
		{
			name:        "Ignore element order",
			ignoreOrder: true,
			expected:    `<root><child1/><child2/></root>`,
			actual:      `<root><child2/><child1/></root>`,
			shouldEqual: true,
			description: "Should ignore element order when configured",
		},
		{
			name:             "Don't ignore whitespace",
			ignoreWhitespace: false,
			expected:         `<root><child>content</child></root>`,
			actual: `<root>
  <child>content</child>
</root>`,
			shouldEqual: false,
			description: "Should detect whitespace differences when not ignored",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			comparator := &XMLComparator{
				IgnoreWhitespace: tt.ignoreWhitespace,
				IgnoreOrder:      tt.ignoreOrder,
				IgnoreNamespaces: tt.ignoreNamespaces,
			}
			
			result, err := comparator.Compare(tt.expected, tt.actual)
			require.NoError(t, err, "Comparison should not error")
			
			assert.Equal(t, tt.shouldEqual, result.Equal, tt.description)
		})
	}
}

func TestXMLComparatorSOAPComparison(t *testing.T) {
	comparator := NewXMLComparator()

	tests := []struct {
		name         string
		expected     string
		actual       string
		shouldEqual  bool
		expectedDiffs int
		description  string
	}{
		{
			name: "Identical SOAP requests",
			expected: `<?xml version="1.0"?>
<soap:Envelope xmlns:soap="http://schemas.xmlsoap.org/soap/envelope/">
  <soap:Body>
    <getUser xmlns="http://example.com/service">
      <userId>123</userId>
    </getUser>
  </soap:Body>
</soap:Envelope>`,
			actual: `<?xml version="1.0"?>
<soap:Envelope xmlns:soap="http://schemas.xmlsoap.org/soap/envelope/">
  <soap:Body>
    <getUser xmlns="http://example.com/service">
      <userId>123</userId>
    </getUser>
  </soap:Body>
</soap:Envelope>`,
			shouldEqual:  true,
			expectedDiffs: 0,
			description:  "Identical SOAP requests should be equal",
		},
		{
			name: "Different operation names",
			expected: `<?xml version="1.0"?>
<soap:Envelope xmlns:soap="http://schemas.xmlsoap.org/soap/envelope/">
  <soap:Body>
    <getUser xmlns="http://example.com/service">
      <userId>123</userId>
    </getUser>
  </soap:Body>
</soap:Envelope>`,
			actual: `<?xml version="1.0"?>
<soap:Envelope xmlns:soap="http://schemas.xmlsoap.org/soap/envelope/">
  <soap:Body>
    <getUserInfo xmlns="http://example.com/service">
      <userId>123</userId>
    </getUserInfo>
  </soap:Body>
</soap:Envelope>`,
			shouldEqual:  false,
			expectedDiffs: 1,
			description:  "Different operation names should be detected",
		},
		{
			name: "Different parameter values",
			expected: `<?xml version="1.0"?>
<soap:Envelope xmlns:soap="http://schemas.xmlsoap.org/soap/envelope/">
  <soap:Body>
    <getUser xmlns="http://example.com/service">
      <userId>123</userId>
    </getUser>
  </soap:Body>
</soap:Envelope>`,
			actual: `<?xml version="1.0"?>
<soap:Envelope xmlns:soap="http://schemas.xmlsoap.org/soap/envelope/">
  <soap:Body>
    <getUser xmlns="http://example.com/service">
      <userId>456</userId>
    </getUser>
  </soap:Body>
</soap:Envelope>`,
			shouldEqual:  false,
			expectedDiffs: 1,
			description:  "Different parameter values should be detected",
		},
		{
			name: "RPC vs Document style namespace handling",
			expected: `<?xml version="1.0"?>
<soap:Envelope xmlns:soap="http://schemas.xmlsoap.org/soap/envelope/">
  <soap:Body>
    <information xmlns="http://example.com/service">
      <userId>123</userId>
    </information>
  </soap:Body>
</soap:Envelope>`,
			actual: `<?xml version="1.0"?>
<soap:Envelope xmlns:soap="http://schemas.xmlsoap.org/soap/envelope/">
  <soap:Body>
    <tns:information xmlns:tns="http://example.com/service">
      <userId>123</userId>
    </tns:information>
  </soap:Body>
</soap:Envelope>`,
			shouldEqual:  false,
			expectedDiffs: 1,
			description:  "Should detect the RPC namespace issue (tns: vs xmlns)",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := comparator.CompareSOAPRequests(tt.expected, tt.actual)
			require.NoError(t, err, "SOAP comparison should not error")
			
			assert.Equal(t, tt.shouldEqual, result.Equal, tt.description)
			assert.Len(t, result.Differences, tt.expectedDiffs, "Expected %d differences, got %d", tt.expectedDiffs, len(result.Differences))
			
			// Log differences for debugging
			if len(result.Differences) > 0 {
				t.Logf("SOAP Differences found:")
				for i, diff := range result.Differences {
					t.Logf("  %d. Type: %s, Path: %s, Expected: %q, Actual: %q, Description: %s", 
						i+1, diff.Type, diff.Path, diff.Expected, diff.Actual, diff.Description)
				}
			}
		})
	}
}

func TestXMLComparatorComplexStructures(t *testing.T) {
	comparator := NewXMLComparator()

	expected := `<?xml version="1.0"?>
<soap:Envelope xmlns:soap="http://schemas.xmlsoap.org/soap/envelope/">
  <soap:Body>
    <getUserDetails xmlns="http://example.com/service">
      <userRequest>
        <userId type="integer">123</userId>
        <includeAddress>true</includeAddress>
        <includeHistory>false</includeHistory>
        <options>
          <format>json</format>
          <locale>en-US</locale>
        </options>
      </userRequest>
    </getUserDetails>
  </soap:Body>
</soap:Envelope>`

	actual := `<?xml version="1.0"?>
<soap:Envelope xmlns:soap="http://schemas.xmlsoap.org/soap/envelope/">
  <soap:Body>
    <getUserDetails xmlns="http://example.com/service">
      <userRequest>
        <userId type="integer">456</userId>
        <includeAddress>true</includeAddress>
        <includeHistory>false</includeHistory>
        <options>
          <format>xml</format>
          <locale>en-US</locale>
        </options>
      </userRequest>
    </getUserDetails>
  </soap:Body>
</soap:Envelope>`

	result, err := comparator.CompareSOAPRequests(expected, actual)
	require.NoError(t, err)
	
	assert.False(t, result.Equal, "Should detect differences in complex structure")
	assert.Equal(t, 2, len(result.Differences), "Should find 2 differences (userId and format)")
	
	// Verify specific differences
	var userIdDiff, formatDiff *Difference
	for _, diff := range result.Differences {
		if strings.Contains(diff.Path, "userId") {
			userIdDiff = &diff
		} else if strings.Contains(diff.Path, "format") {
			formatDiff = &diff
		}
	}
	
	require.NotNil(t, userIdDiff, "Should detect userId difference")
	assert.Equal(t, "123", userIdDiff.Expected)
	assert.Equal(t, "456", userIdDiff.Actual)
	
	require.NotNil(t, formatDiff, "Should detect format difference")
	assert.Equal(t, "json", formatDiff.Expected)
	assert.Equal(t, "xml", formatDiff.Actual)
}

func TestXMLComparatorErrorHandling(t *testing.T) {
	comparator := NewXMLComparator()

	tests := []struct {
		name        string
		expected    string
		actual      string
		expectError bool
		description string
	}{
		{
			name:        "Invalid expected XML",
			expected:    "<invalid xml>",
			actual:      "<root>valid</root>",
			expectError: true,
			description: "Should error on invalid expected XML",
		},
		{
			name:        "Invalid actual XML",
			expected:    "<root>valid</root>",
			actual:      "<invalid xml>",
			expectError: true,
			description: "Should error on invalid actual XML",
		},
		{
			name:        "Empty XML strings",
			expected:    "",
			actual:      "",
			expectError: true,
			description: "Should error on empty XML strings",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := comparator.Compare(tt.expected, tt.actual)
			
			if tt.expectError {
				assert.Error(t, err, tt.description)
				assert.Nil(t, result, "Result should be nil on error")
			} else {
				assert.NoError(t, err, tt.description)
				assert.NotNil(t, result, "Result should not be nil on success")
			}
		})
	}
}