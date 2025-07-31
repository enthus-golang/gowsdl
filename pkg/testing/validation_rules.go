// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at http://mozilla.org/MPL/2.0/.

package testing

import (
	"encoding/xml"
	"fmt"
	"regexp"
	"strings"
)

// ValidationRule represents a rule for validating SOAP requests
type ValidationRule interface {
	Name() string
	Description() string
	Validate(soapXML string, style string) []RuleViolation
}

// RuleViolation represents a violation of a validation rule
type RuleViolation struct {
	RuleName    string
	Severity    string // "error", "warning", "info"
	Message     string
	Path        string
	Suggestion  string
}

// RuleEngine manages and executes validation rules
type RuleEngine struct {
	rules []ValidationRule
}

// NewRuleEngine creates a new rule engine with default rules
func NewRuleEngine() *RuleEngine {
	engine := &RuleEngine{
		rules: make([]ValidationRule, 0),
	}
	
	// Add default rules
	engine.AddRule(&RPCOperationWrapperRule{})
	engine.AddRule(&DocumentStyleRule{})
	engine.AddRule(&NamespaceConsistencyRule{})
	engine.AddRule(&SOAPEnvelopeRule{})
	engine.AddRule(&RequiredElementsRule{})
	
	return engine
}

// AddRule adds a validation rule to the engine
func (re *RuleEngine) AddRule(rule ValidationRule) {
	re.rules = append(re.rules, rule)
}

// ValidateSOAPRequest validates a SOAP request against all rules
func (re *RuleEngine) ValidateSOAPRequest(soapXML, style string) []RuleViolation {
	var violations []RuleViolation
	
	for _, rule := range re.rules {
		ruleViolations := rule.Validate(soapXML, style)
		violations = append(violations, ruleViolations...)
	}
	
	return violations
}

// RPCOperationWrapperRule validates RPC-style operation wrappers
type RPCOperationWrapperRule struct{}

func (r *RPCOperationWrapperRule) Name() string {
	return "RPC Operation Wrapper"
}

func (r *RPCOperationWrapperRule) Description() string {
	return "RPC-style operations must have operation name as wrapper element in correct namespace"
}

func (r *RPCOperationWrapperRule) Validate(soapXML string, style string) []RuleViolation {
	var violations []RuleViolation
	
	if style != "rpc" {
		return violations
	}
	
	// Parse SOAP envelope
	var envelope SOAPEnvelope
	if err := xml.Unmarshal([]byte(soapXML), &envelope); err != nil {
		violations = append(violations, RuleViolation{
			RuleName: r.Name(),
			Severity: "error",
			Message:  "Failed to parse SOAP envelope",
			Path:     "soap:Envelope",
		})
		return violations
	}
	
	bodyContent := string(envelope.Body.Content)
	
	// Check for operation wrapper with namespace
	if !strings.Contains(bodyContent, `xmlns=`) {
		violations = append(violations, RuleViolation{
			RuleName:   r.Name(),
			Severity:   "error",
			Message:    "RPC operation element missing namespace declaration",
			Path:       "soap:Body/*[1]",
			Suggestion: "Operation element should have xmlns attribute with target namespace",
		})
	}
	
	// Check for tns: prefix (indicates incorrect namespace handling)
	if strings.Contains(bodyContent, "tns:") {
		violations = append(violations, RuleViolation{
			RuleName:   r.Name(),
			Severity:   "error",
			Message:    "RPC operation using namespace prefix instead of default namespace",
			Path:       "soap:Body/*[1]",
			Suggestion: "Use xmlns=\"namespace\" instead of tns:operation",
		})
	}
	
	return violations
}

// DocumentStyleRule validates document-style operations
type DocumentStyleRule struct{}

func (r *DocumentStyleRule) Name() string {
	return "Document Style"
}

func (r *DocumentStyleRule) Description() string {
	return "Document-style operations should use element-based messages without operation wrappers"
}

func (r *DocumentStyleRule) Validate(soapXML string, style string) []RuleViolation {
	var violations []RuleViolation
	
	if style != "document" {
		return violations
	}
	
	// Parse SOAP envelope
	var envelope SOAPEnvelope
	if err := xml.Unmarshal([]byte(soapXML), &envelope); err != nil {
		violations = append(violations, RuleViolation{
			RuleName: r.Name(),
			Severity: "error",
			Message:  "Failed to parse SOAP envelope",
			Path:     "soap:Envelope",
		})
		return violations
	}
	
	bodyContent := string(envelope.Body.Content)
	
	// Document style should not have operation wrappers
	// This is a simple check - in practice, you'd analyze the WSDL
	// to know what the expected element names should be
	if strings.Contains(strings.ToLower(bodyContent), "operation") {
		violations = append(violations, RuleViolation{
			RuleName:   r.Name(),
			Severity:   "warning",
			Message:    "Document-style request may contain operation wrapper",
			Path:       "soap:Body",
			Suggestion: "Document style should use message elements directly",
		})
	}
	
	return violations
}

// NamespaceConsistencyRule validates namespace consistency
type NamespaceConsistencyRule struct{}

func (r *NamespaceConsistencyRule) Name() string {
	return "Namespace Consistency"
}

func (r *NamespaceConsistencyRule) Description() string {
	return "Namespace declarations and usage should be consistent"
}

func (r *NamespaceConsistencyRule) Validate(soapXML string, style string) []RuleViolation {
	var violations []RuleViolation
	
	// Simple check for undefined namespace prefixes using regex
	// Look for patterns like "tns:something" without corresponding xmlns:tns declaration
	
	// Extract all namespace prefix usages
	prefixUsagePattern := regexp.MustCompile(`<(\w+):[\w-]+`)
	prefixMatches := prefixUsagePattern.FindAllStringSubmatch(soapXML, -1)
	
	usedPrefixes := make(map[string]bool)
	for _, match := range prefixMatches {
		if len(match) > 1 {
			prefix := match[1]
			// Skip standard prefixes
			if prefix != "soap" && prefix != "xml" && prefix != "xsi" && prefix != "xsd" {
				usedPrefixes[prefix] = true
			}
		}
	}
	
	// Check if each used prefix is declared
	for prefix := range usedPrefixes {
		// Look for xmlns:prefix declaration
		declPattern := regexp.MustCompile(`xmlns:` + prefix + `=`)
		if !declPattern.MatchString(soapXML) {
			violations = append(violations, RuleViolation{
				RuleName:   r.Name(),
				Severity:   "error",
				Message:    fmt.Sprintf("Undefined namespace prefix: %s", prefix),
				Path:       fmt.Sprintf("*[@%s:*]", prefix),
				Suggestion: fmt.Sprintf("Declare namespace prefix: xmlns:%s=\"namespace-uri\"", prefix),
			})
		}
	}
	
	return violations
}

// SOAPEnvelopeRule validates basic SOAP envelope structure
type SOAPEnvelopeRule struct{}

func (r *SOAPEnvelopeRule) Name() string {
	return "SOAP Envelope Structure"
}

func (r *SOAPEnvelopeRule) Description() string {
	return "SOAP requests must have valid envelope structure"
}

func (r *SOAPEnvelopeRule) Validate(soapXML string, style string) []RuleViolation {
	var violations []RuleViolation
	
	// Check for SOAP envelope
	if !strings.Contains(soapXML, "soap:Envelope") && !strings.Contains(soapXML, "Envelope") {
		violations = append(violations, RuleViolation{
			RuleName: r.Name(),
			Severity: "error",
			Message:  "Missing SOAP envelope",
			Path:     "/",
			Suggestion: "Request must be wrapped in soap:Envelope element",
		})
	}
	
	// Check for SOAP body
	if !strings.Contains(soapXML, "soap:Body") && !strings.Contains(soapXML, "Body") {
		violations = append(violations, RuleViolation{
			RuleName: r.Name(),
			Severity: "error",
			Message:  "Missing SOAP body",
			Path:     "soap:Envelope",
			Suggestion: "Envelope must contain soap:Body element",
		})
	}
	
	// Check for SOAP namespace
	if !strings.Contains(soapXML, "http://schemas.xmlsoap.org/soap/envelope/") {
		violations = append(violations, RuleViolation{
			RuleName: r.Name(),
			Severity: "error",
			Message:  "Missing or incorrect SOAP namespace",
			Path:     "soap:Envelope",
			Suggestion: "Use xmlns:soap=\"http://schemas.xmlsoap.org/soap/envelope/\"",
		})
	}
	
	return violations
}

// RequiredElementsRule validates that required elements are present
type RequiredElementsRule struct{}

func (r *RequiredElementsRule) Name() string {
	return "Required Elements"
}

func (r *RequiredElementsRule) Description() string {
	return "Required elements must be present in the request"
}

func (r *RequiredElementsRule) Validate(soapXML string, style string) []RuleViolation {
	var violations []RuleViolation
	
	// This would need to be enhanced with WSDL schema information
	// to know which elements are actually required
	
	// Basic check for empty body
	var envelope SOAPEnvelope
	if err := xml.Unmarshal([]byte(soapXML), &envelope); err == nil {
		bodyContent := strings.TrimSpace(string(envelope.Body.Content))
		if bodyContent == "" {
			violations = append(violations, RuleViolation{
				RuleName:   r.Name(),
				Severity:   "error",
				Message:    "SOAP body is empty",
				Path:       "soap:Body",
				Suggestion: "Body must contain operation or message elements",
			})
		}
	}
	
	return violations
}