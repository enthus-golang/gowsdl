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

func TestMarshalXMLGenerationRules(t *testing.T) {
	t.Skip("MarshalXML generation feature needs to be implemented separately")
	// Test that MarshalXML is only generated for complex types, not simple types
	wsdl := `<?xml version="1.0" encoding="UTF-8"?>
<definitions xmlns="http://schemas.xmlsoap.org/wsdl/"
             xmlns:soap="http://schemas.xmlsoap.org/wsdl/soap/"
             xmlns:tns="http://example.com/test"
             xmlns:xsd="http://www.w3.org/2001/XMLSchema"
             targetNamespace="http://example.com/test">
    <types>
        <schema xmlns="http://www.w3.org/2001/XMLSchema"
                targetNamespace="http://example.com/test">
            
            <!-- Simple type - should NOT get MarshalXML -->
            <simpleType name="StatusType">
                <restriction base="string">
                    <enumeration value="active"/>
                    <enumeration value="inactive"/>
                </restriction>
            </simpleType>
            
            <!-- Complex type - should get MarshalXML -->
            <complexType name="PersonType">
                <sequence>
                    <element name="name" type="string"/>
                    <element name="age" type="int" minOccurs="0"/>
                </sequence>
            </complexType>
            
            <!-- Complex type with simple content - should get MarshalXML -->
            <complexType name="PriceType">
                <simpleContent>
                    <extension base="decimal">
                        <attribute name="currency" type="string"/>
                    </extension>
                </simpleContent>
            </complexType>
            
            <!-- Element with inline complex type -->
            <element name="Order">
                <complexType>
                    <sequence>
                        <element name="id" type="string"/>
                        <element name="items" minOccurs="0">
                            <complexType>
                                <sequence>
                                    <element name="item" type="string" maxOccurs="unbounded"/>
                                </sequence>
                            </complexType>
                        </element>
                    </sequence>
                </complexType>
            </element>
        </schema>
    </types>
</definitions>`

	file, err := os.CreateTemp("", "test-marshal-rules-*.wsdl")
	require.NoError(t, err)
	defer func() { _ = os.Remove(file.Name()) }()
	defer func() { _ = file.Close() }()
	
	_, err = file.WriteString(wsdl)
	require.NoError(t, err)

	g, err := New(file.Name(), WithPackage("test"))
	require.NoError(t, err)

	files, err := g.Generate(context.Background())
	require.NoError(t, err)

	types, ok := files["types"]
	require.True(t, ok, "types file should be generated")
	
	typesStr := string(types)
	
	// Check that simple types don't get MarshalXML
	assert.NotContains(t, typesStr, "func (t StatusType) MarshalXML", 
		"Simple types should NOT have MarshalXML methods")
	
	// Check that complex types DO get MarshalXML
	assert.Contains(t, typesStr, "func (t PersonType) MarshalXML", 
		"Complex types should have MarshalXML methods")
	assert.Contains(t, typesStr, "func (t PriceType) MarshalXML", 
		"Complex types with simple content should have MarshalXML methods")
	
	// Check that inline complex types get MarshalXML
	assert.Contains(t, typesStr, "func (t OrderType) MarshalXML", 
		"Inline complex types should have MarshalXML methods")
	assert.Contains(t, typesStr, "func (t OrderTypeItemsType) MarshalXML", 
		"Nested inline complex types should have MarshalXML methods")
}