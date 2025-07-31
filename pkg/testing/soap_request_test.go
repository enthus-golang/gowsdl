// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at http://mozilla.org/MPL/2.0/.

package testing

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestSOAPRequestFixtures tests SOAP request generation against fixtures
func TestSOAPRequestFixtures(t *testing.T) {
	runner := NewFixtureRunner()
	
	// Load fixtures from test_fixtures directory
	err := runner.LoadFixtures("../../test_fixtures")
	require.NoError(t, err)
	
	fixtures := runner.GetFixtures()
	require.NotEmpty(t, fixtures, "No fixtures loaded")
	
	// Run all fixture tests
	results, err := runner.RunAllTests()
	require.NoError(t, err)
	
	// Report results
	for _, result := range results {
		t.Run(result.TestName, func(t *testing.T) {
			if !result.Passed {
				t.Logf("Test failed for %s (%s style)", result.TestName, result.Style)
				t.Logf("Error: %s", result.Error)
				t.Logf("Expected XML:\n%s", result.ExpectedXML)
				t.Logf("Actual XML:\n%s", result.ActualXML)
				
				for _, diff := range result.Differences {
					t.Logf("Difference: %s at %s - Expected: %s, Actual: %s", 
						diff.Type, diff.Path, diff.Expected, diff.Actual)
				}
			}
			
			assert.True(t, result.Passed, "SOAP request should match expected XML")
		})
	}
}

// TestSOAPCapture tests the SOAP capture functionality
func TestSOAPCapture(t *testing.T) {
	capture := NewSOAPCapture()
	
	// Test that capture starts empty
	assert.Equal(t, 0, capture.GetRequestCount())
	assert.Nil(t, capture.GetLastRequest())
	
	// TODO: Add tests for actual HTTP interception
	// This would require setting up mock HTTP clients
}

// TestXMLComparator tests the XML comparison functionality
func TestXMLComparator(t *testing.T) {
	comparator := NewXMLComparator()
	
	tests := []struct {
		name      string
		expected  string
		actual    string
		shouldMatch bool
	}{
		{
			name: "identical XML",
			expected: `<?xml version="1.0"?><root><child>value</child></root>`,
			actual:   `<?xml version="1.0"?><root><child>value</child></root>`,
			shouldMatch: true,
		},
		{
			name: "different values",
			expected: `<?xml version="1.0"?><root><child>value1</child></root>`,
			actual:   `<?xml version="1.0"?><root><child>value2</child></root>`,
			shouldMatch: false,
		},
		{
			name: "whitespace differences (should ignore)",
			expected: `<?xml version="1.0"?><root><child>value</child></root>`,
			actual:   `<?xml version="1.0"?>
			<root>
				<child>value</child>
			</root>`,
			shouldMatch: true,
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := comparator.Compare(tt.expected, tt.actual)
			require.NoError(t, err)
			
			if tt.shouldMatch {
				assert.True(t, result.Equal, "XML should match")
				assert.Empty(t, result.Differences, "Should have no differences")
			} else {
				assert.False(t, result.Equal, "XML should not match")
				assert.NotEmpty(t, result.Differences, "Should have differences")
			}
		})
	}
}

// TestExtractOperationName tests operation name extraction
func TestExtractOperationName(t *testing.T) {
	tests := []struct {
		name         string
		soapXML      string
		expectedOp   string
		shouldError  bool
	}{
		{
			name: "RPC style operation",
			soapXML: `<?xml version="1.0"?>
			<soap:Envelope xmlns:soap="http://schemas.xmlsoap.org/soap/envelope/">
				<soap:Body>
					<getUser xmlns="http://example.com/rpc">
						<userId>123</userId>
					</getUser>
				</soap:Body>
			</soap:Envelope>`,
			expectedOp: "getUser",
			shouldError: false,
		},
		{
			name: "Document style operation",
			soapXML: `<?xml version="1.0"?>
			<soap:Envelope xmlns:soap="http://schemas.xmlsoap.org/soap/envelope/">
				<soap:Body>
					<GetUserRequest xmlns="http://example.com/simple">
						<userId>123</userId>
					</GetUserRequest>
				</soap:Body>
			</soap:Envelope>`,
			expectedOp: "GetUserRequest",
			shouldError: false,
		},
		{
			name: "Operation with namespace prefix",
			soapXML: `<?xml version="1.0"?>
			<soap:Envelope xmlns:soap="http://schemas.xmlsoap.org/soap/envelope/">
				<soap:Body>
					<tns:information xmlns:tns="http://example.com/test">
						<userId>123</userId>
					</tns:information>
				</soap:Body>
			</soap:Envelope>`,
			expectedOp: "information",
			shouldError: false,
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			op, err := ExtractOperationName(tt.soapXML)
			
			if tt.shouldError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedOp, op)
			}
		})
	}
}