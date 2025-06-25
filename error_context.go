package gowsdl

import (
	"fmt"
	"strings"
)

// ContextualError provides additional context for WSDL parsing errors
type ContextualError struct {
	Operation string // What operation was being performed
	Context   string // Additional context (element name, type, etc.)
	Location  string // File path or URL
	Err       error  // Underlying error
}

func (e *ContextualError) Error() string {
	var parts []string
	
	if e.Operation != "" {
		parts = append(parts, fmt.Sprintf("operation: %s", e.Operation))
	}
	
	if e.Context != "" {
		parts = append(parts, fmt.Sprintf("context: %s", e.Context))
	}
	
	if e.Location != "" {
		parts = append(parts, fmt.Sprintf("location: %s", e.Location))
	}
	
	contextStr := ""
	if len(parts) > 0 {
		contextStr = fmt.Sprintf(" (%s)", strings.Join(parts, ", "))
	}
	
	return fmt.Sprintf("WSDL error%s: %v", contextStr, e.Err)
}

func (e *ContextualError) Unwrap() error {
	return e.Err
}

// newContextualError creates a new contextual error
func newContextualError(operation, context, location string, err error) *ContextualError {
	return &ContextualError{
		Operation: operation,
		Context:   context,
		Location:  location,
		Err:       err,
	}
}

// Common error creation helpers
func (g *GoWSDL) wrapParsingError(elementType, elementName string, err error) error {
	return newContextualError(
		"parse_element",
		fmt.Sprintf("%s '%s'", elementType, elementName),
		g.loc.String(),
		err,
	)
}

func (g *GoWSDL) wrapSchemaError(schemaLocation string, err error) error {
	return newContextualError(
		"resolve_schema",
		fmt.Sprintf("schema '%s'", schemaLocation),
		g.loc.String(),
		err,
	)
}

// Common error messages with suggestions
var commonErrorSuggestions = map[string]string{
	"no such host":           "Check your internet connection and verify the WSDL URL is correct",
	"connection refused":     "The WSDL server may be down or the URL may be incorrect",
	"timeout":               "The WSDL server is taking too long to respond, try again later",
	"invalid character":     "The WSDL file contains invalid XML characters, check the file encoding",
	"unexpected end of XML": "The WSDL file appears to be truncated or corrupted",
	"namespace":             "Check that all XML namespaces are properly declared in the WSDL",
}

// addErrorSuggestion adds helpful suggestions to error messages
func addErrorSuggestion(err error) error {
	errStr := strings.ToLower(err.Error())
	
	for keyword, suggestion := range commonErrorSuggestions {
		if strings.Contains(errStr, keyword) {
			return fmt.Errorf("%w\n\nSuggestion: %s", err, suggestion)
		}
	}
	
	return err
}