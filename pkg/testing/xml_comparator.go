// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at http://mozilla.org/MPL/2.0/.

package testing

import (
	"encoding/xml"
	"fmt"
	"regexp"
	"sort"
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

// XMLNode represents a parsed XML node for semantic comparison
type XMLNode struct {
	Name       xml.Name
	Attrs      []xml.Attr
	Content    string
	Children   []*XMLNode
	Namespaces map[string]string
}

// parseToMap converts XML to a structured tree for comparison
func (xc *XMLComparator) parseToMap(xmlStr string) (map[string]interface{}, error) {
	node, err := xc.parseToXMLNode(xmlStr)
	if err != nil {
		return nil, err
	}
	
	result := map[string]interface{}{
		"root": node,
	}
	
	return result, nil
}

// parseToXMLNode parses XML string into a structured node tree
func (xc *XMLComparator) parseToXMLNode(xmlStr string) (*XMLNode, error) {
	decoder := xml.NewDecoder(strings.NewReader(xmlStr))
	root, err := xc.parseNode(decoder, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to parse XML: %w", err)
	}
	return root, nil
}

// parseNode recursively parses XML tokens into XMLNode structure
func (xc *XMLComparator) parseNode(decoder *xml.Decoder, parent *XMLNode) (*XMLNode, error) {
	for {
		token, err := decoder.Token()
		if err != nil {
			return parent, err
		}
		
		switch t := token.(type) {
		case xml.StartElement:
			node := &XMLNode{
				Name:       t.Name,
				Attrs:      make([]xml.Attr, len(t.Attr)),
				Children:   make([]*XMLNode, 0),
				Namespaces: make(map[string]string),
			}
			
			// Copy attributes
			copy(node.Attrs, t.Attr)
			
			// Sort attributes for consistent comparison
			sort.Slice(node.Attrs, func(i, j int) bool {
				return node.Attrs[i].Name.Local < node.Attrs[j].Name.Local
			})
			
			// Parse child nodes
			child, err := xc.parseNode(decoder, node)
			if err != nil {
				return nil, err
			}
			
			if parent == nil {
				return child, nil
			}
			
			parent.Children = append(parent.Children, child)
			
		case xml.CharData:
			if parent != nil {
				content := strings.TrimSpace(string(t))
				if content != "" {
					parent.Content += content
				}
			}
			
		case xml.EndElement:
			return parent, nil
		}
	}
}

// compareElements compares two element maps using semantic XML comparison
func (xc *XMLComparator) compareElements(expected, actual map[string]interface{}, path string) []Difference {
	var diffs []Difference
	
	// Get root nodes
	expectedNode, ok := expected["root"].(*XMLNode)
	if !ok {
		return []Difference{{
			Type:        "error",
			Path:        path,
			Description: "Failed to parse expected XML structure",
		}}
	}
	
	actualNode, ok := actual["root"].(*XMLNode)
	if !ok {
		return []Difference{{
			Type:        "error",
			Path:        path,
			Description: "Failed to parse actual XML structure",
		}}
	}
	
	// Compare nodes semantically
	diffs = xc.compareNodes(expectedNode, actualNode, path)
	
	return diffs
}

// compareNodes compares two XMLNode structures recursively
func (xc *XMLComparator) compareNodes(expected, actual *XMLNode, path string) []Difference {
	var diffs []Difference
	
	if path == "" {
		path = "/"
	}
	
	// Compare element names
	if !xc.compareElementNames(expected.Name, actual.Name) {
		diffs = append(diffs, Difference{
			Type:        "different",
			Path:        path,
			Expected:    xc.formatElementName(expected.Name),
			Actual:      xc.formatElementName(actual.Name),
			Description: "Element name differs",
		})
	}
	
	// Compare attributes
	attrDiffs := xc.compareAttributes(expected.Attrs, actual.Attrs, path)
	diffs = append(diffs, attrDiffs...)
	
	// Compare text content
	if strings.TrimSpace(expected.Content) != strings.TrimSpace(actual.Content) {
		diffs = append(diffs, Difference{
			Type:        "different",
			Path:        path + "/text()",
			Expected:    strings.TrimSpace(expected.Content),
			Actual:      strings.TrimSpace(actual.Content),
			Description: "Text content differs",
		})
	}
	
	// Compare children
	childDiffs := xc.compareChildren(expected.Children, actual.Children, path)
	diffs = append(diffs, childDiffs...)
	
	return diffs
}

// compareElementNames compares element names considering namespace settings
func (xc *XMLComparator) compareElementNames(expected, actual xml.Name) bool {
	if xc.IgnoreNamespaces {
		return expected.Local == actual.Local
	}
	return expected == actual
}

// formatElementName formats element name for display
func (xc *XMLComparator) formatElementName(name xml.Name) string {
	if name.Space != "" {
		return fmt.Sprintf("{%s}%s", name.Space, name.Local)
	}
	return name.Local
}

// compareAttributes compares attribute lists
func (xc *XMLComparator) compareAttributes(expected, actual []xml.Attr, path string) []Difference {
	var diffs []Difference
	
	// Create maps for easier comparison
	expectedAttrs := make(map[string]string)
	actualAttrs := make(map[string]string)
	
	for _, attr := range expected {
		key := xc.formatElementName(attr.Name)
		expectedAttrs[key] = attr.Value
	}
	
	for _, attr := range actual {
		key := xc.formatElementName(attr.Name)
		actualAttrs[key] = attr.Value
	}
	
	// Check for missing attributes
	for key, value := range expectedAttrs {
		if actualValue, exists := actualAttrs[key]; !exists {
			diffs = append(diffs, Difference{
				Type:        "missing",
				Path:        path + "/@" + key,
				Expected:    value,
				Actual:      "",
				Description: "Missing attribute",
			})
		} else if actualValue != value {
			diffs = append(diffs, Difference{
				Type:        "different",
				Path:        path + "/@" + key,
				Expected:    value,
				Actual:      actualValue,
				Description: "Attribute value differs",
			})
		}
	}
	
	// Check for extra attributes
	for key, value := range actualAttrs {
		if _, exists := expectedAttrs[key]; !exists {
			diffs = append(diffs, Difference{
				Type:        "extra",
				Path:        path + "/@" + key,
				Expected:    "",
				Actual:      value,
				Description: "Extra attribute",
			})
		}
	}
	
	return diffs
}

// compareChildren compares child node lists
func (xc *XMLComparator) compareChildren(expected, actual []*XMLNode, path string) []Difference {
	var diffs []Difference
	
	if xc.IgnoreOrder {
		// Compare children ignoring order
		diffs = xc.compareChildrenUnordered(expected, actual, path)
	} else {
		// Compare children in order
		diffs = xc.compareChildrenOrdered(expected, actual, path)
	}
	
	return diffs
}

// compareChildrenOrdered compares children maintaining order
func (xc *XMLComparator) compareChildrenOrdered(expected, actual []*XMLNode, path string) []Difference {
	var diffs []Difference
	
	maxLen := len(expected)
	if len(actual) > maxLen {
		maxLen = len(actual)
	}
	
	for i := 0; i < maxLen; i++ {
		childPath := fmt.Sprintf("%s[%d]", path, i+1)
		
		if i >= len(expected) {
			// Extra child in actual
			diffs = append(diffs, Difference{
				Type:        "extra",
				Path:        childPath,
				Expected:    "",
				Actual:      xc.formatElementName(actual[i].Name),
				Description: "Extra child element",
			})
		} else if i >= len(actual) {
			// Missing child in actual
			diffs = append(diffs, Difference{
				Type:        "missing",
				Path:        childPath,
				Expected:    xc.formatElementName(expected[i].Name),
				Actual:      "",
				Description: "Missing child element",
			})
		} else {
			// Compare existing children
			childDiffs := xc.compareNodes(expected[i], actual[i], childPath)
			diffs = append(diffs, childDiffs...)
		}
	}
	
	return diffs
}

// compareChildrenUnordered compares children ignoring order
func (xc *XMLComparator) compareChildrenUnordered(expected, actual []*XMLNode, path string) []Difference {
	var diffs []Difference
	
	// Create maps by element name for unordered comparison
	expectedMap := make(map[string][]*XMLNode)
	actualMap := make(map[string][]*XMLNode)
	
	for _, child := range expected {
		key := xc.formatElementName(child.Name)
		expectedMap[key] = append(expectedMap[key], child)
	}
	
	for _, child := range actual {
		key := xc.formatElementName(child.Name)
		actualMap[key] = append(actualMap[key], child)
	}
	
	// Compare element counts and contents
	allKeys := make(map[string]bool)
	for key := range expectedMap {
		allKeys[key] = true
	}
	for key := range actualMap {
		allKeys[key] = true
	}
	
	for key := range allKeys {
		expectedNodes := expectedMap[key]
		actualNodes := actualMap[key]
		
		if len(expectedNodes) != len(actualNodes) {
			diffs = append(diffs, Difference{
				Type:        "different",
				Path:        path + "/" + key,
				Expected:    fmt.Sprintf("%d occurrences", len(expectedNodes)),
				Actual:      fmt.Sprintf("%d occurrences", len(actualNodes)),
				Description: "Different number of child elements",
			})
		} else {
			// Compare each occurrence
			for i, expectedNode := range expectedNodes {
				childPath := fmt.Sprintf("%s/%s[%d]", path, key, i+1)
				childDiffs := xc.compareNodes(expectedNode, actualNodes[i], childPath)
				diffs = append(diffs, childDiffs...)
			}
		}
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

// compareSOAPStructure compares SOAP envelope structures semantically
func (xc *XMLComparator) compareSOAPStructure(expected, actual *SOAPEnvelope) []Difference {
	var diffs []Difference
	
	// Parse and compare body content as XML
	expectedBody := strings.TrimSpace(string(expected.Body.Content))
	actualBody := strings.TrimSpace(string(actual.Body.Content))
	
	if expectedBody != "" && actualBody != "" {
		// Wrap body content in a root element for proper XML parsing
		expectedBodyXML := fmt.Sprintf("<root>%s</root>", expectedBody)
		actualBodyXML := fmt.Sprintf("<root>%s</root>", actualBody)
		
		// Use semantic XML comparison for body content
		result, err := xc.Compare(expectedBodyXML, actualBodyXML)
		if err != nil {
			diffs = append(diffs, Difference{
				Type:        "error",
				Path:        "soap:Body",
				Description: fmt.Sprintf("Failed to parse SOAP body content: %v", err),
			})
		} else if !result.Equal {
			// Remap differences to SOAP body context
			for _, diff := range result.Differences {
				bodyDiff := diff
				bodyDiff.Path = "soap:Body" + strings.TrimPrefix(diff.Path, "/root")
				diffs = append(diffs, bodyDiff)
			}
		}
	} else if expectedBody != actualBody {
		// Fallback to string comparison for empty or simple content
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