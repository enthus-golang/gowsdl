// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at http://mozilla.org/MPL/2.0/.

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
		
		value, err := result.Unwrap()
		assert.NoError(t, err)
		assert.Equal(t, 1, value.ID)
	})

	t.Run("fault result", func(t *testing.T) {
		result := Result[User]{
			Fault: &SOAPFault{Code: "Client", String: "Invalid request"},
		}

		assert.False(t, result.IsSuccess())
		
		_, err := result.Unwrap()
		assert.Error(t, err)
		assert.Equal(t, "Invalid request", err.Error())
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
	callCount := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		callCount++
		// Return different responses based on call count
		var response string
		if callCount == 2 {
			// Return fault for second request
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
			response = `<?xml version="1.0" encoding="UTF-8"?>
			<soap:Envelope xmlns:soap="http://schemas.xmlsoap.org/soap/envelope/">
				<soap:Body>
					<getUserResponse>
						<user>
							<id>` + string(rune('0'+callCount)) + `</id>
							<name>User` + string(rune('0'+callCount)) + `</name>
							<email>user` + string(rune('0'+callCount)) + `@example.com</email>
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

	// Check first result (success)
	assert.Equal(t, 0, results[0].Index)
	assert.True(t, results[0].Result.IsSuccess())

	// Check second result (fault)
	assert.Equal(t, 1, results[1].Index)
	assert.False(t, results[1].Result.IsSuccess())
	assert.NotNil(t, results[1].Result.Fault)

	// Check third result (success)
	assert.Equal(t, 2, results[2].Index)
	assert.True(t, results[2].Result.IsSuccess())
}