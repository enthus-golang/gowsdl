// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at http://mozilla.org/MPL/2.0/.

package generator

import (
	"context"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestEmptyComplexTypeMarshalXML(t *testing.T) {
	wsdl := `<?xml version="1.0" encoding="UTF-8"?>
<definitions xmlns="http://schemas.xmlsoap.org/wsdl/"
             xmlns:soap="http://schemas.xmlsoap.org/wsdl/soap/"
             xmlns:tns="http://example.com/contact"
             xmlns:xsd="http://www.w3.org/2001/XMLSchema"
             targetNamespace="http://example.com/contact">
    <types>
        <schema xmlns="http://www.w3.org/2001/XMLSchema"
                targetNamespace="http://example.com/contact">
            
            <complexType name="ContactType">
                <sequence>
                    <element name="name" type="string"/>
                    <element name="phoneInfo" minOccurs="0">
                        <complexType>
                            <sequence>
                                <element name="phone" type="string" minOccurs="0"/>
                                <element name="mobile" type="string" minOccurs="0"/>
                            </sequence>
                        </complexType>
                    </element>
                </sequence>
            </complexType>
            
            <element name="Contact" type="tns:ContactType"/>
        </schema>
    </types>
</definitions>`

	file, err := os.CreateTemp("", "test-wsdl-*.wsdl")
	require.NoError(t, err)
	defer func() { _ = os.Remove(file.Name()) }()
	defer func() { _ = file.Close() }()
	
	_, err = file.WriteString(wsdl)
	require.NoError(t, err)

	g, err := New(file.Name(), WithPackage("contact"))
	require.NoError(t, err)

	files, err := g.Generate(context.Background())
	require.NoError(t, err)

	types, ok := files["types"]
	require.True(t, ok, "types file should be generated")
	
	typesStr := string(types)
	t.Logf("Generated types:\n%s", typesStr)
	
	// Check that MarshalXML is generated for the inline type
	assert.Contains(t, typesStr, "func (t ContactTypePhoneInfoType) MarshalXML", 
		"Should generate MarshalXML method for inline type")
	assert.Contains(t, typesStr, "t.Phone == \"\"", 
		"Should check if phone field is empty")
	assert.Contains(t, typesStr, "t.Mobile == \"\"", 
		"Should check if mobile field is empty")
	assert.Contains(t, typesStr, "return nil", 
		"Should return nil to skip marshaling empty struct")
}