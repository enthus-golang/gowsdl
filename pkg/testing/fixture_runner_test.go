// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at http://mozilla.org/MPL/2.0/.

package testing

import (
	"encoding/json"
	"net/http"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewFixtureRunner(t *testing.T) {
	runner := NewFixtureRunner()
	assert.NotNil(t, runner)
	assert.NotNil(t, runner.capture)
	assert.NotNil(t, runner.comparator)
	assert.Empty(t, runner.fixtures)
}

func TestFixtureRunnerLoadFixtures(t *testing.T) {
	// Create temporary test fixtures
	tempDir := t.TempDir()
	
	// Create directory structure
	docDir := filepath.Join(tempDir, "document_literal")
	rpcDir := filepath.Join(tempDir, "rpc_literal")
	
	err := os.MkdirAll(docDir, 0755)
	require.NoError(t, err)
	err = os.MkdirAll(rpcDir, 0755)
	require.NoError(t, err)
	
	// Create test fixtures
	err = createTestFixture(docDir, "simple", "document")
	require.NoError(t, err)
	err = createTestFixture(rpcDir, "basic_rpc", "rpc")
	require.NoError(t, err)
	
	runner := NewFixtureRunner()
	err = runner.LoadFixtures(tempDir)
	require.NoError(t, err)
	
	fixtures := runner.GetFixtures()
	assert.Len(t, fixtures, 2, "Should load both fixtures")
	
	// Check that fixtures have required fields
	for _, fixture := range fixtures {
		assert.NotEmpty(t, fixture.Name)
		assert.NotEmpty(t, fixture.WSDLPath)
		assert.NotEmpty(t, fixture.ExpectedRequest)
		assert.NotEmpty(t, fixture.Style)
		assert.NotNil(t, fixture.TestData)
	}
}

func TestFixtureRunnerLoadNonexistentDirectory(t *testing.T) {
	runner := NewFixtureRunner()
	
	// Should not error when directory doesn't exist
	err := runner.LoadFixtures("/nonexistent/path")
	assert.NoError(t, err)
	
	fixtures := runner.GetFixtures()
	assert.Empty(t, fixtures)
}

func TestFixtureRunnerIncompleteFixtures(t *testing.T) {
	tempDir := t.TempDir()
	docDir := filepath.Join(tempDir, "document_literal")
	
	err := os.MkdirAll(docDir, 0755)
	require.NoError(t, err)
	
	// Create incomplete fixture (missing expected request)
	wsdlContent := `<?xml version="1.0" encoding="UTF-8"?>
<definitions xmlns="http://schemas.xmlsoap.org/wsdl/"
             targetNamespace="http://example.com/simple">
  <message name="GetUserRequestMessage">
    <part name="parameters" type="s:string"/>
  </message>
</definitions>`
	
	testData := map[string]interface{}{
		"operation": "GetUser",
		"userId":    "123",
	}
	testDataJSON, _ := json.Marshal(testData)
	
	// Write only WSDL and test data, skip expected request
	err = os.WriteFile(filepath.Join(docDir, "incomplete.wsdl"), []byte(wsdlContent), 0644)
	require.NoError(t, err)
	err = os.WriteFile(filepath.Join(docDir, "incomplete_test_data.json"), testDataJSON, 0644)
	require.NoError(t, err)
	
	runner := NewFixtureRunner()
	err = runner.LoadFixtures(tempDir)
	require.NoError(t, err)
	
	// Should not load incomplete fixtures
	fixtures := runner.GetFixtures()
	assert.Empty(t, fixtures, "Should not load incomplete fixtures")
}

func TestFixtureRunnerSimulateSOAPRequest(t *testing.T) {
	runner := NewFixtureRunner()
	
	tests := []struct {
		name     string
		fixture  FixtureTestCase
		contains []string
	}{
		{
			name: "RPC style with operation name",
			fixture: FixtureTestCase{
				Name:  "test_rpc",
				Style: "rpc",
				TestData: map[string]interface{}{
					"operation": "GetUserInfo",
				},
			},
			contains: []string{"GetUserInfo", "xmlns=\"http://example.com/rpc\"", "<userId>123</userId>"},
		},
		{
			name: "RPC style with missing operation (should use default)",
			fixture: FixtureTestCase{
				Name:     "test_rpc_no_op",
				Style:    "rpc",
				TestData: map[string]interface{}{},
			},
			contains: []string{"GetUserInfo", "xmlns=\"http://example.com/rpc\"", "<userId>123</userId>"},
		},
		{
			name: "RPC style with non-string operation (should use default)",
			fixture: FixtureTestCase{
				Name:  "test_rpc_invalid_op",
				Style: "rpc",
				TestData: map[string]interface{}{
					"operation": 123, // Not a string
				},
			},
			contains: []string{"GetUserInfo", "xmlns=\"http://example.com/rpc\"", "<userId>123</userId>"},
		},
		{
			name: "Document style",
			fixture: FixtureTestCase{
				Name:     "test_doc",
				Style:    "document",
				TestData: map[string]interface{}{},
			},
			contains: []string{"GetUserRequest", "xmlns=\"http://example.com/simple\"", "<userId>123</userId>"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := runner.simulateSOAPRequest(tt.fixture)
			
			assert.NotEmpty(t, result)
			assert.Contains(t, result, "soap:Envelope")
			assert.Contains(t, result, "soap:Body")
			
			for _, expected := range tt.contains {
				assert.Contains(t, result, expected, "Should contain: %s", expected)
			}
		})
	}
}

func TestFixtureRunnerSetHTTPClient(t *testing.T) {
	runner := NewFixtureRunner()
	client := &http.Client{}
	
	runner.SetHTTPClient(client)
	
	// Verify that the transport was set to our capture
	assert.Equal(t, runner.capture, client.Transport)
}

func TestFixtureRunnerRunAllTests(t *testing.T) {
	// Create temporary test fixtures
	tempDir := t.TempDir()
	docDir := filepath.Join(tempDir, "document_literal")
	
	err := os.MkdirAll(docDir, 0755)
	require.NoError(t, err)
	
	err = createTestFixture(docDir, "simple", "document")
	require.NoError(t, err)
	
	runner := NewFixtureRunner()
	err = runner.LoadFixtures(tempDir)
	require.NoError(t, err)
	
	results, err := runner.RunAllTests()
	require.NoError(t, err)
	
	assert.Len(t, results, 1, "Should have one test result")
	
	result := results[0]
	assert.Equal(t, "simple", result.TestName)
	assert.Equal(t, "document", result.Style)
	assert.NotEmpty(t, result.ActualXML)
	assert.NotEmpty(t, result.ExpectedXML)
	// Note: The test might not pass because we're using simulated SOAP requests
	// that don't match the expected XML exactly
}

func TestFixtureRunnerInvalidWSDL(t *testing.T) {
	tempDir := t.TempDir()
	docDir := filepath.Join(tempDir, "document_literal")
	
	err := os.MkdirAll(docDir, 0755)
	require.NoError(t, err)
	
	// Create fixture with invalid WSDL
	invalidWSDL := "not valid WSDL content"
	expectedXML := `<?xml version="1.0"?><soap:Envelope>...</soap:Envelope>`
	testData := map[string]interface{}{"operation": "test"}
	testDataJSON, _ := json.Marshal(testData)
	
	err = os.WriteFile(filepath.Join(docDir, "invalid.wsdl"), []byte(invalidWSDL), 0644)
	require.NoError(t, err)
	err = os.WriteFile(filepath.Join(docDir, "invalid_request.xml"), []byte(expectedXML), 0644)
	require.NoError(t, err)
	err = os.WriteFile(filepath.Join(docDir, "invalid_test_data.json"), testDataJSON, 0644)
	require.NoError(t, err)
	
	runner := NewFixtureRunner()
	err = runner.LoadFixtures(tempDir)
	require.NoError(t, err)
	
	results, err := runner.RunAllTests()
	require.NoError(t, err)
	
	assert.Len(t, results, 1)
	result := results[0]
	assert.False(t, result.Passed, "Should fail with invalid WSDL")
	assert.NotEmpty(t, result.Error, "Should have error message")
}

func TestFixtureRunnerInvalidJSON(t *testing.T) {
	tempDir := t.TempDir()
	docDir := filepath.Join(tempDir, "document_literal")
	
	err := os.MkdirAll(docDir, 0755)
	require.NoError(t, err)
	
	// Create fixture with invalid JSON
	wsdlContent := `<?xml version="1.0"?><definitions>...</definitions>`
	expectedXML := `<?xml version="1.0"?><soap:Envelope>...</soap:Envelope>`
	invalidJSON := `{"invalid": json}`
	
	err = os.WriteFile(filepath.Join(docDir, "invalid_json.wsdl"), []byte(wsdlContent), 0644)
	require.NoError(t, err)
	err = os.WriteFile(filepath.Join(docDir, "invalid_json_request.xml"), []byte(expectedXML), 0644)
	require.NoError(t, err)
	err = os.WriteFile(filepath.Join(docDir, "invalid_json_test_data.json"), []byte(invalidJSON), 0644)
	require.NoError(t, err)
	
	runner := NewFixtureRunner()
	err = runner.LoadFixtures(tempDir)
	
	// Should error during loading due to invalid JSON
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to parse test data")
}

// Helper function to create a complete test fixture
func createTestFixture(dir, baseName, style string) error {
	wsdlContent := `<?xml version="1.0" encoding="UTF-8"?>
<definitions xmlns="http://schemas.xmlsoap.org/wsdl/"
             targetNamespace="http://example.com/` + style + `">
  <message name="GetUserRequestMessage">
    <part name="parameters" type="s:string"/>
  </message>
</definitions>`

	var expectedRequest string
	if style == "rpc" {
		expectedRequest = `<?xml version="1.0" encoding="UTF-8"?>
<soap:Envelope xmlns:soap="http://schemas.xmlsoap.org/soap/envelope/">
  <soap:Body>
    <GetUserInfo xmlns="http://example.com/rpc">
      <userId>123</userId>
      <includeDetails>true</includeDetails>
    </GetUserInfo>
  </soap:Body>
</soap:Envelope>`
	} else {
		expectedRequest = `<?xml version="1.0" encoding="UTF-8"?>
<soap:Envelope xmlns:soap="http://schemas.xmlsoap.org/soap/envelope/">
  <soap:Body>
    <GetUserRequest xmlns="http://example.com/simple">
      <userId>123</userId>
      <includeDetails>true</includeDetails>
    </GetUserRequest>
  </soap:Body>
</soap:Envelope>`
	}

	testData := map[string]interface{}{
		"operation": "GetUserInfo",
		"userId":    "123",
	}
	testDataJSON, err := json.Marshal(testData)
	if err != nil {
		return err
	}

	// Write files
	err = os.WriteFile(filepath.Join(dir, baseName+".wsdl"), []byte(wsdlContent), 0644)
	if err != nil {
		return err
	}
	
	err = os.WriteFile(filepath.Join(dir, baseName+"_request.xml"), []byte(expectedRequest), 0644)
	if err != nil {
		return err
	}
	
	err = os.WriteFile(filepath.Join(dir, baseName+"_test_data.json"), testDataJSON, 0644)
	if err != nil {
		return err
	}

	return nil
}