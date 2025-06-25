// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at http://mozilla.org/MPL/2.0/.

package soap

import (
	"context"
	"encoding/xml"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Test types for demonstrating generics
type GetUserRequest struct {
	XMLName xml.Name `xml:"getUserRequest"`
	UserID  int      `xml:"userId"`
}

type GetUserResponse struct {
	XMLName xml.Name `xml:"getUserResponse"`
	User    User     `xml:"user"`
}

type User struct {
	ID    int    `xml:"id"`
	Name  string `xml:"name"`
	Email string `xml:"email"`
}

func TestGenericClient(t *testing.T) {
	// Create test server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Return a successful response
		response := `<?xml version="1.0" encoding="UTF-8"?>
		<soap:Envelope xmlns:soap="http://schemas.xmlsoap.org/soap/envelope/">
			<soap:Body>
				<getUserResponse>
					<user>
						<id>123</id>
						<name>John Doe</name>
						<email>john@example.com</email>
					</user>
				</getUserResponse>
			</soap:Body>
		</soap:Envelope>`
		w.Header().Set("Content-Type", "text/xml")
		_, _ = w.Write([]byte(response))
	}))
	defer server.Close()

	// Create generic client
	client := NewGenericClient[GetUserRequest, GetUserResponse](server.URL, "getUser")

	// Make request
	request := GetUserRequest{UserID: 123}
	response, err := client.Call(context.Background(), request)

	require.NoError(t, err)
	assert.NotNil(t, response)
	assert.Equal(t, 123, response.User.ID)
	assert.Equal(t, "John Doe", response.User.Name)
	assert.Equal(t, "john@example.com", response.User.Email)
}

func TestGenericClientWithFault(t *testing.T) {
	// Create test server that returns a SOAP fault
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		response := `<?xml version="1.0" encoding="UTF-8"?>
		<soap:Envelope xmlns:soap="http://schemas.xmlsoap.org/soap/envelope/">
			<soap:Body>
				<soap:Fault>
					<faultcode>Client</faultcode>
					<faultstring>User not found</faultstring>
					<detail>
						<errorCode>404</errorCode>
					</detail>
				</soap:Fault>
			</soap:Body>
		</soap:Envelope>`
		w.Header().Set("Content-Type", "text/xml")
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte(response))
	}))
	defer server.Close()

	// Create generic client
	client := NewGenericClient[GetUserRequest, GetUserResponse](server.URL, "getUser")

	// Make request
	request := GetUserRequest{UserID: 999}
	response, fault, err := client.CallWithFault(context.Background(), request)

	assert.NoError(t, err)
	assert.Nil(t, response)
	assert.NotNil(t, fault)
	assert.Equal(t, "User not found", fault.String)
}

func TestResult(t *testing.T) {
	t.Run("success result", func(t *testing.T) {
		result := Result[User]{
			Value: User{ID: 1, Name: "Test", Email: "test@example.com"},
		}

		assert.True(t, result.IsSuccess())
		assert.False(t, result.TransportError)
		
		value, err := result.Unwrap()
		assert.NoError(t, err)
		assert.Equal(t, 1, value.ID)
	})

	t.Run("fault result", func(t *testing.T) {
		result := Result[User]{
			Fault: &SOAPFault{Code: "Client", String: "Invalid request"},
		}

		assert.False(t, result.IsSuccess())
		assert.False(t, result.TransportError)
		
		_, err := result.Unwrap()
		assert.Error(t, err)
		assert.Equal(t, "Invalid request", err.Error())
	})

	t.Run("transport error result", func(t *testing.T) {
		result := Result[User]{
			Fault: &SOAPFault{Code: "Client.Transport", String: "connection refused"},
			TransportError: true,
		}

		assert.False(t, result.IsSuccess())
		assert.True(t, result.TransportError)
		
		_, err := result.Unwrap()
		assert.Error(t, err)
		assert.Equal(t, "connection refused", err.Error())
	})
}

func TestSOAPArray(t *testing.T) {
	arr := SOAPArray[User]{}

	// Test Add
	arr.Add(User{ID: 1, Name: "User1"})
	arr.Add(User{ID: 2, Name: "User2"})

	assert.Equal(t, 2, arr.Length())

	// Test Get
	user, err := arr.Get(0)
	assert.NoError(t, err)
	assert.Equal(t, "User1", user.Name)

	// Test out of bounds
	_, err = arr.Get(5)
	assert.Error(t, err)
}

func TestBatchClient(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Parse the request to determine which user is being requested
		var req struct {
			XMLName xml.Name `xml:"Envelope"`
			Body    struct {
				GetUserRequest GetUserRequest `xml:"getUserRequest"`
			} `xml:"Body"`
		}
		
		decoder := xml.NewDecoder(r.Body)
		_ = decoder.Decode(&req)
		
		// Return different responses based on user ID
		var response string
		if req.Body.GetUserRequest.UserID == 2 {
			// Return fault for user ID 2
			response = `<?xml version="1.0" encoding="UTF-8"?>
			<soap:Envelope xmlns:soap="http://schemas.xmlsoap.org/soap/envelope/">
				<soap:Body>
					<soap:Fault>
						<faultcode>Server</faultcode>
						<faultstring>Internal error</faultstring>
					</soap:Fault>
				</soap:Body>
			</soap:Envelope>`
			w.WriteHeader(http.StatusInternalServerError)
		} else {
			// Return success for other user IDs
			response = `<?xml version="1.0" encoding="UTF-8"?>
			<soap:Envelope xmlns:soap="http://schemas.xmlsoap.org/soap/envelope/">
				<soap:Body>
					<getUserResponse>
						<user>
							<id>` + fmt.Sprintf("%d", req.Body.GetUserRequest.UserID) + `</id>
							<name>User` + fmt.Sprintf("%d", req.Body.GetUserRequest.UserID) + `</name>
							<email>user` + fmt.Sprintf("%d", req.Body.GetUserRequest.UserID) + `@example.com</email>
						</user>
					</getUserResponse>
				</soap:Body>
			</soap:Envelope>`
		}
		w.Header().Set("Content-Type", "text/xml")
		_, _ = w.Write([]byte(response))
	}))
	defer server.Close()

	// Create batch client
	batchClient := NewBatchClient[GetUserRequest, GetUserResponse](server.URL, "getUser")

	// Prepare requests
	requests := []GetUserRequest{
		{UserID: 1},
		{UserID: 2},
		{UserID: 3},
	}

	// Execute batch
	results := batchClient.CallBatch(context.Background(), requests)

	assert.Len(t, results, 3)

	// Check results - order is preserved in results array
	// First result (user 1 - success)
	assert.Equal(t, 0, results[0].Index)
	assert.True(t, results[0].Result.IsSuccess())
	if results[0].Result.IsSuccess() {
		user, _ := results[0].Result.Unwrap()
		assert.Equal(t, 1, user.User.ID)
	}

	// Second result (user 2 - fault)
	assert.Equal(t, 1, results[1].Index)
	assert.False(t, results[1].Result.IsSuccess())
	assert.NotNil(t, results[1].Result.Fault)
	assert.Equal(t, "Internal error", results[1].Result.Fault.String)

	// Third result (user 3 - success)
	assert.Equal(t, 2, results[2].Index)
	assert.True(t, results[2].Result.IsSuccess())
	if results[2].Result.IsSuccess() {
		user, _ := results[2].Result.Unwrap()
		assert.Equal(t, 3, user.User.ID)
	}
}

func TestBatchClientWithContextCancellation(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Simulate slow response
		time.Sleep(100 * time.Millisecond)
		response := `<?xml version="1.0" encoding="UTF-8"?>
		<soap:Envelope xmlns:soap="http://schemas.xmlsoap.org/soap/envelope/">
			<soap:Body>
				<getUserResponse>
					<user>
						<id>1</id>
						<name>User1</name>
						<email>user1@example.com</email>
					</user>
				</getUserResponse>
			</soap:Body>
		</soap:Envelope>`
		w.Header().Set("Content-Type", "text/xml")
		_, _ = w.Write([]byte(response))
	}))
	defer server.Close()

	// Create batch client
	batchClient := NewBatchClient[GetUserRequest, GetUserResponse](server.URL, "getUser")

	// Create a context that will be cancelled
	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()

	// Prepare multiple requests
	requests := []GetUserRequest{
		{UserID: 1},
		{UserID: 2},
		{UserID: 3},
	}

	// Execute batch
	results := batchClient.CallBatch(ctx, requests)

	assert.Len(t, results, 3)

	// All results should have transport errors due to context cancellation
	for i, result := range results {
		assert.Equal(t, i, result.Index)
		assert.False(t, result.Result.IsSuccess())
		assert.NotNil(t, result.Result.Fault)
		assert.True(t, result.Result.TransportError)
		assert.Contains(t, result.Result.Fault.String, "context deadline exceeded")
	}
}