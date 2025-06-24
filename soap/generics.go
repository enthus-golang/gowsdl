// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at http://mozilla.org/MPL/2.0/.

package soap

import (
	"context"
	"encoding/xml"
	"fmt"
	"io"
)

// GenericClient is a type-safe SOAP client using Go generics
type GenericClient[TReq any, TResp any] struct {
	client *Client
	action string
}

// NewGenericClient creates a new generic SOAP client
func NewGenericClient[TReq any, TResp any](url string, action string, opts ...Option) *GenericClient[TReq, TResp] {
	return &GenericClient[TReq, TResp]{
		client: NewClient(url, opts...),
		action: action,
	}
}

// Call executes a SOAP request with type safety
func (c *GenericClient[TReq, TResp]) Call(ctx context.Context, request TReq) (*TResp, error) {
	var response TResp
	err := c.client.CallContext(ctx, c.action, request, &response)
	if err != nil {
		return nil, err
	}
	return &response, nil
}

// CallWithFault executes a SOAP request and returns both response and fault
func (c *GenericClient[TReq, TResp]) CallWithFault(ctx context.Context, request TReq) (*TResp, *SOAPFault, error) {
	var response TResp
	err := c.client.CallContext(ctx, c.action, request, &response)
	
	// Check if error is a SOAP fault
	if err != nil {
		if fault, ok := err.(*SOAPFault); ok {
			return nil, fault, nil
		}
		return nil, nil, err
	}
	
	return &response, nil, nil
}

// Result represents a SOAP call result with either a value or a fault
type Result[T any] struct {
	Value T
	Fault *SOAPFault
}

// Unwrap returns the value or an error if a fault occurred
func (r Result[T]) Unwrap() (T, error) {
	var zero T
	if r.Fault != nil {
		return zero, r.Fault
	}
	return r.Value, nil
}

// IsSuccess returns true if the result contains a value (no fault)
func (r Result[T]) IsSuccess() bool {
	return r.Fault == nil
}

// CallAsResult executes a SOAP request and returns a Result
func (c *GenericClient[TReq, TResp]) CallAsResult(ctx context.Context, request TReq) Result[TResp] {
	value, fault, err := c.CallWithFault(ctx, request)
	if err != nil {
		// Convert non-SOAP errors to faults
		return Result[TResp]{
			Fault: &SOAPFault{
				Code:   "Client",
				String: err.Error(),
			},
		}
	}
	if fault != nil {
		return Result[TResp]{Fault: fault}
	}
	return Result[TResp]{Value: *value}
}

// SOAPArray is a generic container for SOAP array types
type SOAPArray[T any] struct {
	Items []T `xml:"item"`
}

// Add appends an item to the array
func (a *SOAPArray[T]) Add(item T) {
	a.Items = append(a.Items, item)
}

// Get returns the item at the specified index
func (a *SOAPArray[T]) Get(index int) (T, error) {
	var zero T
	if index < 0 || index >= len(a.Items) {
		return zero, fmt.Errorf("index %d out of bounds", index)
	}
	return a.Items[index], nil
}

// Length returns the number of items in the array
func (a *SOAPArray[T]) Length() int {
	return len(a.Items)
}

// GenericEnvelope is a type-safe SOAP envelope
type GenericEnvelope[T any] struct {
	XMLName xml.Name      `xml:"http://schemas.xmlsoap.org/soap/envelope/ Envelope"`
	Headers []interface{} `xml:"http://schemas.xmlsoap.org/soap/envelope/ Header,omitempty"`
	Body    GenericBody[T]
}

// GenericBody is a type-safe SOAP body
type GenericBody[T any] struct {
	XMLName xml.Name `xml:"http://schemas.xmlsoap.org/soap/envelope/ Body"`
	Content T        `xml:",omitempty"`
	Fault   *SOAPFault
}

// BatchClient allows executing multiple SOAP calls concurrently
type BatchClient[TReq any, TResp any] struct {
	client *GenericClient[TReq, TResp]
}

// NewBatchClient creates a new batch client
func NewBatchClient[TReq any, TResp any](url string, action string, opts ...Option) *BatchClient[TReq, TResp] {
	return &BatchClient[TReq, TResp]{
		client: NewGenericClient[TReq, TResp](url, action, opts...),
	}
}

// BatchResult represents the result of a batch operation
type BatchResult[T any] struct {
	Index  int
	Result Result[T]
}

// CallBatch executes multiple SOAP requests concurrently
func (b *BatchClient[TReq, TResp]) CallBatch(ctx context.Context, requests []TReq) []BatchResult[TResp] {
	results := make([]BatchResult[TResp], len(requests))
	ch := make(chan BatchResult[TResp], len(requests))
	
	for i, req := range requests {
		go func(index int, request TReq) {
			result := b.client.CallAsResult(ctx, request)
			ch <- BatchResult[TResp]{Index: index, Result: result}
		}(i, req)
	}
	
	for i := 0; i < len(requests); i++ {
		result := <-ch
		results[result.Index] = result
	}
	
	return results
}

// TypedSOAPDecoder provides type-safe XML decoding for SOAP responses
type TypedSOAPDecoder[T any] struct {
	decoder *xml.Decoder
}

// NewTypedSOAPDecoder creates a new typed decoder
func NewTypedSOAPDecoder[T any](r io.Reader) *TypedSOAPDecoder[T] {
	return &TypedSOAPDecoder[T]{
		decoder: xml.NewDecoder(r),
	}
}

// Decode decodes a SOAP response into the specified type
func (d *TypedSOAPDecoder[T]) Decode() (*T, error) {
	var envelope GenericEnvelope[T]
	if err := d.decoder.Decode(&envelope); err != nil {
		return nil, err
	}
	
	if envelope.Body.Fault != nil {
		return nil, envelope.Body.Fault
	}
	
	return &envelope.Body.Content, nil
}