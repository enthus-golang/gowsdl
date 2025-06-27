package core

import (
	"testing"
)

func TestNamespaceManager(t *testing.T) {
	nm := NewNamespaceManager()

	t.Run("AddNamespace", func(t *testing.T) {
		nm.AddNamespace("test", "http://example.com/test")
		
		uri, ok := nm.GetNamespaceURI("test")
		if !ok {
			t.Error("Expected to find namespace URI for prefix 'test'")
		}
		if uri != "http://example.com/test" {
			t.Errorf("Expected URI 'http://example.com/test', got '%s'", uri)
		}
		
		prefix, ok := nm.GetPrefix("http://example.com/test")
		if !ok {
			t.Error("Expected to find prefix for URI")
		}
		if prefix != "test" {
			t.Errorf("Expected prefix 'test', got '%s'", prefix)
		}
	})

	t.Run("NamespaceCollision", func(t *testing.T) {
		// Add first namespace
		nm.AddNamespace("ns1", "http://example.com/namespace")
		
		// Try to add same URI with different prefix
		nm.AddNamespace("ns2", "http://example.com/namespace")
		
		// Should generate unique prefix
		uri, ok := nm.GetNamespaceURI("ns2")
		if ok && uri == "http://example.com/namespace" {
			t.Error("Expected ns2 to have a different URI or be renamed")
		}
	})

	t.Run("ResolveQName", func(t *testing.T) {
		nm.AddNamespace("xsd", "http://www.w3.org/2001/XMLSchema")
		
		// Test with prefix
		ns, local := nm.ResolveQName("xsd:string", "http://default.com")
		if ns != "http://www.w3.org/2001/XMLSchema" {
			t.Errorf("Expected namespace 'http://www.w3.org/2001/XMLSchema', got '%s'", ns)
		}
		if local != "string" {
			t.Errorf("Expected local name 'string', got '%s'", local)
		}
		
		// Test without prefix
		ns, local = nm.ResolveQName("element", "http://default.com")
		if ns != "http://default.com" {
			t.Errorf("Expected default namespace, got '%s'", ns)
		}
		if local != "element" {
			t.Errorf("Expected local name 'element', got '%s'", local)
		}
	})

	t.Run("GenerateQName", func(t *testing.T) {
		nm.AddNamespace("soap", "http://schemas.xmlsoap.org/wsdl/soap/")
		
		qname := nm.GenerateQName("http://schemas.xmlsoap.org/wsdl/soap/", "binding")
		if qname != "soap:binding" {
			t.Errorf("Expected 'soap:binding', got '%s'", qname)
		}
		
		// Test with unknown namespace
		qname = nm.GenerateQName("http://unknown.com", "element")
		if qname != "element" {
			t.Errorf("Expected 'element', got '%s'", qname)
		}
	})

	t.Run("RegisterNamespaces", func(t *testing.T) {
		xmlns := map[string]string{
			"wsdl": "http://schemas.xmlsoap.org/wsdl/",
			"http": "http://schemas.xmlsoap.org/wsdl/http/",
		}
		
		nm.RegisterNamespaces(xmlns)
		
		uri, ok := nm.GetNamespaceURI("wsdl")
		if !ok || uri != "http://schemas.xmlsoap.org/wsdl/" {
			t.Error("Failed to register wsdl namespace")
		}
		
		uri, ok = nm.GetNamespaceURI("http")
		if !ok || uri != "http://schemas.xmlsoap.org/wsdl/http/" {
			t.Error("Failed to register http namespace")
		}
	})

	t.Run("Clone", func(t *testing.T) {
		nm.AddNamespace("custom", "http://custom.com")
		
		clone := nm.Clone()
		
		// Verify clone has the namespace
		uri, ok := clone.GetNamespaceURI("custom")
		if !ok || uri != "http://custom.com" {
			t.Error("Clone missing namespace")
		}
		
		// Verify independence
		clone.AddNamespace("new", "http://new.com")
		_, ok = nm.GetNamespaceURI("new")
		if ok {
			t.Error("Original should not have namespace added to clone")
		}
	})
}