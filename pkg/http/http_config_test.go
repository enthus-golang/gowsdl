// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at http://mozilla.org/MPL/2.0/.

package http

import (
	"crypto/tls"
	"testing"
	"time"
)

func TestDefaultHTTPClientConfig(t *testing.T) {
	config := DefaultHTTPClientConfig()

	if config.Timeout != 30*time.Second {
		t.Errorf("Expected timeout 30s, got %v", config.Timeout)
	}

	if config.MaxRetries != 0 {
		t.Errorf("Expected max retries 0, got %d", config.MaxRetries)
	}

	if config.MaxResponseSize != 10*1024*1024 {
		t.Errorf("Expected max response size 10MB, got %d", config.MaxResponseSize)
	}

	if config.UserAgent != "gowsdl/1.0" {
		t.Errorf("Expected user agent 'gowsdl/1.0', got %s", config.UserAgent)
	}

	if config.TLSConfig.MinVersion != tls.VersionTLS12 {
		t.Errorf("Expected TLS 1.2 minimum, got %d", config.TLSConfig.MinVersion)
	}
}

func TestHTTPClientConfigWithTimeout(t *testing.T) {
	config := DefaultHTTPClientConfig().WithTimeout(60 * time.Second)

	if config.Timeout != 60*time.Second {
		t.Errorf("Expected timeout 60s, got %v", config.Timeout)
	}
}

func TestHTTPClientConfigWithRetries(t *testing.T) {
	config := DefaultHTTPClientConfig().WithRetries(3, 2*time.Second)

	if config.MaxRetries != 3 {
		t.Errorf("Expected max retries 3, got %d", config.MaxRetries)
	}

	if config.RetryDelay != 2*time.Second {
		t.Errorf("Expected retry delay 2s, got %v", config.RetryDelay)
	}
}

func TestHTTPClientConfigBuild(t *testing.T) {
	config := DefaultHTTPClientConfig()
	client := config.Build()

	if client == nil {
		t.Fatal("Expected HTTP client to be built")
	}

	if client.Timeout != config.Timeout {
		t.Errorf("Expected client timeout %v, got %v", config.Timeout, client.Timeout)
	}
}

