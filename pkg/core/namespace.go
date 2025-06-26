package core

import (
	"fmt"
	"strings"
	"sync"
)

// NamespaceManager handles XML namespace prefix resolution and management
type NamespaceManager struct {
	mu         sync.RWMutex
	namespaces map[string]string // prefix -> URI
	prefixes   map[string]string // URI -> prefix
	counter    int               // for generating unique prefixes
}

// NewNamespaceManager creates a new namespace manager
func NewNamespaceManager() *NamespaceManager {
	nm := &NamespaceManager{
		namespaces: make(map[string]string),
		prefixes:   make(map[string]string),
	}
	// Add common namespaces
	nm.AddNamespace("soap", "http://schemas.xmlsoap.org/wsdl/soap/")
	nm.AddNamespace("soap12", "http://schemas.xmlsoap.org/wsdl/soap12/")
	nm.AddNamespace("http", "http://schemas.xmlsoap.org/wsdl/http/")
	nm.AddNamespace("mime", "http://schemas.xmlsoap.org/wsdl/mime/")
	nm.AddNamespace("xsd", "http://www.w3.org/2001/XMLSchema")
	nm.AddNamespace("wsdl", "http://schemas.xmlsoap.org/wsdl/")
	nm.AddNamespace("wsdl2", "http://www.w3.org/ns/wsdl/")
	return nm
}

// AddNamespace registers a namespace with a prefix
func (nm *NamespaceManager) AddNamespace(prefix, uri string) {
	nm.mu.Lock()
	defer nm.mu.Unlock()
	
	// Check if URI already has a prefix
	if existingPrefix, exists := nm.prefixes[uri]; exists && existingPrefix != prefix {
		// Namespace collision - generate unique prefix
		prefix = nm.generateUniquePrefix(prefix)
	}
	
	nm.namespaces[prefix] = uri
	nm.prefixes[uri] = prefix
}

// GetNamespaceURI returns the namespace URI for a given prefix
func (nm *NamespaceManager) GetNamespaceURI(prefix string) (string, bool) {
	nm.mu.RLock()
	defer nm.mu.RUnlock()
	uri, ok := nm.namespaces[prefix]
	return uri, ok
}

// GetPrefix returns the prefix for a given namespace URI
func (nm *NamespaceManager) GetPrefix(uri string) (string, bool) {
	nm.mu.RLock()
	defer nm.mu.RUnlock()
	prefix, ok := nm.prefixes[uri]
	return prefix, ok
}

// RegisterNamespaces registers multiple namespaces from a map
func (nm *NamespaceManager) RegisterNamespaces(xmlns map[string]string) {
	for prefix, uri := range xmlns {
		// Skip default namespace (empty prefix)
		if prefix == "" {
			continue
		}
		nm.AddNamespace(prefix, uri)
	}
}

// ResolveQName resolves a qualified name (prefix:localName) to its full form
func (nm *NamespaceManager) ResolveQName(qname string, defaultNS string) (namespace, localName string) {
	parts := strings.SplitN(qname, ":", 2)
	if len(parts) == 2 {
		// Has prefix
		prefix := parts[0]
		localName = parts[1]
		if uri, ok := nm.GetNamespaceURI(prefix); ok {
			namespace = uri
		} else {
			// Unknown prefix, use as-is
			namespace = defaultNS
		}
	} else {
		// No prefix, use default namespace
		localName = qname
		namespace = defaultNS
	}
	return namespace, localName
}

// GenerateQName generates a qualified name from namespace and local name
func (nm *NamespaceManager) GenerateQName(namespace, localName string) string {
	if namespace == "" {
		return localName
	}
	
	if prefix, ok := nm.GetPrefix(namespace); ok && prefix != "" {
		return fmt.Sprintf("%s:%s", prefix, localName)
	}
	
	return localName
}

// generateUniquePrefix creates a unique prefix when collision occurs
func (nm *NamespaceManager) generateUniquePrefix(base string) string {
	nm.counter++
	return fmt.Sprintf("%s%d", base, nm.counter)
}

// Clone creates a copy of the namespace manager
func (nm *NamespaceManager) Clone() *NamespaceManager {
	nm.mu.RLock()
	defer nm.mu.RUnlock()
	
	clone := &NamespaceManager{
		namespaces: make(map[string]string),
		prefixes:   make(map[string]string),
		counter:    nm.counter,
	}
	
	for prefix, uri := range nm.namespaces {
		clone.namespaces[prefix] = uri
	}
	for uri, prefix := range nm.prefixes {
		clone.prefixes[uri] = prefix
	}
	
	return clone
}

// GetAllNamespaces returns all registered namespaces
func (nm *NamespaceManager) GetAllNamespaces() map[string]string {
	nm.mu.RLock()
	defer nm.mu.RUnlock()
	
	result := make(map[string]string)
	for prefix, uri := range nm.namespaces {
		result[prefix] = uri
	}
	return result
}