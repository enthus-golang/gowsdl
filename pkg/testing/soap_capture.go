// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at http://mozilla.org/MPL/2.0/.

package testing

import (
	"bytes"
	"encoding/xml"
	"fmt"
	"io"
	"net/http"
	"strings"
	"sync"
)

// SOAPCapture intercepts HTTP requests and captures SOAP XML for testing
type SOAPCapture struct {
	mu        sync.Mutex
	Requests  []CapturedRequest
	Responses []CapturedResponse
}

// CapturedRequest represents a captured SOAP request
type CapturedRequest struct {
	URL     string
	Headers map[string]string
	Body    string
	XMLDoc  *XMLDocument
}

// CapturedResponse represents a captured SOAP response  
type CapturedResponse struct {
	StatusCode int
	Headers    map[string]string
	Body       string
	XMLDoc     *XMLDocument
}

// XMLDocument represents parsed XML for semantic comparison
type XMLDocument struct {
	Envelope *SOAPEnvelope
	Raw      string
}

// SOAPEnvelope represents the basic SOAP envelope structure
type SOAPEnvelope struct {
	XMLName xml.Name `xml:"Envelope"`
	Body    SOAPBody `xml:"Body"`
}

// SOAPBody represents the SOAP body content
type SOAPBody struct {
	XMLName xml.Name `xml:"Body"`
	Content []byte   `xml:",innerxml"`
}

// NewSOAPCapture creates a new SOAP capture instance
func NewSOAPCapture() *SOAPCapture {
	return &SOAPCapture{
		Requests:  make([]CapturedRequest, 0),
		Responses: make([]CapturedResponse, 0),
	}
}

// RoundTrip implements http.RoundTripper to capture requests
func (sc *SOAPCapture) RoundTrip(req *http.Request) (*http.Response, error) {
	// Capture request
	captured := CapturedRequest{
		URL:     req.URL.String(),
		Headers: make(map[string]string),
	}
	
	// Copy headers
	for key, values := range req.Header {
		captured.Headers[key] = strings.Join(values, ", ")
	}
	
	// Read and capture request body
	if req.Body != nil {
		bodyBytes, err := io.ReadAll(req.Body)
		if err != nil {
			return nil, fmt.Errorf("failed to read request body: %w", err)
		}
		captured.Body = string(bodyBytes)
		
		// Parse XML
		if xmlDoc, err := sc.parseXML(captured.Body); err == nil {
			captured.XMLDoc = xmlDoc
		}
		
		// Restore request body for actual use
		req.Body = io.NopCloser(bytes.NewReader(bodyBytes))
	}
	
	sc.mu.Lock()
	sc.Requests = append(sc.Requests, captured)
	sc.mu.Unlock()
	
	// Return mock response for testing
	mockResponse := &http.Response{
		StatusCode: 200,
		Header:     make(http.Header),
		Body:       io.NopCloser(strings.NewReader(`<?xml version="1.0" encoding="UTF-8"?><soap:Envelope xmlns:soap="http://schemas.xmlsoap.org/soap/envelope/"><soap:Body><response>Mock Response</response></soap:Body></soap:Envelope>`)),
	}
	mockResponse.Header.Set("Content-Type", "text/xml; charset=utf-8")
	
	return mockResponse, nil
}

// parseXML parses SOAP XML into structured format
func (sc *SOAPCapture) parseXML(xmlData string) (*XMLDocument, error) {
	var envelope SOAPEnvelope
	if err := xml.Unmarshal([]byte(xmlData), &envelope); err != nil {
		return nil, fmt.Errorf("failed to parse XML: %w", err)
	}
	
	return &XMLDocument{
		Envelope: &envelope,
		Raw:      xmlData,
	}, nil
}

// GetLastRequest returns the most recent captured request
func (sc *SOAPCapture) GetLastRequest() *CapturedRequest {
	sc.mu.Lock()
	defer sc.mu.Unlock()
	
	if len(sc.Requests) == 0 {
		return nil
	}
	return &sc.Requests[len(sc.Requests)-1]
}

// GetRequestCount returns the number of captured requests
func (sc *SOAPCapture) GetRequestCount() int {
	sc.mu.Lock()
	defer sc.mu.Unlock()
	return len(sc.Requests)
}

// Reset clears all captured requests and responses
func (sc *SOAPCapture) Reset() {
	sc.mu.Lock()
	defer sc.mu.Unlock()
	sc.Requests = sc.Requests[:0]
	sc.Responses = sc.Responses[:0]
}