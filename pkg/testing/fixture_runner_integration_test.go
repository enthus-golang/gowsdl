// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at http://mozilla.org/MPL/2.0/.

// +build integration

package testing

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestFixtureRunnerWithRealCode(t *testing.T) {
	// Skip if not in integration test mode
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	// Create fixture runner with real code generation
	runner, err := NewFixtureRunnerWithRealCode()
	require.NoError(t, err, "Should create fixture runner")
	defer func() {
		if cleanupErr := runner.Cleanup(); cleanupErr != nil {
			t.Logf("Cleanup error: %v", cleanupErr)
		}
	}()

	// Create a temporary test fixture
	tempDir, err := ioutil.TempDir("", "fixture_test_*")
	require.NoError(t, err)
	defer os.RemoveAll(tempDir)

	// Create a simple WSDL for testing
	wsdlContent := `<?xml version="1.0" encoding="UTF-8"?>
<definitions xmlns="http://schemas.xmlsoap.org/wsdl/"
             xmlns:soap="http://schemas.xmlsoap.org/wsdl/soap/"
             xmlns:tns="http://example.com/simple"
             targetNamespace="http://example.com/simple">
             
  <types>
    <schema xmlns="http://www.w3.org/2001/XMLSchema"
            targetNamespace="http://example.com/simple">
      <element name="GetUserRequest">
        <complexType>
          <sequence>
            <element name="userId" type="string"/>
          </sequence>
        </complexType>
      </element>
      <element name="GetUserResponse">
        <complexType>
          <sequence>
            <element name="userName" type="string"/>
          </sequence>
        </complexType>
      </element>
    </schema>
  </types>
  
  <message name="GetUserRequestMessage">
    <part name="parameters" element="tns:GetUserRequest"/>
  </message>
  <message name="GetUserResponseMessage">
    <part name="parameters" element="tns:GetUserResponse"/>
  </message>
  
  <portType name="UserServicePortType">
    <operation name="GetUser">
      <input message="tns:GetUserRequestMessage"/>
      <output message="tns:GetUserResponseMessage"/>
    </operation>
  </portType>
  
  <binding name="UserServiceBinding" type="tns:UserServicePortType">
    <soap:binding style="document" transport="http://schemas.xmlsoap.org/soap/http"/>
    <operation name="GetUser">
      <soap:operation soapAction="getUser"/>
      <input>
        <soap:body use="literal"/>
      </input>
      <output>
        <soap:body use="literal"/>
      </output>
    </operation>
  </binding>
  
  <service name="UserService">
    <port name="UserServicePort" binding="tns:UserServiceBinding">
      <soap:address location="http://example.com/service"/>
    </port>
  </service>
</definitions>`

	// Create fixture directory structure
	fixtureDir := filepath.Join(tempDir, "document_literal")
	require.NoError(t, os.MkdirAll(fixtureDir, 0755))

	// Write WSDL file
	wsdlPath := filepath.Join(fixtureDir, "simple.wsdl")
	require.NoError(t, ioutil.WriteFile(wsdlPath, []byte(wsdlContent), 0644))

	// Create expected request XML
	expectedRequest := `<?xml version="1.0" encoding="UTF-8"?>
<soap:Envelope xmlns:soap="http://schemas.xmlsoap.org/soap/envelope/">
  <soap:Body>
    <GetUserRequest xmlns="http://example.com/simple">
      <userId>123</userId>
    </GetUserRequest>
  </soap:Body>
</soap:Envelope>`

	requestPath := filepath.Join(fixtureDir, "simple_request.xml") 
	require.NoError(t, ioutil.WriteFile(requestPath, []byte(expectedRequest), 0644))

	// Create test data
	testData := `{
		"operation": "GetUser",
		"userId": "123"
	}`

	testDataPath := filepath.Join(fixtureDir, "simple_test_data.json")
	require.NoError(t, ioutil.WriteFile(testDataPath, []byte(testData), 0644))

	// Load fixtures
	err = runner.LoadFixtures(tempDir)
	require.NoError(t, err, "Should load fixtures successfully")

	fixtures := runner.GetFixtures()
	assert.Len(t, fixtures, 1, "Should load one fixture")

	fixture := fixtures[0]
	assert.Equal(t, "simple", fixture.Name)
	assert.Equal(t, "document", fixture.Style)
	assert.Equal(t, wsdlPath, fixture.WSDLPath)
	assert.Equal(t, expectedRequest, fixture.ExpectedRequest)

	// Run simulation test (should work without real code generation)
	runner.useRealCode = false
	results, err := runner.RunAllTests()
	require.NoError(t, err, "Simulation tests should run successfully")
	assert.Len(t, results, 1, "Should have one test result")

	result := results[0]
	assert.Equal(t, "simple", result.TestName)
	assert.Equal(t, "document", result.Style)
	assert.NotEmpty(t, result.ActualXML, "Should have actual XML output")

	t.Logf("Simulation test result: Passed=%v, Error=%s", result.Passed, result.Error)
	if len(result.Differences) > 0 {
		t.Logf("Differences found:")
		for i, diff := range result.Differences {
			t.Logf("  %d. %s: %s", i+1, diff.Type, diff.Description)
		}
	}
}

func TestFixtureRunnerCleanup(t *testing.T) {
	// Test cleanup functionality
	runner, err := NewFixtureRunnerWithRealCode()
	require.NoError(t, err)

	tempDir := runner.tempDir
	assert.NotEmpty(t, tempDir, "Should have temp directory")

	// Check that temp directory exists
	_, err = os.Stat(tempDir)
	assert.NoError(t, err, "Temp directory should exist")

	// Cleanup
	err = runner.Cleanup()
	assert.NoError(t, err, "Cleanup should succeed")

	// Check that temp directory is removed
	_, err = os.Stat(tempDir)
	assert.True(t, os.IsNotExist(err), "Temp directory should be removed")
}

func TestFixtureRunnerSimulationMode(t *testing.T) {
	// Test that simulation mode works without temp directories
	runner := NewFixtureRunner()
	assert.False(t, runner.useRealCode, "Should default to simulation mode")
	assert.Empty(t, runner.tempDir, "Should not have temp directory in simulation mode")

	// Create a mock fixture
	fixture := FixtureTestCase{
		Name:  "test",
		Style: "rpc",
		TestData: map[string]interface{}{
			"operation": "GetUserInfo",
		},
	}

	// Test simulation
	actualXML := runner.simulateSOAPRequest(fixture)
	assert.Contains(t, actualXML, "GetUserInfo", "Should contain operation name")
	assert.Contains(t, actualXML, "soap:Envelope", "Should contain SOAP envelope")
	assert.Contains(t, actualXML, "xmlns=", "Should contain namespace")
}

func TestFixtureRunnerFixtureLoading(t *testing.T) {
	runner := NewFixtureRunner()

	// Create temporary fixture structure
	tempDir, err := ioutil.TempDir("", "fixture_loading_test_*")
	require.NoError(t, err)
	defer os.RemoveAll(tempDir)

	// Create RPC fixture
	rpcDir := filepath.Join(tempDir, "rpc_literal")
	require.NoError(t, os.MkdirAll(rpcDir, 0755))

	// Write RPC WSDL
	wsdlContent := `<definitions xmlns="http://schemas.xmlsoap.org/wsdl/"></definitions>`
	require.NoError(t, ioutil.WriteFile(filepath.Join(rpcDir, "basic_rpc.wsdl"), []byte(wsdlContent), 0644))

	// Write RPC request
	requestContent := `<soap:Envelope><soap:Body></soap:Body></soap:Envelope>`
	require.NoError(t, ioutil.WriteFile(filepath.Join(rpcDir, "basic_rpc_request.xml"), []byte(requestContent), 0644))

	// Write RPC test data
	testDataContent := `{"operation": "BasicRPC"}`
	require.NoError(t, ioutil.WriteFile(filepath.Join(rpcDir, "basic_rpc_test_data.json"), []byte(testDataContent), 0644))

	// Load fixtures
	err = runner.LoadFixtures(tempDir)
	require.NoError(t, err)

	fixtures := runner.GetFixtures()
	assert.Len(t, fixtures, 1, "Should load RPC fixture")

	fixture := fixtures[0]
	assert.Equal(t, "basic_rpc", fixture.Name)
	assert.Equal(t, "rpc", fixture.Style)
	assert.NotEmpty(t, fixture.WSDLPath)
	assert.Equal(t, requestContent, fixture.ExpectedRequest)
	assert.Contains(t, fixture.TestData, "operation")
}

func TestFixtureRunnerErrorHandling(t *testing.T) {
	runner := NewFixtureRunner()

	// Test loading from non-existent directory
	err := runner.LoadFixtures("/non/existent/path")
	assert.NoError(t, err, "Should handle non-existent directories gracefully")

	// Test with incomplete fixtures (missing files)
	tempDir, err := ioutil.TempDir("", "error_handling_test_*")
	require.NoError(t, err)
	defer os.RemoveAll(tempDir)

	docDir := filepath.Join(tempDir, "document_literal")
	require.NoError(t, os.MkdirAll(docDir, 0755))

	// Only create WSDL file, missing request and test data
	wsdlContent := `<definitions></definitions>`
	require.NoError(t, ioutil.WriteFile(filepath.Join(docDir, "incomplete.wsdl"), []byte(wsdlContent), 0644))

	err = runner.LoadFixtures(tempDir)
	assert.NoError(t, err, "Should handle incomplete fixtures gracefully")

	fixtures := runner.GetFixtures()
	assert.Len(t, fixtures, 0, "Should not load incomplete fixtures")
}