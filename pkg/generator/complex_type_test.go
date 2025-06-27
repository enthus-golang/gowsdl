package generator

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestComplexTypeWithInlineSequence(t *testing.T) {
	wsdlFile := "../../fixtures/complex_sequence_test.wsdl"
	
	gen, err := New(wsdlFile, WithPackage("testpkg"))
	require.NoError(t, err)
	
	code, err := gen.Generate(nil)
	require.NoError(t, err)
	
	// Get the generated code - concatenate all files
	var codeStr string
	for filename, content := range code {
		t.Logf("Generated file: %s", filename)
		codeStr += string(content) + "\n"
	}
	
	// Check that Person type has Address field with inline struct
	assert.Contains(t, codeStr, "type Person struct {")
	assert.Contains(t, codeStr, "Address struct {")
	
	// Check that Address has the expected fields, not just Value
	assert.Contains(t, codeStr, "Street string")
	assert.Contains(t, codeStr, "City string")
	assert.Contains(t, codeStr, "ZipCode string")
	
	// Check Contact field too
	assert.Contains(t, codeStr, "Contact struct {")
	assert.Contains(t, codeStr, "Phone string")
	assert.Contains(t, codeStr, "Email string")
	
	// Should NOT contain Value string for these complex types
	assert.NotContains(t, codeStr, "Address struct {\n\t\tValue string")
	assert.NotContains(t, codeStr, "Contact struct {\n\t\tValue string")
}

func TestComplexTypeFromOriginalWSDL(t *testing.T) {
	// Test with the original problematic WSDL
	wsdlFile := "/mnt/c/Users/info/repo/api/assets/go_webservice_demo.wsdl"
	
	// Skip if file doesn't exist
	if !fileExists(wsdlFile) {
		t.Skip("Original WSDL file not found")
	}
	
	gen, err := New(wsdlFile, WithPackage("godemo"))
	require.NoError(t, err)
	
	code, err := gen.Generate(nil)
	require.NoError(t, err)
	
	// Find the generated file
	require.NotEmpty(t, code, "Generated code should not be empty")
	
	// Look for the types file which contains the struct definitions
	var codeStr string
	for filename, content := range code {
		t.Logf("Generated file: %s (size: %d)", filename, len(content))
		if filename == "types" || len(content) > len(codeStr) {
			codeStr = string(content)
		}
	}
	
	require.NotEmpty(t, codeStr, "Should have generated code content")
	
	// Check that Abholadresse has proper fields
	// First, just check if the generated code contains the expected fields
	assert.Contains(t, codeStr, "Firmenname1 string", "Generated code should contain Firmenname1 field")
	assert.NotContains(t, codeStr, `Abholadresse struct {
		Value string`, "Abholadresse should not have just Value field")
}

func fileExists(path string) bool {
	// Simple file existence check
	_, err := os.Stat(path)
	return err == nil
}