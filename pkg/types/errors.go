// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at http://mozilla.org/MPL/2.0/.

package types

import "fmt"

// WSDLError represents an error that occurred during WSDL processing
type WSDLError struct {
	Op   string // operation that failed
	Path string // file path or URL
	Err  error  // underlying error
}

func (e *WSDLError) Error() string {
	if e.Path != "" {
		return fmt.Sprintf("wsdl %s %q: %v", e.Op, e.Path, e.Err)
	}
	return fmt.Sprintf("wsdl %s: %v", e.Op, e.Err)
}

func (e *WSDLError) Unwrap() error {
	return e.Err
}

// SchemaError represents an error that occurred during XSD schema processing
type SchemaError struct {
	Op     string // operation that failed
	Schema string // schema location or reference
	Err    error  // underlying error
}

func (e *SchemaError) Error() string {
	if e.Schema != "" {
		return fmt.Sprintf("schema %s %q: %v", e.Op, e.Schema, e.Err)
	}
	return fmt.Sprintf("schema %s: %v", e.Op, e.Err)
}

func (e *SchemaError) Unwrap() error {
	return e.Err
}