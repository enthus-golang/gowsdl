// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at http://mozilla.org/MPL/2.0/.

package soap

import (
	"encoding/xml"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewEnhancedSOAPEnvelope_Basic(t *testing.T) {
	// Test with empty namespace
	env1 := NewEnhancedSOAPEnvelope("")
	assert.NotNil(t, env1)
	assert.Equal(t, XmlNsSoapEnv, env1.XMLNSSoap)
	assert.Empty(t, env1.XMLNSTns)
	assert.Empty(t, env1.XMLNSXSI)
	assert.Empty(t, env1.XMLNSXSD)

	// Test with namespace
	env2 := NewEnhancedSOAPEnvelope("http://example.com/service")
	assert.NotNil(t, env2)
	assert.Equal(t, XmlNsSoapEnv, env2.XMLNSSoap)
	assert.Equal(t, "http://example.com/service", env2.XMLNSTns)
	assert.Equal(t, "http://www.w3.org/2001/XMLSchema-instance", env2.XMLNSXSI)
	assert.Equal(t, "http://www.w3.org/2001/XMLSchema", env2.XMLNSXSD)
}

func TestEnhancedSOAPEnvelope_Marshal(t *testing.T) {
	env := NewEnhancedSOAPEnvelope("http://example.com/service")
	env.Body.Content = "test content"

	data, err := xml.Marshal(env)
	assert.NoError(t, err)
	assert.Contains(t, string(data), "soap:Envelope")
	assert.Contains(t, string(data), "xmlns:tns=\"http://example.com/service\"")
}