// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at http://mozilla.org/MPL/2.0/.

package gowsdl

import (
	"net/http"
	"time"
)

// rateLimitedTransport wraps an http.RoundTripper with rate limiting
type rateLimitedTransport struct {
	transport http.RoundTripper
	limiter   <-chan time.Time
}

// newRateLimitedTransport creates a new rate limited transport
func newRateLimitedTransport(transport http.RoundTripper, ratePerSecond int) http.RoundTripper {
	if ratePerSecond <= 0 {
		return transport
	}

	interval := time.Second / time.Duration(ratePerSecond)
	limiter := time.Tick(interval)

	return &rateLimitedTransport{
		transport: transport,
		limiter:   limiter,
	}
}

// RoundTrip implements the http.RoundTripper interface
func (t *rateLimitedTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	<-t.limiter // Wait for rate limit
	return t.transport.RoundTrip(req)
}