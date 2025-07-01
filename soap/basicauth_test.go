package soap

import (
	"context"
	"encoding/xml"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestWithBasicAuth(t *testing.T) {
	// Test request/response types
	type TestRequest struct {
		XMLName xml.Name `xml:"TestRequest"`
		Value   string   `xml:"value"`
	}

	type TestResponse struct {
		XMLName xml.Name `xml:"TestResponse"`
		Result  string   `xml:"result"`
	}

	tests := []struct {
		name           string
		username       string
		password       string
		serverAuth     bool
		expectedStatus int
		expectedError  bool
	}{
		{
			name:           "valid credentials",
			username:       "testuser",
			password:       "testpass",
			serverAuth:     true,
			expectedStatus: http.StatusOK,
			expectedError:  false,
		},
		{
			name:           "invalid credentials",
			username:       "wronguser",
			password:       "wrongpass",
			serverAuth:     true,
			expectedStatus: http.StatusUnauthorized,
			expectedError:  true,
		},
		{
			name:           "empty credentials",
			username:       "",
			password:       "",
			serverAuth:     true,
			expectedStatus: http.StatusUnauthorized,
			expectedError:  true,
		},
		{
			name:           "no auth required by server",
			username:       "anyuser",
			password:       "anypass",
			serverAuth:     false,
			expectedStatus: http.StatusOK,
			expectedError:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create test server
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				// Check if server requires auth
				if tt.serverAuth {
					username, password, ok := r.BasicAuth()
					if !ok || username != "testuser" || password != "testpass" {
						w.WriteHeader(http.StatusUnauthorized)
						return
					}
				}

				// Send response
				w.Header().Set("Content-Type", "text/xml; charset=utf-8")
				w.WriteHeader(tt.expectedStatus)
				if tt.expectedStatus == http.StatusOK {
					// Send a properly formatted SOAP response
					responseXML := `<?xml version="1.0" encoding="UTF-8"?>
<soap:Envelope xmlns:soap="http://schemas.xmlsoap.org/soap/envelope/">
	<soap:Body>
		<TestResponse>
			<result>success</result>
		</TestResponse>
	</soap:Body>
</soap:Envelope>`
					w.Write([]byte(responseXML))
				}
			}))
			defer server.Close()

			// Create SOAP client with basic auth
			client := NewClient(server.URL, WithBasicAuth(tt.username, tt.password))

			// Make request
			request := &TestRequest{Value: "test"}
			response := &TestResponse{}
			
			err := client.CallContext(context.Background(), "", request, response)

			if tt.expectedError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, "success", response.Result)
			}
		})
	}
}

func TestBasicAuthOption(t *testing.T) {
	t.Run("creates basicAuth with correct values", func(t *testing.T) {
		opts := &options{}
		opt := WithBasicAuth("myuser", "mypass")
		opt(opts)

		require.NotNil(t, opts.auth)
		assert.Equal(t, "myuser", opts.auth.Login)
		assert.Equal(t, "mypass", opts.auth.Password)
	})

	t.Run("auth is applied to HTTP request", func(t *testing.T) {
		var capturedAuth string
		
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			auth := r.Header.Get("Authorization")
			capturedAuth = auth
			
			// Send minimal valid SOAP response
			w.Header().Set("Content-Type", "text/xml; charset=utf-8")
			w.Write([]byte(`<?xml version="1.0" encoding="UTF-8"?>
				<soap:Envelope xmlns:soap="http://schemas.xmlsoap.org/soap/envelope/">
					<soap:Body></soap:Body>
				</soap:Envelope>`))
		}))
		defer server.Close()

		client := NewClient(server.URL, WithBasicAuth("testuser", "testpass"))
		
		type EmptyRequest struct{}
		type EmptyResponse struct{}
		
		err := client.CallContext(context.Background(), "", &EmptyRequest{}, &EmptyResponse{})
		require.NoError(t, err)

		// Basic auth header should be: "Basic " + base64("testuser:testpass")
		// base64("testuser:testpass") = "dGVzdHVzZXI6dGVzdHBhc3M="
		assert.Equal(t, "Basic dGVzdHVzZXI6dGVzdHBhc3M=", capturedAuth)
	})
}

func TestBasicAuthIntegration(t *testing.T) {
	t.Run("multiple requests use same auth", func(t *testing.T) {
		requestCount := 0
		authHeaders := []string{}
		
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			requestCount++
			authHeaders = append(authHeaders, r.Header.Get("Authorization"))
			
			w.Header().Set("Content-Type", "text/xml; charset=utf-8")
			w.Write([]byte(`<?xml version="1.0" encoding="UTF-8"?>
				<soap:Envelope xmlns:soap="http://schemas.xmlsoap.org/soap/envelope/">
					<soap:Body></soap:Body>
				</soap:Envelope>`))
		}))
		defer server.Close()

		client := NewClient(server.URL, WithBasicAuth("user", "pass"))
		
		type EmptyRequest struct{}
		type EmptyResponse struct{}
		
		// Make multiple requests
		for i := 0; i < 3; i++ {
			err := client.CallContext(context.Background(), "", &EmptyRequest{}, &EmptyResponse{})
			require.NoError(t, err)
		}
		
		assert.Equal(t, 3, requestCount)
		// All requests should have the same auth header
		expectedAuth := "Basic dXNlcjpwYXNz" // base64("user:pass")
		for _, auth := range authHeaders {
			assert.Equal(t, expectedAuth, auth)
		}
	})

	t.Run("auth works with other options", func(t *testing.T) {
		capturedHeaders := make(map[string]string)
		
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Capture all headers
			capturedHeaders["Authorization"] = r.Header.Get("Authorization")
			capturedHeaders["X-Custom-Header"] = r.Header.Get("X-Custom-Header")
			capturedHeaders["SOAPAction"] = r.Header.Get("SOAPAction")
			
			w.Header().Set("Content-Type", "text/xml; charset=utf-8")
			w.Write([]byte(`<?xml version="1.0" encoding="UTF-8"?>
				<soap:Envelope xmlns:soap="http://schemas.xmlsoap.org/soap/envelope/">
					<soap:Body></soap:Body>
				</soap:Envelope>`))
		}))
		defer server.Close()

		// Create client with multiple options
		client := NewClient(server.URL, 
			WithBasicAuth("myuser", "mypass"),
			WithHTTPHeaders(map[string]string{
				"X-Custom-Header": "custom-value",
			}),
		)
		
		type EmptyRequest struct{}
		type EmptyResponse struct{}
		
		err := client.CallContext(context.Background(), "testAction", &EmptyRequest{}, &EmptyResponse{})
		require.NoError(t, err)

		// Check all headers are present
		assert.Equal(t, "Basic bXl1c2VyOm15cGFzcw==", capturedHeaders["Authorization"]) // base64("myuser:mypass")
		assert.Equal(t, "custom-value", capturedHeaders["X-Custom-Header"])
		assert.Equal(t, "testAction", capturedHeaders["SOAPAction"])
	})
}

func TestBasicAuthWithSpecialCharacters(t *testing.T) {
	tests := []struct {
		name     string
		username string
		password string
		expected string
	}{
		{
			name:     "username with colon",
			username: "user:name",
			password: "pass",
			expected: "Basic dXNlcjpuYW1lOnBhc3M=", // base64("user:name:pass")
		},
		{
			name:     "password with special chars",
			username: "user",
			password: "p@$$w0rd!",
			expected: "Basic dXNlcjpwQCQkdzByZCE=", // base64("user:p@$$w0rd!")
		},
		{
			name:     "empty username",
			username: "",
			password: "pass",
			expected: "Basic OnBhc3M=", // base64(":pass")
		},
		{
			name:     "empty password",
			username: "user",
			password: "",
			expected: "Basic dXNlcjo=", // base64("user:")
		},
		{
			name:     "unicode characters",
			username: "用户",
			password: "密码",
			expected: "Basic 55So5oi3OuWvhueggQ==", // base64("用户:密码")
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var capturedAuth string
			
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				capturedAuth = r.Header.Get("Authorization")
				
				w.Header().Set("Content-Type", "text/xml; charset=utf-8")
				w.Write([]byte(`<?xml version="1.0" encoding="UTF-8"?>
					<soap:Envelope xmlns:soap="http://schemas.xmlsoap.org/soap/envelope/">
						<soap:Body></soap:Body>
					</soap:Envelope>`))
			}))
			defer server.Close()

			client := NewClient(server.URL, WithBasicAuth(tt.username, tt.password))
			
			type EmptyRequest struct{}
			type EmptyResponse struct{}
			
			err := client.CallContext(context.Background(), "", &EmptyRequest{}, &EmptyResponse{})
			require.NoError(t, err)

			assert.Equal(t, tt.expected, capturedAuth)
		})
	}
}

func TestBasicAuthErrorHandling(t *testing.T) {
	t.Run("server returns 401", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusUnauthorized)
			w.Write([]byte("Unauthorized"))
		}))
		defer server.Close()

		client := NewClient(server.URL, WithBasicAuth("wrong", "creds"))
		
		type EmptyRequest struct{}
		type EmptyResponse struct{}
		
		err := client.CallContext(context.Background(), "", &EmptyRequest{}, &EmptyResponse{})
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "401")
	})

	t.Run("nil auth not applied", func(t *testing.T) {
		var capturedAuth string
		hasAuth := false
		
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			capturedAuth = r.Header.Get("Authorization")
			hasAuth = capturedAuth != ""
			
			w.Header().Set("Content-Type", "text/xml; charset=utf-8")
			w.Write([]byte(`<?xml version="1.0" encoding="UTF-8"?>
				<soap:Envelope xmlns:soap="http://schemas.xmlsoap.org/soap/envelope/">
					<soap:Body></soap:Body>
				</soap:Envelope>`))
		}))
		defer server.Close()

		// Create client without auth
		client := NewClient(server.URL)
		
		type EmptyRequest struct{}
		type EmptyResponse struct{}
		
		err := client.CallContext(context.Background(), "", &EmptyRequest{}, &EmptyResponse{})
		require.NoError(t, err)

		assert.False(t, hasAuth)
		assert.Empty(t, capturedAuth)
	})
}