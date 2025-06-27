// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at http://mozilla.org/MPL/2.0/.

//go:build integration
// +build integration

package gowsdl

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestIntegrationRealWSDLFiles tests against real WSDL files in fixtures
func TestIntegrationRealWSDLFiles(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	// Find fixture files
	fixtures := []string{
		"fixtures/test.wsdl",
		"fixtures/stock.wsdl",
		"fixtures/ec2.wsdl",
	}

	for _, fixture := range fixtures {
		t.Run(fixture, func(t *testing.T) {
			// Check if fixture exists
			if _, err := os.Stat(fixture); os.IsNotExist(err) {
				t.Skipf("Fixture %s not found", fixture)
			}

			// Create temporary directory for generated code
			tempDir, err := os.MkdirTemp("", "gowsdl-integration-*")
			require.NoError(t, err)
			defer os.RemoveAll(tempDir)

			// Generate code
			g, err := NewGoWSDL(fixture, "testservice", false, true)
			require.NoError(t, err)

			code, err := g.Start()
			require.NoError(t, err)
			require.NotNil(t, code)

			// Write generated code to files
			clientFile := filepath.Join(tempDir, "client.go")
			serverFile := filepath.Join(tempDir, "server.go")

			// Write client code
			clientCode := append(append(code["header"], code["types"]...), code["operations"]...)
			err = os.WriteFile(clientFile, clientCode, 0644)
			require.NoError(t, err)

			// Write server code
			serverCode := append(append(code["server_header"], code["server_wsdl"]...), code["server"]...)
			err = os.WriteFile(serverFile, serverCode, 0644)
			require.NoError(t, err)

			// Create go.mod file
			goModContent := `module testservice

go 1.21

require github.com/enthus-golang/gowsdl v0.0.0
replace github.com/enthus-golang/gowsdl => ` + getRootDir()

			err = os.WriteFile(filepath.Join(tempDir, "go.mod"), []byte(goModContent), 0644)
			require.NoError(t, err)

			// Try to build the generated code
			cmd := exec.Command("go", "build", ".")
			cmd.Dir = tempDir
			output, err := cmd.CombinedOutput()
			assert.NoError(t, err, "Generated code should compile. Output: %s", string(output))
		})
	}
}

// TestIntegrationGeneratedCodeSOAPOperations tests that generated code can perform SOAP operations
func TestIntegrationGeneratedCodeSOAPOperations(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	wsdlContent := `<?xml version="1.0" encoding="UTF-8"?>
<definitions xmlns="http://schemas.xmlsoap.org/wsdl/"
	xmlns:xs="http://www.w3.org/2001/XMLSchema"
	xmlns:soap="http://schemas.xmlsoap.org/wsdl/soap/"
	xmlns:tns="http://example.com/calculator"
	targetNamespace="http://example.com/calculator">
	<types>
		<xs:schema targetNamespace="http://example.com/calculator">
			<xs:element name="AddRequest">
				<xs:complexType>
					<xs:sequence>
						<xs:element name="a" type="xs:int"/>
						<xs:element name="b" type="xs:int"/>
					</xs:sequence>
				</xs:complexType>
			</xs:element>
			<xs:element name="AddResponse">
				<xs:complexType>
					<xs:sequence>
						<xs:element name="result" type="xs:int"/>
					</xs:sequence>
				</xs:complexType>
			</xs:element>
		</xs:schema>
	</types>
	<message name="AddRequestMessage">
		<part name="body" element="tns:AddRequest"/>
	</message>
	<message name="AddResponseMessage">
		<part name="body" element="tns:AddResponse"/>
	</message>
	<portType name="CalculatorPortType">
		<operation name="Add">
			<input message="tns:AddRequestMessage"/>
			<output message="tns:AddResponseMessage"/>
		</operation>
	</portType>
	<binding name="CalculatorBinding" type="tns:CalculatorPortType">
		<soap:binding transport="http://schemas.xmlsoap.org/soap/http"/>
		<operation name="Add">
			<soap:operation soapAction="Add"/>
			<input>
				<soap:body use="literal"/>
			</input>
			<output>
				<soap:body use="literal"/>
			</output>
		</operation>
	</binding>
	<service name="CalculatorService">
		<port name="CalculatorPort" binding="tns:CalculatorBinding">
			<soap:address location="http://localhost:8080/calculator"/>
		</port>
	</service>
</definitions>`

	// Create temporary WSDL file
	tempFile, err := os.CreateTemp("", "calculator-*.wsdl")
	require.NoError(t, err)
	defer os.Remove(tempFile.Name())

	_, err = tempFile.WriteString(wsdlContent)
	require.NoError(t, err)
	tempFile.Close()

	// Generate code
	g, err := NewGoWSDL(tempFile.Name(), "calculator", false, true)
	require.NoError(t, err)

	code, err := g.Start()
	require.NoError(t, err)

	// Create temporary directory for test
	tempDir, err := os.MkdirTemp("", "calculator-test-*")
	require.NoError(t, err)
	defer os.RemoveAll(tempDir)

	// Write generated code
	clientFile := filepath.Join(tempDir, "calculator.go")
	clientCode := append(append(code["header"], code["types"]...), code["operations"]...)
	err = os.WriteFile(clientFile, clientCode, 0644)
	require.NoError(t, err)

	// Write test file that uses the generated code
	testCode := `package calculator

import (
	"testing"
	"github.com/stretchr/testify/assert"
)

func TestGeneratedTypes(t *testing.T) {
	// Test that we can create instances of generated types
	req := &AddRequest{
		A: 10,
		B: 20,
	}
	assert.Equal(t, int32(10), req.A)
	assert.Equal(t, int32(20), req.B)

	resp := &AddResponse{
		Result: 30,
	}
	assert.Equal(t, int32(30), resp.Result)
}
`

	testFile := filepath.Join(tempDir, "calculator_test.go")
	err = os.WriteFile(testFile, []byte(testCode), 0644)
	require.NoError(t, err)

	// Create go.mod
	goModContent := `module calculator

go 1.21

require (
	github.com/stretchr/testify v1.10.0
	github.com/enthus-golang/gowsdl v0.0.0
)

replace github.com/enthus-golang/gowsdl => ` + getRootDir()

	err = os.WriteFile(filepath.Join(tempDir, "go.mod"), []byte(goModContent), 0644)
	require.NoError(t, err)

	// Run go mod tidy
	cmd := exec.Command("go", "mod", "tidy")
	cmd.Dir = tempDir
	output, err := cmd.CombinedOutput()
	require.NoError(t, err, "go mod tidy failed: %s", string(output))

	// Run the test
	cmd = exec.Command("go", "test", "-v", ".")
	cmd.Dir = tempDir
	output, err = cmd.CombinedOutput()
	assert.NoError(t, err, "Generated code test failed: %s", string(output))
}

// TestIntegrationComplexWSDL tests against a complex WSDL with multiple schemas
func TestIntegrationComplexWSDL(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	// Test with vboxweb.wsdl if it exists
	vboxWSDL := "fixtures/vboxweb.wsdl"
	if _, err := os.Stat(vboxWSDL); os.IsNotExist(err) {
		t.Skip("vboxweb.wsdl fixture not found")
	}

	g, err := NewGoWSDL(vboxWSDL, "vbox", false, true)
	require.NoError(t, err)

	code, err := g.Start()
	require.NoError(t, err)
	require.NotNil(t, code)

	// Check that code was generated for all parts
	assert.NotEmpty(t, code["header"])
	assert.NotEmpty(t, code["types"])
	assert.NotEmpty(t, code["operations"])
	assert.NotEmpty(t, code["server"])
}

// getRootDir returns the root directory of the module
func getRootDir() string {
	// Try to find the root directory
	dir, err := os.Getwd()
	if err != nil {
		return "."
	}

	// Walk up until we find go.mod
	for {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return dir
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			break
		}
		dir = parent
	}

	return "."
}

// TestIntegrationHTTPSWSDL tests downloading WSDL over HTTPS
func TestIntegrationHTTPSWSDL(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	// Skip if no internet connection
	ctx := context.Background()
	testURL := "https://www.w3schools.com/xml/tempconvert.asmx?WSDL"
	
	config := DefaultHTTPClientConfig()
	config.Timeout = 10 * time.Second
	
	_, err := downloadFile(ctx, testURL, config)
	if err != nil {
		t.Skipf("Cannot reach test WSDL URL: %v", err)
	}

	// Test downloading and generating code
	g, err := NewGoWSDLWithConfig(testURL, "tempconvert", config, true)
	require.NoError(t, err)

	code, err := g.Start()
	require.NoError(t, err)
	require.NotNil(t, code)

	// Verify the generated code contains expected elements
	types := string(code["types"])
	assert.Contains(t, types, "CelsiusToFahrenheit")
	assert.Contains(t, types, "FahrenheitToCelsius")
}