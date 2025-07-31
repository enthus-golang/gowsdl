// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at http://mozilla.org/MPL/2.0/.

package soap

import (
	"encoding/xml"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSOAPBody_ErrorFromFault_Simple(t *testing.T) {
	// Test no fault
	body1 := &SOAPBody{
		faultOccurred: false,
		Fault:         &SOAPFault{String: "test"},
	}
	err := body1.ErrorFromFault()
	assert.NoError(t, err)
	assert.Nil(t, body1.Fault)

	// Test with fault
	body2 := &SOAPBody{
		faultOccurred: true,
		Fault:         &SOAPFault{String: "error occurred"},
	}
	err = body2.ErrorFromFault()
	assert.Error(t, err)
	assert.Equal(t, "error occurred", err.Error())
}

func TestSOAPBodyResponse_ErrorFromFault_Simple(t *testing.T) {
	// Test no fault
	body1 := &SOAPBodyResponse{
		faultOccurred: false,
		Fault:         &SOAPFault{String: "test"},
	}
	err := body1.ErrorFromFault()
	assert.NoError(t, err)
	assert.Nil(t, body1.Fault)

	// Test with fault
	body2 := &SOAPBodyResponse{
		faultOccurred: true,
		Fault:         &SOAPFault{String: "error occurred"},
	}
	err = body2.ErrorFromFault()
	assert.Error(t, err)
	assert.Equal(t, "error occurred", err.Error())
}

func TestSOAPFault_Error_Simple(t *testing.T) {
	// Test fault string only
	fault1 := &SOAPFault{String: "simple error"}
	assert.Equal(t, "simple error", fault1.Error())

	// Test with nil detail
	fault2 := &SOAPFault{String: "error with nil detail", Detail: nil}
	assert.Equal(t, "error with nil detail", fault2.Error())
}

func TestHTTPError_Error_Simple(t *testing.T) {
	err := &HTTPError{
		StatusCode:   404,
		ResponseBody: []byte("Not Found"),
	}
	assert.Equal(t, "HTTP Status 404: Not Found", err.Error())
}

func TestNewWSSSecurityHeader_Simple(t *testing.T) {
	header := NewWSSSecurityHeader("user", "pass", "token1", "1")
	
	assert.NotNil(t, header)
	assert.Equal(t, WssNsWSSE, header.XmlNSWsse)
	assert.Equal(t, "1", header.MustUnderstand)
	
	assert.NotNil(t, header.Token)
	assert.Equal(t, "token1", header.Token.Id)
	
	assert.NotNil(t, header.Token.Username)
	assert.Equal(t, "user", header.Token.Username.Data)
	
	assert.NotNil(t, header.Token.Password)
	assert.Equal(t, "pass", header.Token.Password.Data)
}

func TestClient_Methods(t *testing.T) {
	client := NewClient("http://example.com")
	
	// Test URL
	assert.Equal(t, "http://example.com", client.URL())
	
	// Test AddHeader
	assert.Empty(t, client.headers)
	client.AddHeader("header1")
	assert.Len(t, client.headers, 1)
	
	// Test SetHeaders
	client.SetHeaders("new1", "new2")
	assert.Len(t, client.headers, 2)
	
	// Test HTTP headers
	headers := map[string]string{"X-Test": "value"}
	client.SetHttpClientHeaders(headers)
	assert.Equal(t, headers, client.GetHttpClientHeaders())
	
	// Test AddMIMEMultipartAttachment
	attachment := MIMEMultipartAttachment{Name: "test.txt", Data: []byte("data")}
	client.AddMIMEMultipartAttachment(attachment)
	assert.Len(t, client.attachments, 1)
}

func TestSOAPBodyResponse_UnmarshalXML_Simple(t *testing.T) {
	// Test nil content error
	body := &SOAPBodyResponse{Body: nil}
	err := xml.Unmarshal([]byte("<Body></Body>"), body)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "Body must be a pointer to a struct")
	
	// Test proper SOAP body unmarshal
	type TestValue struct {
		XMLName xml.Name `xml:"TestValue"`
		Value   string   `xml:"value"`
	}
	
	body2 := &SOAPBodyResponse{
		Body: &TestValue{},
		Fault:   &SOAPFault{},
	}
	
	// Proper SOAP Body structure - UnmarshalXML expects the content element directly
	xmlData := `<Body><TestValue><value>test</value></TestValue></Body>`
	err = xml.Unmarshal([]byte(xmlData), body2)
	assert.NoError(t, err)
	
	content := body2.Body.(*TestValue)
	assert.Equal(t, "test", content.Value)
	
	// Test SOAP fault unmarshal
	body3 := &SOAPBodyResponse{
		Body: &TestValue{},
		Fault:   &SOAPFault{},
	}
	
	xmlFault := `<Body xmlns:soap="http://schemas.xmlsoap.org/soap/envelope/">
		<soap:Fault>
			<faultcode>Server</faultcode>
			<faultstring>Test error</faultstring>
		</soap:Fault>
	</Body>`
	
	err = xml.Unmarshal([]byte(xmlFault), body3)
	assert.NoError(t, err)
	assert.True(t, body3.Fault != nil)
	assert.Equal(t, "Server", body3.Fault.Code)
	assert.Equal(t, "Test error", body3.Fault.String)
	assert.Nil(t, body3.Body) // Body should be nil when fault occurs
}