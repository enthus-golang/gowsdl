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

