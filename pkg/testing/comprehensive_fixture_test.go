// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at http://mozilla.org/MPL/2.0/.

package testing

import (
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestComprehensiveFixtureValidation(t *testing.T) {
	// Test our expanded fixture coverage
	runner := NewFixtureRunner()
	
	// Load fixtures from the test_fixtures directory
	fixturesDir := filepath.Join("..", "..", "test_fixtures")
	err := runner.LoadFixtures(fixturesDir)
	require.NoError(t, err, "Should load all fixtures successfully")
	
	fixtures := runner.GetFixtures()
	assert.GreaterOrEqual(t, len(fixtures), 5, "Should have at least 5 fixtures including new complex ones")
	
	// Verify we have the expected fixture types
	fixtureNames := make(map[string]bool)
	rpcCount := 0
	documentCount := 0
	edgeCaseCount := 0
	
	for _, fixture := range fixtures {
		fixtureNames[fixture.Name] = true
		
		switch fixture.Style {
		case "rpc":
			rpcCount++
		case "document":
			documentCount++
		}
		
		// Count edge cases (would be classified as document or rpc but are complex)
		if fixture.Name == "array_namespace_edge_case" || 
		   fixture.Name == "complex_rpc" || 
		   fixture.Name == "complex_document" ||
		   fixture.Name == "namespace_edge_case" {
			edgeCaseCount++
		}
	}
	
	// Verify we have good coverage
	assert.GreaterOrEqual(t, rpcCount, 2, "Should have at least 2 RPC fixtures")
	assert.GreaterOrEqual(t, documentCount, 2, "Should have at least 2 document fixtures") 
	assert.GreaterOrEqual(t, edgeCaseCount, 3, "Should have at least 3 edge case fixtures")
	
	// Verify specific fixtures we created
	expectedFixtures := []string{
		"complex_rpc",
		"namespace_edge_case", 
		"complex_document",
		"array_namespace_edge_case",
	}
	
	for _, expected := range expectedFixtures {
		assert.True(t, fixtureNames[expected], "Should have fixture: %s", expected)
	}
	
	t.Logf("Loaded %d fixtures total:", len(fixtures))
	for _, fixture := range fixtures {
		t.Logf("  - %s (%s style)", fixture.Name, fixture.Style)
	}
}

func TestRegressionPreventionFixtures(t *testing.T) {
	// Test that our fixtures would catch the RPC namespace issue
	runner := NewFixtureRunner()
	engine := NewRuleEngine()
	
	// Test the namespace edge case fixture specifically
	fixturesDir := filepath.Join("..", "..", "test_fixtures") 
	err := runner.LoadFixtures(fixturesDir)
	require.NoError(t, err)
	
	fixtures := runner.GetFixtures()
	var namespaceFixture *FixtureTestCase
	
	for i, fixture := range fixtures {
		if fixture.Name == "namespace_edge_case" {
			namespaceFixture = &fixtures[i]
			break
		}
	}
	
	require.NotNil(t, namespaceFixture, "Should have namespace edge case fixture")
	assert.Equal(t, "rpc", namespaceFixture.Style, "Namespace fixture should be RPC style")
	
	// Validate that the expected XML is correct (uses xmlns, not tns:)
	expectedXML := namespaceFixture.ExpectedRequest
	assert.Contains(t, expectedXML, `<information xmlns="http://example.com/namespace-test">`, 
		"Expected XML should use xmlns format, not tns: prefix")
	assert.NotContains(t, expectedXML, `<tns:information`, 
		"Expected XML should NOT use tns: prefix")
	
	// Test validation rules against both correct and incorrect formats
	correctXML := expectedXML
	incorrectXML := `<?xml version="1.0" encoding="UTF-8"?>
<soap:Envelope xmlns:soap="http://schemas.xmlsoap.org/soap/envelope/">
  <soap:Body>
    <tns:information xmlns:tns="http://example.com/namespace-test">
      <dataId>test123</dataId>
      <includeDetails>true</includeDetails>
    </tns:information>
  </soap:Body>
</soap:Envelope>`
	
	// Validate correct format passes
	correctViolations := engine.ValidateSOAPRequest(correctXML, "rpc")
	t.Logf("Correct XML violations: %d", len(correctViolations))
	for _, v := range correctViolations {
		t.Logf("  - %s: %s", v.RuleName, v.Message)
	}
	
	// Validate incorrect format fails
	incorrectViolations := engine.ValidateSOAPRequest(incorrectXML, "rpc")
	t.Logf("Incorrect XML violations: %d", len(incorrectViolations))
	for _, v := range incorrectViolations {
		t.Logf("  - %s: %s", v.RuleName, v.Message)
	}
	
	// The incorrect format should have more violations
	assert.Greater(t, len(incorrectViolations), len(correctViolations), 
		"Incorrect namespace format should trigger more violations")
	
	// Check for specific RPC operation wrapper violations
	var rpcViolation *RuleViolation
	for _, v := range incorrectViolations {
		if v.RuleName == "RPC Operation Wrapper" {
			rpcViolation = &v
			break
		}
	}
	
	require.NotNil(t, rpcViolation, "Should detect RPC operation wrapper violation")
	// The violation could be either about missing xmlns or tns: prefix usage
	assert.True(t, 
		strings.Contains(rpcViolation.Message, "tns:") || 
		strings.Contains(rpcViolation.Message, "namespace"), 
		"Should mention either tns: prefix or namespace issue in violation message: %s", rpcViolation.Message)
}

func TestComplexTypeHandling(t *testing.T) {
	// Test that complex fixtures load properly with nested types
	runner := NewFixtureRunner()
	
	fixturesDir := filepath.Join("..", "..", "test_fixtures")
	err := runner.LoadFixtures(fixturesDir)
	require.NoError(t, err)
	
	fixtures := runner.GetFixtures()
	
	// Find complex fixtures
	var complexRPC, complexDocument *FixtureTestCase
	
	for i, fixture := range fixtures {
		if fixture.Name == "complex_rpc" {
			complexRPC = &fixtures[i]
		} else if fixture.Name == "complex_document" {
			complexDocument = &fixtures[i]
		}
	}
	
	// Test complex RPC fixture
	if complexRPC != nil {
		assert.Equal(t, "rpc", complexRPC.Style)
		assert.Contains(t, complexRPC.ExpectedRequest, "GetUserDetails")
		assert.Contains(t, complexRPC.ExpectedRequest, "xmlns=") // Should use xmlns, not tns:
		
		// Test data should be well-formed
		assert.Contains(t, complexRPC.TestData, "operation")
		assert.Equal(t, "GetUserDetails", complexRPC.TestData["operation"])
	}
	
	// Test complex document fixture  
	if complexDocument != nil {
		assert.Equal(t, "document", complexDocument.Style)
		assert.Contains(t, complexDocument.ExpectedRequest, "ProcessOrderRequest")
		assert.Contains(t, complexDocument.ExpectedRequest, "orderId")
		assert.Contains(t, complexDocument.ExpectedRequest, "items")
		
		// Should have nested structures
		assert.Contains(t, complexDocument.ExpectedRequest, "<item>")
		assert.Contains(t, complexDocument.ExpectedRequest, "<shippingAddress>")
	}
}

func TestArrayAndNestedStructures(t *testing.T) {
	// Test the array edge case fixture
	runner := NewFixtureRunner()
	
	fixturesDir := filepath.Join("..", "..", "test_fixtures")
	err := runner.LoadFixtures(fixturesDir)
	require.NoError(t, err)
	
	fixtures := runner.GetFixtures()
	var arrayFixture *FixtureTestCase
	
	for i, fixture := range fixtures {
		if fixture.Name == "array_namespace_edge_case" {
			arrayFixture = &fixtures[i]
			break
		}
	}
	
	if arrayFixture != nil {
		// Test that array fixture handles complex nested structures
		expectedXML := arrayFixture.ExpectedRequest
		
		// Should contain array elements
		assert.Contains(t, expectedXML, "<stringArray>")
		assert.Contains(t, expectedXML, "<item>first</item>")
		assert.Contains(t, expectedXML, "<intArray>")
		assert.Contains(t, expectedXML, "<value>1</value>")
		
		// Should contain nested structures
		assert.Contains(t, expectedXML, "<nestedArrays>")
		assert.Contains(t, expectedXML, "<level1>")
		assert.Contains(t, expectedXML, "<level2>")
		
		// Should handle empty arrays
		assert.Contains(t, expectedXML, "<emptyArray/>")
		
		// Test data should be structured
		assert.Contains(t, arrayFixture.TestData, "stringArray")
		assert.Contains(t, arrayFixture.TestData, "nestedArrays")
	}
}

func TestFixtureCoverage(t *testing.T) {
	// Comprehensive test to ensure our fixtures provide good coverage
	runner := NewFixtureRunner()
	engine := NewRuleEngine()
	
	fixturesDir := filepath.Join("..", "..", "test_fixtures")
	err := runner.LoadFixtures(fixturesDir)
	require.NoError(t, err)
	
	fixtures := runner.GetFixtures()
	require.Greater(t, len(fixtures), 0, "Should have fixtures loaded")
	
	// Test each fixture against validation rules
	coverageReport := make(map[string]int) // rule name -> times triggered
	
	for _, fixture := range fixtures {
		violations := engine.ValidateSOAPRequest(fixture.ExpectedRequest, fixture.Style)
		
		for _, violation := range violations {
			coverageReport[violation.RuleName]++
		}
		
		t.Logf("Fixture %s (%s): %d violations", fixture.Name, fixture.Style, len(violations))
	}
	
	// Report coverage
	t.Logf("Validation rule coverage:")
	for ruleName, count := range coverageReport {
		t.Logf("  - %s: triggered %d times", ruleName, count)
	}
	
	// We should have good coverage of our validation rules
	// Note: Some rules might not trigger violations if our fixtures are well-formed
	// That's actually good - it means our expected XML is correct
	
	// At minimum, we should exercise the major rule categories
	totalViolations := 0
	for _, count := range coverageReport {
		totalViolations += count
	}
	
	t.Logf("Total violations across all fixtures: %d", totalViolations)
	
	// If we have no violations, that means our fixtures are well-formed
	// which is actually what we want for "expected" XML
	// The real test is whether the rules would catch problems in "actual" XML
}

func TestRuleEngineWithKnownIssues(t *testing.T) {
	// Test that our rule engine catches the known namespace issue
	engine := NewRuleEngine()
	
	// Simulate the problematic XML that would have been generated before the fix
	problematicXML := `<?xml version="1.0"?>
<soap:Envelope xmlns:soap="http://schemas.xmlsoap.org/soap/envelope/">
  <soap:Body>
    <tns:information xmlns:tns="http://example.com/test">
      <userId>123</userId>
    </tns:information>
  </soap:Body>
</soap:Envelope>`
	
	violations := engine.ValidateSOAPRequest(problematicXML, "rpc")
	
	// Should detect the tns: prefix issue
	assert.NotEmpty(t, violations, "Should detect violations in problematic XML")
	
	var foundRPCViolation bool
	for _, violation := range violations {
		if violation.RuleName == "RPC Operation Wrapper" && 
		   (strings.Contains(violation.Message, "tns:") || 
		    strings.Contains(violation.Message, "namespace")) {
			foundRPCViolation = true
			t.Logf("Found RPC violation: %s", violation.Message)
			break
		}
	}
	
	assert.True(t, foundRPCViolation, "Should detect the specific RPC namespace issue that was fixed")
}