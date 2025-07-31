// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at http://mozilla.org/MPL/2.0/.

//go:build integration
// +build integration

package testing

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestComprehensiveSOAPValidation demonstrates the complete testing framework
func TestComprehensiveSOAPValidation(t *testing.T) {
	// Initialize components
	runner := NewFixtureRunner()
	ruleEngine := NewRuleEngine()
	
	// Load fixtures
	err := runner.LoadFixtures("../../test_fixtures")
	require.NoError(t, err)
	
	fixtures := runner.GetFixtures()
	require.NotEmpty(t, fixtures, "No fixtures loaded")
	
	t.Logf("Loaded %d fixtures", len(fixtures))
	
	// Run fixture tests
	results, err := runner.RunAllTests()
	require.NoError(t, err)
	
	// Validate each result with rules
	for _, result := range results {
		t.Run(fmt.Sprintf("Fixture_%s", result.TestName), func(t *testing.T) {
			// Basic fixture test
			if !result.Passed {
				t.Logf("Fixture test failed: %s", result.Error)
				for _, diff := range result.Differences {
					t.Logf("  - %s: %s", diff.Type, diff.Description)
				}
			}
			
			// Run validation rules
			violations := ruleEngine.ValidateSOAPRequest(result.ActualXML, result.Style)
			
			// Report rule violations
			errorCount := 0
			warningCount := 0
			
			for _, violation := range violations {
				switch violation.Severity {
				case "error":
					errorCount++
					t.Errorf("Rule violation [%s]: %s at %s", 
						violation.RuleName, violation.Message, violation.Path)
					if violation.Suggestion != "" {
						t.Logf("  Suggestion: %s", violation.Suggestion)
					}
				case "warning":
					warningCount++
					t.Logf("Rule warning [%s]: %s at %s", 
						violation.RuleName, violation.Message, violation.Path)
				}
			}
			
			t.Logf("Validation complete: %d errors, %d warnings", errorCount, warningCount)
			
			// Test should pass if no errors (warnings are OK)
			assert.Equal(t, 0, errorCount, "Should have no rule violations")
		})
	}
}

// TestRuleEngineWithKnownIssues tests the rule engine with known problematic XML
func TestRuleEngineWithKnownIssues(t *testing.T) {
	ruleEngine := NewRuleEngine()
	
	testCases := []struct {
		name         string
		soapXML      string
		style        string
		expectErrors int
		expectWarnings int
	}{
		{
			name: "RPC with tns prefix (should fail)",
			soapXML: `<?xml version="1.0" encoding="UTF-8"?>
<soap:Envelope xmlns:soap="http://schemas.xmlsoap.org/soap/envelope/">
  <soap:Body>
    <tns:information xmlns:tns="http://example.com/rpc">
      <userId>123</userId>
    </tns:information>
  </soap:Body>
</soap:Envelope>`,
			style:          "rpc",
			expectErrors:   2, // Should catch both missing namespace declaration and tns: prefix issue
			expectWarnings: 0,
		},
		{
			name: "RPC with correct namespace (should pass)",
			soapXML: `<?xml version="1.0" encoding="UTF-8"?>
<soap:Envelope xmlns:soap="http://schemas.xmlsoap.org/soap/envelope/">
  <soap:Body>
    <information xmlns="http://example.com/rpc">
      <userId>123</userId>
    </information>
  </soap:Body>
</soap:Envelope>`,
			style:          "rpc",
			expectErrors:   0,
			expectWarnings: 0,
		},
		{
			name: "Missing SOAP namespace (should fail)",
			soapXML: `<?xml version="1.0" encoding="UTF-8"?>
<Envelope>
  <Body>
    <GetUser>
      <userId>123</userId>
    </GetUser>
  </Body>
</Envelope>`,
			style:          "document",
			expectErrors:   1, // Missing SOAP namespace
			expectWarnings: 0,
		},
		{
			name: "Empty SOAP body (should fail)",
			soapXML: `<?xml version="1.0" encoding="UTF-8"?>
<soap:Envelope xmlns:soap="http://schemas.xmlsoap.org/soap/envelope/">
  <soap:Body>
  </soap:Body>
</soap:Envelope>`,
			style:          "document",
			expectErrors:   1, // Empty body
			expectWarnings: 0,
		},
	}
	
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			violations := ruleEngine.ValidateSOAPRequest(tc.soapXML, tc.style)
			
			errorCount := 0
			warningCount := 0
			
			for _, violation := range violations {
				switch violation.Severity {
				case "error":
					errorCount++
					t.Logf("Error [%s]: %s", violation.RuleName, violation.Message)
				case "warning":
					warningCount++
					t.Logf("Warning [%s]: %s", violation.RuleName, violation.Message)
				}
			}
			
			assert.Equal(t, tc.expectErrors, errorCount, 
				"Expected %d errors, got %d", tc.expectErrors, errorCount)
			assert.Equal(t, tc.expectWarnings, warningCount, 
				"Expected %d warnings, got %d", tc.expectWarnings, warningCount)
		})
	}
}

// TestFixtureValidationIntegration demonstrates how fixtures and rules work together
func TestFixtureValidationIntegration(t *testing.T) {
	// This test shows how the testing framework would catch the RPC namespace issue
	
	ruleEngine := NewRuleEngine()
	
	// Simulate the problematic XML that was generated before the fix
	problematicXML := `<?xml version="1.0" encoding="UTF-8"?>
<soap:Envelope xmlns:soap="http://schemas.xmlsoap.org/soap/envelope/">
  <soap:Body>
    <tns:information xmlns:tns="https://xbeserve.bedirect.de/soap/2">
      <KID>10320</KID>
      <IDSTRING>meweew3xaj9A</IDSTRING>
      <BIPID>asdasd</BIPID>
    </tns:information>
  </soap:Body>
</soap:Envelope>`
	
	// Expected XML (correct)
	correctXML := `<?xml version="1.0" encoding="UTF-8"?>
<soap:Envelope xmlns:soap="http://schemas.xmlsoap.org/soap/envelope/">
  <soap:Body>
    <information xmlns="https://xbeserve.bedirect.de/soap/2">
      <KID>10320</KID>
      <IDSTRING>meweew3xaj9A</IDSTRING>
      <BIPID>asdasd</BIPID>
    </information>
  </soap:Body>
</soap:Envelope>`
	
	t.Run("Problematic XML should fail rules", func(t *testing.T) {
		violations := ruleEngine.ValidateSOAPRequest(problematicXML, "rpc")
		
		errorCount := 0
		for _, violation := range violations {
			if violation.Severity == "error" {
				errorCount++
				t.Logf("Caught error: %s", violation.Message)
			}
		}
		
		assert.Greater(t, errorCount, 0, "Should catch the tns: prefix issue")
	})
	
	t.Run("Correct XML should pass rules", func(t *testing.T) {
		violations := ruleEngine.ValidateSOAPRequest(correctXML, "rpc")
		
		errorCount := 0
		for _, violation := range violations {
			if violation.Severity == "error" {
				errorCount++
				t.Logf("Unexpected error: %s", violation.Message)
			}
		}
		
		assert.Equal(t, 0, errorCount, "Correct XML should pass all rules")
	})
	
	t.Run("XML comparison should show differences", func(t *testing.T) {
		comparator := NewXMLComparator()
		result, err := comparator.CompareSOAPRequests(correctXML, problematicXML)
		require.NoError(t, err)
		
		assert.False(t, result.Equal, "XMLs should be different")
		assert.NotEmpty(t, result.Differences, "Should detect differences")
	})
}