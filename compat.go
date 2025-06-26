// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at http://mozilla.org/MPL/2.0/.

package gowsdl

import (
	"crypto/tls"
	"time"

	"github.com/hooklift/gowsdl/pkg/core"
	"github.com/hooklift/gowsdl/pkg/http"
	"github.com/hooklift/gowsdl/pkg/parser"
	"github.com/hooklift/gowsdl/pkg/types"
)

// Re-export types for backward compatibility
type (
	// Error types
	WSDLError   = types.WSDLError
	SchemaError = types.SchemaError
	
	// Parser types
	WSDL            = parser.WSDL
	WSDL2           = parser.WSDL2
	XSDSchema       = parser.XSDSchema
	XSDComplexType  = parser.XSDComplexType
	XSDElement      = parser.XSDElement
	XSDAttribute    = parser.XSDAttribute
	WSDLBinding     = parser.WSDLBinding
	WSDLOperation   = parser.WSDLOperation
	WSDLService     = parser.WSDLService
	WSDLPort        = parser.WSDLPort
	WSDLImport      = parser.WSDLImport
	XSDImport       = parser.XSDImport
	
	// Core types
	NamespaceManager = core.NamespaceManager
	Location         = http.Location
)

// Re-export functions for backward compatibility
// HTTPClientConfig is an alias for backward compatibility
type HTTPClientConfig = http.HTTPClientConfig

// Re-export functions for backward compatibility
var (
	NewNamespaceManager = core.NewNamespaceManager
	ParseLocation       = http.ParseLocation
	DefaultHTTPClientConfig = func() *HTTPClientConfig {
		return &HTTPClientConfig{
			Timeout:         30 * time.Second,
			MaxRetries:      0,
			RetryDelay:      1 * time.Second,
			MaxResponseSize: 10 * 1024 * 1024,
			UserAgent:       "gowsdl/1.0",
			TLSConfig: &tls.Config{
				MinVersion: tls.VersionTLS12,
			},
		}
	}
)