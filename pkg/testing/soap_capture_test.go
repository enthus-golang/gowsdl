// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at http://mozilla.org/MPL/2.0/.

package testing

import (
	"io"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSOAPCaptureBasic(t *testing.T) {
	capture := NewSOAPCapture()
	
	// Test initial state
	assert.Equal(t, 0, capture.GetRequestCount())
	assert.Nil(t, capture.GetLastRequest())
	
	// Test reset
	capture.Reset()
	assert.Equal(t, 0, capture.GetRequestCount())
}

func TestSOAPCaptureRoundTrip(t *testing.T) {
	capture := NewSOAPCapture()
	
	// Create a test request
	soapBody := `<?xml version="1.0"?>
<soap:Envelope xmlns:soap="http://schemas.xmlsoap.org/soap/envelope/">
  <soap:Body>
    <getUser xmlns="http://example.com/test">
      <userId>123</userId>
    </getUser>
  </soap:Body>
</soap:Envelope>`
	
	req := &http.Request{
		Method: "POST",
		URL:    &url.URL{Scheme: "http", Host: "example.com", Path: "/service"},
		Header: make(http.Header),
		Body:   io.NopCloser(strings.NewReader(soapBody)),
	}
	req.Header.Set("Content-Type", "text/xml; charset=utf-8")
	req.Header.Set("SOAPAction", "getUser")
	
	// Execute round trip
	resp, err := capture.RoundTrip(req)
	require.NoError(t, err)
	require.NotNil(t, resp)
	
	// Verify response
	assert.Equal(t, 200, resp.StatusCode)
	assert.Equal(t, "text/xml; charset=utf-8", resp.Header.Get("Content-Type"))
	
	// Verify request was captured
	assert.Equal(t, 1, capture.GetRequestCount())
	
	lastReq := capture.GetLastRequest()
	require.NotNil(t, lastReq)
	assert.Equal(t, "http://example.com/service", lastReq.URL)
	assert.Equal(t, "text/xml; charset=utf-8", lastReq.Headers["Content-Type"])
	assert.Equal(t, "getUser", lastReq.Headers["Soapaction"]) // HTTP headers are canonicalized by Go
	assert.Contains(t, lastReq.Body, "getUser")
	assert.NotNil(t, lastReq.XMLDoc)
	assert.NotNil(t, lastReq.XMLDoc.Envelope)
}

func TestSOAPCaptureRequestWithoutBody(t *testing.T) {
	capture := NewSOAPCapture()
	
	req := &http.Request{
		Method: "GET",
		URL:    &url.URL{Scheme: "http", Host: "example.com", Path: "/service"},
		Header: make(http.Header),
	}
	
	resp, err := capture.RoundTrip(req)
	require.NoError(t, err)
	require.NotNil(t, resp)
	
	// Should still capture the request
	assert.Equal(t, 1, capture.GetRequestCount())
	
	lastReq := capture.GetLastRequest()
	require.NotNil(t, lastReq)
	assert.Empty(t, lastReq.Body)
	assert.Nil(t, lastReq.XMLDoc)
}

func TestSOAPCaptureInvalidXML(t *testing.T) {
	capture := NewSOAPCapture()
	
	invalidXML := "not xml content"
	req := &http.Request{
		Method: "POST",
		URL:    &url.URL{Scheme: "http", Host: "example.com", Path: "/service"},
		Header: make(http.Header),
		Body:   io.NopCloser(strings.NewReader(invalidXML)),
	}
	
	resp, err := capture.RoundTrip(req)
	require.NoError(t, err)
	require.NotNil(t, resp)
	
	// Should capture request but XMLDoc should be nil due to parse error
	assert.Equal(t, 1, capture.GetRequestCount())
	
	lastReq := capture.GetLastRequest()
	require.NotNil(t, lastReq)
	assert.Equal(t, invalidXML, lastReq.Body)
	assert.Nil(t, lastReq.XMLDoc) // Should be nil due to parse error
}

func TestSOAPCaptureThreadSafety(t *testing.T) {
	capture := NewSOAPCapture()
	const numGoroutines = 10
	const requestsPerGoroutine = 5
	
	var wg sync.WaitGroup
	wg.Add(numGoroutines)
	
	// Launch multiple goroutines making requests concurrently
	for i := 0; i < numGoroutines; i++ {
		go func(goroutineID int) {
			defer wg.Done()
			
			for j := 0; j < requestsPerGoroutine; j++ {
				soapBody := `<?xml version="1.0"?>
<soap:Envelope xmlns:soap="http://schemas.xmlsoap.org/soap/envelope/">
  <soap:Body>
    <getUser xmlns="http://example.com/test">
      <userId>` + string(rune('0'+goroutineID)) + string(rune('0'+j)) + `</userId>
    </getUser>
  </soap:Body>
</soap:Envelope>`
				
				req := &http.Request{
					Method: "POST",
					URL:    &url.URL{Scheme: "http", Host: "example.com", Path: "/service"},
					Header: make(http.Header),
					Body:   io.NopCloser(strings.NewReader(soapBody)),
				}
				
				_, err := capture.RoundTrip(req)
				assert.NoError(t, err)
			}
		}(i)
	}
	
	wg.Wait()
	
	// Verify all requests were captured
	expectedCount := numGoroutines * requestsPerGoroutine
	assert.Equal(t, expectedCount, capture.GetRequestCount())
	
	// Verify we can safely access the last request
	lastReq := capture.GetLastRequest()
	assert.NotNil(t, lastReq)
	assert.NotEmpty(t, lastReq.Body)
}

func TestSOAPCaptureMultipleHeaders(t *testing.T) {
	capture := NewSOAPCapture()
	
	req := &http.Request{
		Method: "POST",
		URL:    &url.URL{Scheme: "http", Host: "example.com", Path: "/service"},
		Header: make(http.Header),
		Body:   io.NopCloser(strings.NewReader("<test/>")),
	}
	
	// Add multiple values for the same header
	req.Header.Add("Accept", "text/xml")
	req.Header.Add("Accept", "application/xml")
	req.Header.Set("Content-Type", "text/xml; charset=utf-8")
	
	_, err := capture.RoundTrip(req)
	require.NoError(t, err)
	
	lastReq := capture.GetLastRequest()
	require.NotNil(t, lastReq)
	
	// Should join multiple header values with ", "
	assert.Equal(t, "text/xml, application/xml", lastReq.Headers["Accept"])
	assert.Equal(t, "text/xml; charset=utf-8", lastReq.Headers["Content-Type"])
}

func TestSOAPCaptureReset(t *testing.T) {
	capture := NewSOAPCapture()
	
	// Add some requests
	for i := 0; i < 3; i++ {
		req := &http.Request{
			Method: "POST",
			URL:    &url.URL{Scheme: "http", Host: "example.com", Path: "/service"},
			Header: make(http.Header),
			Body:   io.NopCloser(strings.NewReader("<test/>")),
		}
		
		_, err := capture.RoundTrip(req)
		require.NoError(t, err)
	}
	
	assert.Equal(t, 3, capture.GetRequestCount())
	
	// Reset should clear everything
	capture.Reset()
	assert.Equal(t, 0, capture.GetRequestCount())
	assert.Nil(t, capture.GetLastRequest())
}

func TestSOAPCaptureParseXML(t *testing.T) {
	capture := NewSOAPCapture()
	
	tests := []struct {
		name      string
		xmlData   string
		shouldErr bool
	}{
		{
			name: "Valid SOAP envelope",
			xmlData: `<?xml version="1.0"?>
<soap:Envelope xmlns:soap="http://schemas.xmlsoap.org/soap/envelope/">
  <soap:Body>
    <getUser xmlns="http://example.com/test">
      <userId>123</userId>
    </getUser>
  </soap:Body>
</soap:Envelope>`,
			shouldErr: false,
		},
		{
			name:      "Invalid XML",
			xmlData:   "<invalid xml>",
			shouldErr: true,
		},
		{
			name:      "Empty string",
			xmlData:   "",
			shouldErr: true,
		},
		{
			name: "XML without SOAP envelope",
			xmlData: `<?xml version="1.0"?>
<root>
  <child>value</child>
</root>`,
			shouldErr: true, // Will fail to unmarshal into SOAPEnvelope
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			xmlDoc, err := capture.parseXML(tt.xmlData)
			
			if tt.shouldErr {
				assert.Error(t, err)
				assert.Nil(t, xmlDoc)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, xmlDoc)
				assert.NotNil(t, xmlDoc.Envelope)
				assert.Equal(t, tt.xmlData, xmlDoc.Raw)
			}
		})
	}
}

func TestSOAPCaptureBodyRestore(t *testing.T) {
	capture := NewSOAPCapture()
	
	originalBody := "test body content"
	req := &http.Request{
		Method: "POST",
		URL:    &url.URL{Scheme: "http", Host: "example.com", Path: "/service"},
		Header: make(http.Header),
		Body:   io.NopCloser(strings.NewReader(originalBody)),
	}
	
	// Execute round trip
	_, err := capture.RoundTrip(req)
	require.NoError(t, err)
	
	// Verify original request body is still readable
	bodyBytes, err := io.ReadAll(req.Body)
	require.NoError(t, err)
	assert.Equal(t, originalBody, string(bodyBytes))
	
	// Verify captured body matches
	lastReq := capture.GetLastRequest()
	require.NotNil(t, lastReq)
	assert.Equal(t, originalBody, lastReq.Body)
}

func TestSOAPCaptureReadError(t *testing.T) {
	capture := NewSOAPCapture()
	
	// Create a request with a body that will error on read
	req := &http.Request{
		Method: "POST",
		URL:    &url.URL{Scheme: "http", Host: "example.com", Path: "/service"},
		Header: make(http.Header),
		Body:   &errorReader{},
	}
	
	// Should return error from RoundTrip
	_, err := capture.RoundTrip(req)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to read request body")
}

// errorReader is a test helper that always returns an error on Read
type errorReader struct{}

func (e *errorReader) Read(p []byte) (n int, err error) {
	return 0, io.ErrUnexpectedEOF
}

func (e *errorReader) Close() error {
	return nil
}