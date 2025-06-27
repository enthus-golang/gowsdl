// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at http://mozilla.org/MPL/2.0/.

package http

import (
	"fmt"
	"net/url"
	"os"
	"path/filepath"
)

// A Location encapsulate information about the loc of WSDL/XSD.
//
// It could be either URL or an absolute file path.
type Location struct {
	u *url.URL
	f string
}

// ParseLocation parses a rawloc into a Location structure.
//
// If rawloc is URL then it should be absolute.
// Relative file path will be converted into absolute path.
func ParseLocation(rawloc string) (*Location, error) {
	u, _ := url.Parse(rawloc)
	if u.Scheme != "" {
		return &Location{u: u}, nil
	}

	absURI, err := filepath.Abs(rawloc)
	if err != nil {
		return nil, err
	}

	return &Location{f: absURI}, nil
}

// Parse parses path in the context of the receiver. The provided path may be relative or absolute.
// Parse returns nil, err on parse failure.
func (r *Location) Parse(ref string) (*Location, error) {
	if r.u != nil {
		u, err := r.u.Parse(ref)
		if err != nil {
			return nil, err
		}
		return &Location{u: u}, nil
	}

	if filepath.IsAbs(ref) {
		return &Location{f: ref}, nil
	}

	if u, err := url.Parse(ref); err == nil {
		if u.Scheme != "" {
			return &Location{u: u}, nil
		}
	}

	return &Location{f: filepath.Join(filepath.Dir(r.f), ref)}, nil
}

// IsLocal determines whether the Location contains a file path.
func (r *Location) IsLocal() bool {
	return r.f != ""
}

// IsFile determines whether the Location contains a file path.
func (r *Location) isFile() bool {
	return r.f != ""
}

// IsFile determines whether the Location contains URL.
func (r *Location) isURL() bool {
	return r.u != nil
}

// URL returns the URL string if this is a URL location
func (r *Location) URL() string {
	if r.u != nil {
		return r.u.String()
	}
	return ""
}

// Path returns the file path if this is a file location
func (r *Location) Path() string {
	return r.f
}

// ReadFile reads the file if it's local
func (r *Location) ReadFile() ([]byte, error) {
	if !r.IsLocal() {
		return nil, fmt.Errorf("cannot read remote file: %s", r.URL())
	}
	return os.ReadFile(r.f)
}

// String reassembles the Location either into a valid URL string or a file path.
func (r *Location) String() string {
	if r.isFile() {
		return r.f
	}
	if r.isURL() {
		return r.u.String()
	}
	return ""
}
