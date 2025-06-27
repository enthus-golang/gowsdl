// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at http://mozilla.org/MPL/2.0/.

package http

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"net"
	"net/http"
	"net/url"
	"os"
	"time"
)

// HTTPClientConfig holds configuration for the HTTP client used to download WSDL and XSD files
type HTTPClientConfig struct {
	// Timeout for HTTP requests (default: 30s)
	Timeout time.Duration

	// MaxRetries for failed requests (default: 0)
	MaxRetries int

	// RetryDelay between retries (default: 1s)
	RetryDelay time.Duration

	// RateLimitPerSecond limits requests per second (0 = no limit)
	RateLimitPerSecond int

	// MaxResponseSize limits the response size in bytes (default: 10MB)
	MaxResponseSize int64

	// UserAgent to use for requests (default: "gowsdl/1.0")
	UserAgent string

	// TLSConfig for custom TLS settings
	TLSConfig *tls.Config

	// InsecureSkipVerify skips TLS verification (deprecated, use TLSConfig)
	InsecureSkipVerify bool

	// ProxyURL for HTTP proxy
	ProxyURL string
}

// DefaultHTTPClientConfig returns a default HTTP client configuration
func DefaultHTTPClientConfig() *HTTPClientConfig {
	return &HTTPClientConfig{
		Timeout:         30 * time.Second,
		MaxRetries:      0,
		RetryDelay:      1 * time.Second,
		MaxResponseSize: 10 * 1024 * 1024, // 10MB
		UserAgent:       "gowsdl/1.0",
		TLSConfig: &tls.Config{
			MinVersion: tls.VersionTLS12,
		},
	}
}

// WithTimeout sets the HTTP timeout
func (c *HTTPClientConfig) WithTimeout(timeout time.Duration) *HTTPClientConfig {
	c.Timeout = timeout
	return c
}

// WithRetries sets the retry configuration
func (c *HTTPClientConfig) WithRetries(maxRetries int, delay time.Duration) *HTTPClientConfig {
	c.MaxRetries = maxRetries
	c.RetryDelay = delay
	return c
}

// WithTLSConfig sets a custom TLS configuration
func (c *HTTPClientConfig) WithTLSConfig(tlsConfig *tls.Config) *HTTPClientConfig {
	c.TLSConfig = tlsConfig
	return c
}

// WithCACert adds a CA certificate from file
func (c *HTTPClientConfig) WithCACert(certFile string) (*HTTPClientConfig, error) {
	caCert, err := os.ReadFile(certFile)
	if err != nil {
		return c, fmt.Errorf("failed to read CA certificate: %w", err)
	}

	if c.TLSConfig == nil {
		c.TLSConfig = &tls.Config{
			MinVersion: tls.VersionTLS12,
		}
	}

	if c.TLSConfig.RootCAs == nil {
		c.TLSConfig.RootCAs = x509.NewCertPool()
	}

	if ok := c.TLSConfig.RootCAs.AppendCertsFromPEM(caCert); !ok {
		return c, fmt.Errorf("failed to parse CA certificate")
	}

	return c, nil
}

// WithClientCert adds a client certificate for mutual TLS
func (c *HTTPClientConfig) WithClientCert(certFile, keyFile string) (*HTTPClientConfig, error) {
	cert, err := tls.LoadX509KeyPair(certFile, keyFile)
	if err != nil {
		return c, fmt.Errorf("failed to load client certificate: %w", err)
	}

	if c.TLSConfig == nil {
		c.TLSConfig = &tls.Config{
			MinVersion: tls.VersionTLS12,
		}
	}

	c.TLSConfig.Certificates = append(c.TLSConfig.Certificates, cert)
	return c, nil
}

// Build creates an HTTP client from the configuration
func (c *HTTPClientConfig) Build() *http.Client {
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
		transport.Proxy = http.ProxyURL(mustParseURL(c.ProxyURL))
	}

	// Wrap with rate limiter if configured
	var finalTransport http.RoundTripper = transport
	if c.RateLimitPerSecond > 0 {
		finalTransport = newRateLimitedTransport(transport, c.RateLimitPerSecond)
	}

	return &http.Client{
		Transport: finalTransport,
		Timeout:   c.Timeout,
	}
}

func mustParseURL(rawURL string) *url.URL {
	u, err := url.Parse(rawURL)
	if err != nil {
		panic(fmt.Sprintf("invalid proxy URL: %v", err))
	}
	return u
}

// GetUserAgent returns the user agent
func (c *HTTPClientConfig) GetUserAgent() string {
	return c.UserAgent
}

// GetMaxRetries returns the maximum number of retries
func (c *HTTPClientConfig) GetMaxRetries() int {
	return c.MaxRetries
}

// GetRetryDelay returns the delay between retries
func (c *HTTPClientConfig) GetRetryDelay() time.Duration {
	return c.RetryDelay
}

// GetMaxResponseSize returns the maximum response size
func (c *HTTPClientConfig) GetMaxResponseSize() int64 {
	return c.MaxResponseSize
}