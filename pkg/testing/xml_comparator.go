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

// XMLComparator provides semantic XML comparison functionality
type XMLComparator struct {
	IgnoreWhitespace bool
	IgnoreOrder      bool
	IgnoreNamespaces bool
}

// ComparisonResult represents the result of XML comparison
type ComparisonResult struct {
	Equal       bool
	Differences []Difference
}

// Difference represents a difference found between two XML documents
type Difference struct {
	Type        string // "missing", "extra", "different", "namespace"
	Path        string
	Expected    string
	Actual      string
	Description string
}

// NewXMLComparator creates a new XML comparator with default settings
func NewXMLComparator() *XMLComparator {
	return &XMLComparator{
		IgnoreWhitespace: true,
		IgnoreOrder:      false,
		IgnoreNamespaces: false,
	}
}

// Compare compares two XML documents semantically
func (xc *XMLComparator) Compare(expected, actual string) (*ComparisonResult, error) {
	result := &ComparisonResult{
		Equal:       true,
		Differences: make([]Difference, 0),
	}
	
	// Normalize XML for comparison
	expectedNorm, err := xc.normalizeXML(expected)
	if err != nil {
		return nil, fmt.Errorf("failed to normalize expected XML: %w", err)
	}
	
	actualNorm, err := xc.normalizeXML(actual)
	if err != nil {
		return nil, fmt.Errorf("failed to normalize actual XML: %w", err)
	}
	
	// Parse normalized XML
	expectedDoc, err := xc.parseToMap(expectedNorm)
	if err != nil {
		return nil, fmt.Errorf("failed to parse expected XML: %w", err)
	}
	
	actualDoc, err := xc.parseToMap(actualNorm)
	if err != nil {
		return nil, fmt.Errorf("failed to parse actual XML: %w", err)
	}
	
	// Compare documents
	diffs := xc.compareElements(expectedDoc, actualDoc, "")
	result.Differences = diffs
	result.Equal = len(diffs) == 0
	
	return result, nil
}

// normalizeXML normalizes XML for comparison
func (xc *XMLComparator) normalizeXML(xmlStr string) (string, error) {
	if xc.IgnoreWhitespace {
		// Remove extra whitespace between elements
		re := regexp.MustCompile(`>\s+<`)
		xmlStr = re.ReplaceAllString(xmlStr, "><")
		xmlStr = strings.TrimSpace(xmlStr)
	}
	
	return xmlStr, nil
}

// parseToMap converts XML to a map structure for comparison
func (xc *XMLComparator) parseToMap(xmlStr string) (map[string]interface{}, error) {
	// Simple XML parsing to map - this is a basic implementation
	// In production, you'd want a more robust XML-to-map parser
	
	// For now, return a simple structure
	// This would need to be expanded for full XML comparison
	result := map[string]interface{}{
		"xml": xmlStr,
	}
	
	return result, nil
}

// compareElements compares two element maps
func (xc *XMLComparator) compareElements(expected, actual map[string]interface{}, path string) []Difference {
	var diffs []Difference
	
	// Simple string comparison for now
	// This would need to be expanded for full semantic comparison
	expectedXML := expected["xml"].(string)
	actualXML := actual["xml"].(string)
	
	if expectedXML != actualXML {
		diffs = append(diffs, Difference{
			Type:        "different",
			Path:        "/",  // Set root path if path is empty
			Expected:    expectedXML,
			Actual:      actualXML,
			Description: "XML content differs",
		})
	}
	
	return diffs
}

// CompareSOAPRequests compares two SOAP request XMLs with SOAP-specific rules
func (xc *XMLComparator) CompareSOAPRequests(expected, actual string) (*ComparisonResult, error) {
	// First do standard XML comparison
	result, err := xc.Compare(expected, actual)
	if err != nil {
		return nil, err
	}
	
	// Add SOAP-specific validations
	expectedEnv, err := xc.parseSOAPEnvelope(expected)
	if err != nil {
		return nil, fmt.Errorf("failed to parse expected SOAP envelope: %w", err)
	}
	
	actualEnv, err := xc.parseSOAPEnvelope(actual)
	if err != nil {
		return nil, fmt.Errorf("failed to parse actual SOAP envelope: %w", err)
	}
	
	// Check SOAP envelope structure
	soapDiffs := xc.compareSOAPStructure(expectedEnv, actualEnv)
	result.Differences = append(result.Differences, soapDiffs...)
	result.Equal = result.Equal && len(soapDiffs) == 0
	
	return result, nil
}

// parseSOAPEnvelope parses SOAP envelope for comparison
func (xc *XMLComparator) parseSOAPEnvelope(xmlStr string) (*SOAPEnvelope, error) {
	var envelope SOAPEnvelope
	if err := xml.Unmarshal([]byte(xmlStr), &envelope); err != nil {
		return nil, err
	}
	return &envelope, nil
}

// compareSOAPStructure compares SOAP envelope structures
func (xc *XMLComparator) compareSOAPStructure(expected, actual *SOAPEnvelope) []Difference {
	var diffs []Difference
	
	// Compare body content
	expectedBody := strings.TrimSpace(string(expected.Body.Content))
	actualBody := strings.TrimSpace(string(actual.Body.Content))
	
	if expectedBody != actualBody {
		diffs = append(diffs, Difference{
			Type:        "different",
			Path:        "soap:Body",
			Expected:    expectedBody,
			Actual:      actualBody,
			Description: "SOAP body content differs",
		})
	}
	
	return diffs
}

// ExtractOperationName extracts the operation name from SOAP body
func ExtractOperationName(soapXML string) (string, error) {
	var envelope SOAPEnvelope
	if err := xml.Unmarshal([]byte(soapXML), &envelope); err != nil {
		return "", err
	}
	
	// Parse the body content to find the first element (operation)
	bodyContent := string(envelope.Body.Content)
	
	// Extract first element name using regex
	re := regexp.MustCompile(`<(\w+:)?(\w+)`)
	matches := re.FindStringSubmatch(bodyContent)
	if len(matches) >= 3 {
		return matches[2], nil
	}
	
	return "", fmt.Errorf("could not extract operation name from SOAP body")
}