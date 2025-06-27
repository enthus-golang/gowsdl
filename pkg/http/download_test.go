// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at http://mozilla.org/MPL/2.0/.

package http

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/hooklift/gowsdl/pkg/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDownloadFile(t *testing.T) {
	tests := []struct {
		name           string
		serverResponse func(w http.ResponseWriter, r *http.Request)
		config         *HTTPClientConfig
		wantErr        bool
		errContains    string
	}{
		{
			name: "successful download",
			serverResponse: func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
				_, _ = w.Write([]byte("test content"))
			},
			config:  DefaultHTTPClientConfig(),
			wantErr: false,
		},
		{
			name: "404 not found",
			serverResponse: func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusNotFound)
			},
			config:      DefaultHTTPClientConfig(),
			wantErr:     true,
			errContains: "received HTTP 404",
		},
		{
			name: "500 server error",
			serverResponse: func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusInternalServerError)
			},
			config:      DefaultHTTPClientConfig(),
			wantErr:     true,
			errContains: "received HTTP 500",
		},
		{
			name: "timeout",
			serverResponse: func(w http.ResponseWriter, r *http.Request) {
				time.Sleep(100 * time.Millisecond)
				w.WriteHeader(http.StatusOK)
			},
			config: func() *HTTPClientConfig {
				cfg := DefaultHTTPClientConfig()
				cfg.Timeout = 10 * time.Millisecond
				return cfg
			}(),
			wantErr:     true,
			errContains: "context deadline exceeded",
		},
		{
			name: "retry on server error",
			serverResponse: func() func(w http.ResponseWriter, r *http.Request) {
				count := 0
				return func(w http.ResponseWriter, r *http.Request) {
					count++
					if count < 3 {
						w.WriteHeader(http.StatusInternalServerError)
					} else {
						w.WriteHeader(http.StatusOK)
						_, _ = w.Write([]byte("success after retry"))
					}
				}
			}(),
			config: func() *HTTPClientConfig {
				cfg := DefaultHTTPClientConfig()
				cfg.MaxRetries = 3
				cfg.RetryDelay = 1 * time.Millisecond
				return cfg
			}(),
			wantErr: false,
		},
		{
			name: "max retries exceeded",
			serverResponse: func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusInternalServerError)
			},
			config: func() *HTTPClientConfig {
				cfg := DefaultHTTPClientConfig()
				cfg.MaxRetries = 2
				cfg.RetryDelay = 1 * time.Millisecond
				return cfg
			}(),
			wantErr:     true,
			errContains: "received HTTP 500",
		},
		{
			name: "custom user agent",
			serverResponse: func(w http.ResponseWriter, r *http.Request) {
				assert.Equal(t, "test-agent/1.0", r.Header.Get("User-Agent"))
				w.WriteHeader(http.StatusOK)
			},
			config: func() *HTTPClientConfig {
				cfg := DefaultHTTPClientConfig()
				cfg.UserAgent = "test-agent/1.0"
				return cfg
			}(),
			wantErr: false,
		},
		{
			name: "response size limit",
			serverResponse: func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
				// Write more than the limit
				data := make([]byte, 1024)
				for i := 0; i < 10; i++ {
					_, _ = w.Write(data)
				}
			},
			config: func() *HTTPClientConfig {
				cfg := DefaultHTTPClientConfig()
				cfg.MaxResponseSize = 1024
				return cfg
			}(),
			wantErr: false, // Should not error, just truncate
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(tt.serverResponse))
			defer server.Close()

			ctx := context.Background()
			data, err := DownloadFile(ctx, server.URL, tt.config)

			if tt.wantErr {
				require.Error(t, err)
				if tt.errContains != "" {
					assert.Contains(t, err.Error(), tt.errContains)
				}
			} else {
				require.NoError(t, err)
				assert.NotNil(t, data)
			}
		})
	}
}

func TestDownloadFileWithContext(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Simulate slow response
		time.Sleep(100 * time.Millisecond)
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("test"))
	}))
	defer server.Close()

	// Test with cancelled context
	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	_, err := DownloadFile(ctx, server.URL, DefaultHTTPClientConfig())
	require.Error(t, err)
	assert.Contains(t, err.Error(), "context canceled")
}


func TestDownloadFileRateLimit(t *testing.T) {
	requestTimes := make([]time.Time, 0)
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestTimes = append(requestTimes, time.Now())
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok"))
	}))
	defer server.Close()

	config := DefaultHTTPClientConfig()
	config.RateLimitPerSecond = 2 // Allow 2 requests per second

	// Make 3 requests
	for i := 0; i < 3; i++ {
		_, err := DownloadFile(context.Background(), server.URL, config)
		require.NoError(t, err)
	}

	// Check that requests were rate limited
	assert.Len(t, requestTimes, 3)
	if len(requestTimes) >= 3 {
		// The third request should have been delayed
		elapsed := requestTimes[2].Sub(requestTimes[0])
		assert.GreaterOrEqual(t, elapsed, 1*time.Second)
	}
}

func BenchmarkDownloadFile(b *testing.B) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`<?xml version="1.0" encoding="UTF-8"?>
<definitions xmlns="http://schemas.xmlsoap.org/wsdl/">
	<types>
		<schema xmlns="http://www.w3.org/2001/XMLSchema">
			<element name="TestElement" type="string"/>
		</schema>
	</types>
</definitions>`))
	}))
	defer server.Close()

	config := DefaultHTTPClientConfig()
	ctx := context.Background()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := DownloadFile(ctx, server.URL, config)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func TestWSDLErrorInDownload(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusForbidden)
	}))
	defer server.Close()

	config := DefaultHTTPClientConfig()
	_, err := DownloadFile(context.Background(), server.URL, config)
	
	require.Error(t, err)
	var wsdlErr *types.WSDLError
	assert.ErrorAs(t, err, &wsdlErr)
	assert.Equal(t, "download", wsdlErr.Op)
	assert.Contains(t, wsdlErr.Path, server.URL)
}