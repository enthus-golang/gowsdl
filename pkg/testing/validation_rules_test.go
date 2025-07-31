// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at http://mozilla.org/MPL/2.0/.

package testing

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRuleEngine(t *testing.T) {
	engine := NewRuleEngine()
	assert.NotNil(t, engine)
	assert.Len(t, engine.rules, 5, "Should have 5 default rules")
}

func TestRPCOperationWrapperRule(t *testing.T) {
	rule := &RPCOperationWrapperRule{}
	
	tests := []struct {
		name           string
		soapXML        string
		style          string
		expectedErrors int
	}{
		{
			name: "Valid RPC operation with xmlns",
			soapXML: `<?xml version="1.0"?>
<soap:Envelope xmlns:soap="http://schemas.xmlsoap.org/soap/envelope/">
  <soap:Body>
    <getUser xmlns="http://example.com/rpc">
      <userId>123</userId>
    </getUser>
  </soap:Body>
</soap:Envelope>`,
			style:          "rpc",
			expectedErrors: 0,
		},
		{
			name: "Invalid RPC operation with tns prefix",
			soapXML: `<?xml version="1.0"?>
<soap:Envelope xmlns:soap="http://schemas.xmlsoap.org/soap/envelope/">
  <soap:Body>
    <tns:information xmlns:tns="http://example.com/test">
      <userId>123</userId>
    </tns:information>
  </soap:Body>
</soap:Envelope>`,
			style:          "rpc",
			expectedErrors: 1, // Should detect tns: prefix usage
		},
		{
			name: "RPC operation missing namespace",
			soapXML: `<?xml version="1.0"?>
<soap:Envelope xmlns:soap="http://schemas.xmlsoap.org/soap/envelope/">
  <soap:Body>
    <getUser>
      <userId>123</userId>
    </getUser>
  </soap:Body>
</soap:Envelope>`,
			style:          "rpc",
			expectedErrors: 1, // Should detect missing xmlns
		},
		{
			name: "Document style - should be ignored",
			soapXML: `<?xml version="1.0"?>
<soap:Envelope xmlns:soap="http://schemas.xmlsoap.org/soap/envelope/">
  <soap:Body>
    <tns:information xmlns:tns="http://example.com/test">
      <userId>123</userId>
    </tns:information>
  </soap:Body>
</soap:Envelope>`,
			style:          "document",
			expectedErrors: 0, // Should ignore document style
		},
		{
			name: "Invalid XML",
			soapXML: `<invalid xml`,
			style:          "rpc",
			expectedErrors: 1, // Should detect parse error
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			violations := rule.Validate(tt.soapXML, tt.style)
			assert.Len(t, violations, tt.expectedErrors, "Expected %d violations, got %d", tt.expectedErrors, len(violations))
			
			for _, violation := range violations {
				assert.Equal(t, rule.Name(), violation.RuleName)
				assert.NotEmpty(t, violation.Message)
				assert.NotEmpty(t, violation.Severity)
			}
		})
	}
}

func TestDocumentStyleRule(t *testing.T) {
	rule := &DocumentStyleRule{}
	
	tests := []struct {
		name           string
		soapXML        string
		style          string
		expectedErrors int
	}{
		{
			name: "Valid document style",
			soapXML: `<?xml version="1.0"?>
<soap:Envelope xmlns:soap="http://schemas.xmlsoap.org/soap/envelope/">
  <soap:Body>
    <GetUserRequest xmlns="http://example.com/simple">
      <userId>123</userId>
    </GetUserRequest>
  </soap:Body>
</soap:Envelope>`,
			style:          "document",
			expectedErrors: 0,
		},
		{
			name: "Document with operation wrapper (warning)",
			soapXML: `<?xml version="1.0"?>
<soap:Envelope xmlns:soap="http://schemas.xmlsoap.org/soap/envelope/">
  <soap:Body>
    <getUserOperation xmlns="http://example.com/simple">
      <userId>123</userId>
    </getUserOperation>
  </soap:Body>
</soap:Envelope>`,
			style:          "document",
			expectedErrors: 1, // Should warn about operation wrapper
		},
		{
			name: "RPC style - should be ignored",
			soapXML: `<?xml version="1.0"?>
<soap:Envelope xmlns:soap="http://schemas.xmlsoap.org/soap/envelope/">
  <soap:Body>
    <getUserOperation xmlns="http://example.com/rpc">
      <userId>123</userId>
    </getUserOperation>
  </soap:Body>
</soap:Envelope>`,
			style:          "rpc",
			expectedErrors: 0, // Should ignore RPC style
		},
		{
			name: "Invalid XML",
			soapXML: `<invalid xml`,
			style:          "document",
			expectedErrors: 1, // Should detect parse error
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			violations := rule.Validate(tt.soapXML, tt.style)
			assert.Len(t, violations, tt.expectedErrors)
			
			for _, violation := range violations {
				assert.Equal(t, rule.Name(), violation.RuleName)
				assert.NotEmpty(t, violation.Message)
				assert.NotEmpty(t, violation.Severity)
			}
		})
	}
}

func TestNamespaceConsistencyRule(t *testing.T) {
	rule := &NamespaceConsistencyRule{}
	
	tests := []struct {
		name           string
		soapXML        string
		style          string
		expectedErrors int
	}{
		{
			name: "Valid namespaces",
			soapXML: `<?xml version="1.0"?>
<soap:Envelope xmlns:soap="http://schemas.xmlsoap.org/soap/envelope/">
  <soap:Body>
    <tns:getUser xmlns:tns="http://example.com/test">
      <tns:userId>123</tns:userId>
    </tns:getUser>
  </soap:Body>
</soap:Envelope>`,
			style:          "rpc",
			expectedErrors: 0,
		},
		{
			name: "Undefined prefix",
			soapXML: `<?xml version="1.0"?>
<soap:Envelope xmlns:soap="http://schemas.xmlsoap.org/soap/envelope/">
  <soap:Body>
    <tns:getUser>
      <tns:userId>123</tns:userId>
    </tns:getUser>
  </soap:Body>
</soap:Envelope>`,
			style:          "rpc",
			expectedErrors: 1, // Should detect undefined tns prefix
		},
		{
			name: "Standard prefixes are allowed",
			soapXML: `<?xml version="1.0"?>
<soap:Envelope xmlns:soap="http://schemas.xmlsoap.org/soap/envelope/">
  <soap:Body xml:lang="en">
    <getUser xmlns="http://example.com/test">
      <userId>123</userId>
    </getUser>
  </soap:Body>
</soap:Envelope>`,
			style:          "rpc",
			expectedErrors: 0, // xml: prefix is standard
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			violations := rule.Validate(tt.soapXML, tt.style)
			assert.Len(t, violations, tt.expectedErrors)
			
			for _, violation := range violations {
				assert.Equal(t, rule.Name(), violation.RuleName)
				assert.NotEmpty(t, violation.Message)
				assert.NotEmpty(t, violation.Severity)
			}
		})
	}
}

func TestSOAPEnvelopeRule(t *testing.T) {
	rule := &SOAPEnvelopeRule{}
	
	tests := []struct {
		name           string
		soapXML        string
		style          string
		expectedErrors int
	}{
		{
			name: "Valid SOAP envelope",
			soapXML: `<?xml version="1.0"?>
<soap:Envelope xmlns:soap="http://schemas.xmlsoap.org/soap/envelope/">
  <soap:Body>
    <getUser xmlns="http://example.com/test">
      <userId>123</userId>
    </getUser>
  </soap:Body>
</soap:Envelope>`,
			style:          "rpc",
			expectedErrors: 0,
		},
		{
			name: "Missing envelope",
			soapXML: `<?xml version="1.0"?>
<root>
  <getUser>
    <userId>123</userId>
  </getUser>
</root>`,
			style:          "rpc",
			expectedErrors: 3, // Missing envelope, body, and namespace
		},
		{
			name: "Missing body",
			soapXML: `<?xml version="1.0"?>
<soap:Envelope xmlns:soap="http://schemas.xmlsoap.org/soap/envelope/">
  <getUser>
    <userId>123</userId>
  </getUser>
</soap:Envelope>`,
			style:          "rpc",
			expectedErrors: 1, // Missing body
		},
		{
			name: "Wrong namespace",
			soapXML: `<?xml version="1.0"?>
<soap:Envelope xmlns:soap="http://wrong.namespace.com/">
  <soap:Body>
    <getUser>
      <userId>123</userId>
    </getUser>
  </soap:Body>
</soap:Envelope>`,
			style:          "rpc",
			expectedErrors: 1, // Wrong SOAP namespace
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			violations := rule.Validate(tt.soapXML, tt.style)
			assert.Len(t, violations, tt.expectedErrors)
			
			for _, violation := range violations {
				assert.Equal(t, rule.Name(), violation.RuleName)
				assert.NotEmpty(t, violation.Message)
				assert.NotEmpty(t, violation.Severity)
			}
		})
	}
}

func TestRequiredElementsRule(t *testing.T) {
	rule := &RequiredElementsRule{}
	
	tests := []struct {
		name           string
		soapXML        string
		style          string
		expectedErrors int
	}{
		{
			name: "Valid request with content",
			soapXML: `<?xml version="1.0"?>
<soap:Envelope xmlns:soap="http://schemas.xmlsoap.org/soap/envelope/">
  <soap:Body>
    <getUser xmlns="http://example.com/test">
      <userId>123</userId>
    </getUser>
  </soap:Body>
</soap:Envelope>`,
			style:          "rpc",
			expectedErrors: 0,
		},
		{
			name: "Empty body",
			soapXML: `<?xml version="1.0"?>
<soap:Envelope xmlns:soap="http://schemas.xmlsoap.org/soap/envelope/">
  <soap:Body>
  </soap:Body>
</soap:Envelope>`,
			style:          "rpc",
			expectedErrors: 1, // Empty body
		},
		{
			name: "Body with only whitespace",
			soapXML: `<?xml version="1.0"?>
<soap:Envelope xmlns:soap="http://schemas.xmlsoap.org/soap/envelope/">
  <soap:Body>
    
  </soap:Body>
</soap:Envelope>`,
			style:          "rpc",
			expectedErrors: 1, // Empty body (whitespace only)
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			violations := rule.Validate(tt.soapXML, tt.style)
			assert.Len(t, violations, tt.expectedErrors)
			
			for _, violation := range violations {
				assert.Equal(t, rule.Name(), violation.RuleName)
				assert.NotEmpty(t, violation.Message)
				assert.NotEmpty(t, violation.Severity)
			}
		})
	}
}

func TestRuleEngineIntegration(t *testing.T) {
	engine := NewRuleEngine()
	
	// Test with a problematic RPC request (the namespace issue)
	soapXML := `<?xml version="1.0"?>
<soap:Envelope xmlns:soap="http://schemas.xmlsoap.org/soap/envelope/">
  <soap:Body>
    <tns:information xmlns:tns="http://example.com/test">
      <userId>123</userId>
    </tns:information>
  </soap:Body>
</soap:Envelope>`
	
	violations := engine.ValidateSOAPRequest(soapXML, "rpc")
	
	// Should catch the tns: prefix issue
	assert.NotEmpty(t, violations, "Should detect violations")
	
	var rpcViolation *RuleViolation
	for _, v := range violations {
		if v.RuleName == "RPC Operation Wrapper" {
			rpcViolation = &v
			break
		}
	}
	
	require.NotNil(t, rpcViolation, "Should detect RPC operation wrapper violation")
	assert.Contains(t, rpcViolation.Message, "tns:")
	assert.Equal(t, "error", rpcViolation.Severity)
}

func TestRuleEngineAddRule(t *testing.T) {
	engine := NewRuleEngine()
	initialCount := len(engine.rules)
	
	// Create a custom rule
	customRule := &testCustomRule{}
	engine.AddRule(customRule)
	
	assert.Len(t, engine.rules, initialCount+1, "Should have added one rule")
	
	// Test that custom rule is called
	violations := engine.ValidateSOAPRequest("<test/>", "rpc")
	var customViolation *RuleViolation
	for _, v := range violations {
		if v.RuleName == "Test Custom Rule" {
			customViolation = &v
			break
		}
	}
	
	require.NotNil(t, customViolation, "Custom rule should have been executed")
	assert.Equal(t, "test violation", customViolation.Message)
}

// testCustomRule is a test rule for testing rule engine extensibility
type testCustomRule struct{}

func (r *testCustomRule) Name() string {
	return "Test Custom Rule"
}

func (r *testCustomRule) Description() string {
	return "A test rule for unit testing"
}

func (r *testCustomRule) Validate(soapXML string, style string) []RuleViolation {
	return []RuleViolation{
		{
			RuleName: r.Name(),
			Severity: "info",
			Message:  "test violation",
			Path:     "/",
		},
	}
}