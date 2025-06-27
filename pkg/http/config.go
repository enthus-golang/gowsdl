// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at http://mozilla.org/MPL/2.0/.

package http

import (
	"crypto/tls"
	"net"
	"net/http"
	"net/url"
	"time"
)

// Config holds configuration for the HTTP client
type Config struct {
	Timeout            time.Duration
	MaxRetries         int
	RetryDelay         time.Duration
	RateLimitPerSecond int
	MaxResponseSize    int64
	UserAgent          string
	TLSConfig          *tls.Config
	InsecureSkipVerify bool
	ProxyURL           string
}

// Build creates an HTTP client from the configuration
func (c *Config) Build() *http.Client {
	// Handle deprecated InsecureSkipVerify
	if c.InsecureSkipVerify && c.TLSConfig != nil {
		c.TLSConfig.InsecureSkipVerify = true
	}

	transport := &http.Transport{
		TLSClientConfig: c.TLSConfig,
		DialContext: (&net.Dialer{
			Timeout:   30 * time.Second,
			KeepAlive: 30 * time.Second,
		}).DialContext,
		MaxIdleConns:          100,
		IdleConnTimeout:       90 * time.Second,
		TLSHandshakeTimeout:   10 * time.Second,
		ExpectContinueTimeout: 1 * time.Second,
	}

	if c.ProxyURL != "" {
		proxyURL, _ := url.Parse(c.ProxyURL)
		transport.Proxy = http.ProxyURL(proxyURL)
	}

	return &http.Client{
		Transport: transport,
		Timeout:   c.Timeout,
	}
}