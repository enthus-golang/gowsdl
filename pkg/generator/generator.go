// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at http://mozilla.org/MPL/2.0/.

package generator

import (
	"bytes"
	"context"
	"encoding/xml"
	"errors"
	"fmt"
	"strings"
	"sync"
	"text/template"

	"github.com/enthus-golang/gowsdl/pkg/core"
	"github.com/enthus-golang/gowsdl/pkg/generator/templates"
	"github.com/enthus-golang/gowsdl/pkg/http"
	"github.com/enthus-golang/gowsdl/pkg/parser"
	"github.com/enthus-golang/gowsdl/pkg/types"
	"github.com/enthus-golang/gowsdl/pkg/utils"
)

// Generator handles WSDL code generation
type Generator struct {
	loc              *http.Location
	rawWSDL          []byte
	pkg              string
	makePublicFn     func(string) string
	wsdl             *parser.WSDL
	wsdl2            *parser.WSDL2
	wsdlVersion      string // "1.1" or "2.0"
	httpConfig       http.HTTPConfig
	useGenerics      bool
	generateServer   bool
	namespaceManager *core.NamespaceManager
	typeMapper       *types.TypeMapper
}

// Option defines a function that configures Generator
type Option func(*Generator) error

// WithPackage sets the package name
func WithPackage(pkg string) Option {
	return func(g *Generator) error {
		pkg = strings.TrimSpace(pkg)
		if pkg == "" {
			pkg = "myservice"
		}
		g.pkg = pkg
		return nil
	}
}

// WithHTTPConfig sets the HTTP configuration
func WithHTTPConfig(config http.HTTPConfig) Option {
	return func(g *Generator) error {
		g.httpConfig = config
		return nil
	}
}

// WithGenerics enables generic code generation
func WithGenerics() Option {
	return func(g *Generator) error {
		g.useGenerics = true
		return nil
	}
}

// WithExportAllTypes controls type visibility
func WithExportAllTypes(export bool) Option {
	return func(g *Generator) error {
		if export {
			g.makePublicFn = utils.MakePublic
		} else {
			g.makePublicFn = func(id string) string { return id }
		}
		return nil
	}
}

// WithServerGeneration controls whether to generate server code
func WithServerGeneration(generate bool) Option {
	return func(g *Generator) error {
		g.generateServer = generate
		return nil
	}
}

// New creates a new Generator instance
func New(file string, opts ...Option) (*Generator, error) {
	loc, err := http.ParseLocation(file)
	if err != nil {
		return nil, err
	}

	g := &Generator{
		loc:              loc,
		pkg:              "myservice",
		makePublicFn:     func(id string) string { return id },
		namespaceManager: core.NewNamespaceManager(),
		typeMapper:       types.NewTypeMapper(),
	}

	for _, opt := range opts {
		if err := opt(g); err != nil {
			return nil, err
		}
	}

	return g, nil
}

// Generate initiates the code generation process
func (g *Generator) Generate(ctx context.Context) (map[string][]byte, error) {
	if err := g.unmarshal(ctx); err != nil {
		return nil, err
	}

	// Process WSDL nodes
	var schemas []*parser.XSDSchema
	if g.wsdlVersion == "2.0" && g.wsdl2 != nil {
		schemas = g.wsdl2.Types.Schemas
	} else if g.wsdl != nil {
		schemas = g.wsdl.Types.Schemas
	}
	
	for _, schema := range schemas {
		parser.NewTraverser(schema, schemas).Traverse()
	}

	return g.generateCode(ctx)
}

func (g *Generator) generateCode(ctx context.Context) (map[string][]byte, error) {
	gocode := make(map[string][]byte)
	
	var wg sync.WaitGroup
	var mu sync.Mutex
	
	// Calculate number of goroutines based on generateServer flag
	numGoroutines := 3 // header, types, operations
	if g.generateServer {
		numGoroutines = 6 // add server, server_header, server_wsdl
	}
	errChan := make(chan error, numGoroutines)
	
	// Generate all components in parallel
	wg.Add(numGoroutines)
	
	go func() {
		defer wg.Done()
		if data, err := g.genHeader(); err != nil {
			errChan <- err
		} else {
			mu.Lock()
			gocode["header"] = data
			mu.Unlock()
		}
	}()
	
	go func() {
		defer wg.Done()
		if data, err := g.genTypes(); err != nil {
			errChan <- err
		} else {
			mu.Lock()
			gocode["types"] = data
			mu.Unlock()
		}
	}()
	
	go func() {
		defer wg.Done()
		if data, err := g.genOperations(); err != nil {
			errChan <- err
		} else {
			mu.Lock()
			gocode["operations"] = data
			mu.Unlock()
		}
	}()
	
	// Only generate server files if generateServer is enabled
	if g.generateServer {
		go func() {
			defer wg.Done()
			if data, err := g.genServer(); err != nil {
				errChan <- err
			} else {
				mu.Lock()
				gocode["server"] = data
				mu.Unlock()
			}
		}()
		
		go func() {
			defer wg.Done()
			if data, err := g.genServerHeader(); err != nil {
				errChan <- err
			} else {
				mu.Lock()
				gocode["server_header"] = data
				mu.Unlock()
			}
		}()
		
		go func() {
			defer wg.Done()
			if data, err := g.genServerWSDL(ctx); err != nil {
				errChan <- err
			} else {
				mu.Lock()
				gocode["server_wsdl"] = data
				mu.Unlock()
			}
		}()
	}
	
	wg.Wait()
	close(errChan)
	
	// Check for errors
	for err := range errChan {
		return nil, err
	}
	
	return gocode, nil
}

func (g *Generator) unmarshal(ctx context.Context) error {
	data, err := g.fetchFile(ctx, g.loc)
	if err != nil {
		return &types.WSDLError{
			Op:   "fetch",
			Path: g.loc.String(),
			Err:  err,
		}
	}
	g.rawWSDL = data

	// Detect WSDL version
	version, err := parser.DetectWSDLVersion(data)
	if err != nil {
		return &types.WSDLError{
			Op:   "detect_version",
			Path: g.loc.String(),
			Err:  err,
		}
	}
	g.wsdlVersion = version

	// Parse based on version
	if version == "2.0" {
		g.wsdl2 = new(parser.WSDL2)
		err = xml.Unmarshal(data, g.wsdl2)
		if err != nil {
			return &types.WSDLError{
				Op:   "parse",
				Path: g.loc.String(),
				Err:  fmt.Errorf("failed to unmarshal WSDL 2.0: %w", err),
			}
		}
	} else {
		g.wsdl = new(parser.WSDL)
		err = xml.Unmarshal(data, g.wsdl)
		if err != nil {
			return &types.WSDLError{
				Op:   "parse",
				Path: g.loc.String(),
				Err:  fmt.Errorf("failed to unmarshal WSDL 1.1: %w", err),
			}
		}
	}

	return nil
}

func (g *Generator) fetchFile(ctx context.Context, loc *http.Location) ([]byte, error) {
	if loc.IsLocal() {
		return loc.ReadFile()
	}
	return http.DownloadFile(ctx, loc.URL(), g.httpConfig)
}

// Template generation methods
func (g *Generator) genHeader() ([]byte, error) {
	funcMap := g.createFuncMap()
	tmpl, err := template.New("header").Funcs(funcMap).Parse(templates.HeaderTemplate)
	if err != nil {
		return nil, err
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, g.pkg); err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

func (g *Generator) genTypes() ([]byte, error) {
	funcMap := g.createFuncMap()
	
	// Pre-process complex types into a map for efficient lookup
	complexTypeMap := make(map[string]bool)
	
	// Collect all complex types first
	var schemas []*parser.XSDSchema
	if g.wsdlVersion == "2.0" && g.wsdl2 != nil {
		schemas = g.wsdl2.Types.Schemas
	} else if g.wsdl != nil {
		schemas = g.wsdl.Types.Schemas
	}
	
	for _, schema := range schemas {
		if schema.ComplexTypes != nil {
			for _, ct := range schema.ComplexTypes {
				complexTypeMap[ct.Name] = true
			}
		}
	}
	
	// Collect inline complex types
	inlineTypes := CollectInlineComplexTypes(schemas)
	
	// Add function to check if a type is complex
	funcMap["isComplexType"] = func(typeName string) bool {
		// Remove namespace prefix if present
		if idx := strings.LastIndex(typeName, ":"); idx >= 0 {
			typeName = typeName[idx+1:]
		}
		return complexTypeMap[typeName]
	}
	
	// Template State Management for Inline Type Name Resolution
	// ========================================================
	// The following code implements a state management pattern within the template
	// execution to handle nested inline complex types. This is necessary because
	// templates don't have access to the parent context when processing nested structures.
	//
	// How it works:
	// 1. When the template starts processing a complex type, it calls setCurrentType
	//    to store the type name in currentTypeName
	// 2. When processing nested inline types, the template uses "$" as a special
	//    marker for parentName to indicate "use the current context"
	// 3. getInlineTypeName checks if parentName is "$" and substitutes the stored
	//    currentTypeName value
	//
	// Example template usage:
	//   {{setCurrentType .Name}}  <!-- Store "OrderType" -->
	//   {{range .Sequence}}
	//     {{$inlineTypeName := getInlineTypeName "$" .Name}}  <!-- "$" becomes "OrderType" -->
	//   {{end}}
	//
	// WARNING: This pattern is fragile as it relies on proper sequencing of template
	// calls. Always ensure setCurrentType is called before processing nested elements.
	
	currentTypeName := ""
	
	// getInlineTypeName retrieves the generated name for an inline complex type
	// Parameters:
	//   - parentName: The parent type name, or "$" to use the current context type
	//   - elementName: The element name containing the inline type
	funcMap["getInlineTypeName"] = func(parentName, elementName string) string {
		// Special case: "$" means use the type currently being processed
		// This is a magic string used by templates to reference the current context
		if parentName == "$" {
			parentName = currentTypeName
		}
		if typeName, ok := FindInlineType(inlineTypes, parentName, elementName); ok {
			return typeName
		}
		// Fallback to generating a name if not found
		return utils.MakePublic(parentName) + utils.MakePublic(elementName) + "Type"
	}
	
	// setCurrentType stores the name of the type currently being processed by the template
	// This enables nested template calls to access the parent context via "$"
	// Returns empty string so it can be used inline in templates without output
	funcMap["setCurrentType"] = func(typeName string) string {
		currentTypeName = typeName
		return ""
	}
	
	// Add function to check if this should be an optional pointer
	funcMap["isOptionalType"] = IsOptionalType
	
	// Add dict function for creating template data structures
	funcMap["dict"] = func(values ...interface{}) (map[string]interface{}, error) {
		if len(values)%2 != 0 {
			return nil, fmt.Errorf("dict must have even number of arguments")
		}
		dict := make(map[string]interface{})
		for i := 0; i < len(values); i += 2 {
			key, ok := values[i].(string)
			if !ok {
				return nil, fmt.Errorf("dict keys must be strings")
			}
			dict[key] = values[i+1]
		}
		return dict, nil
	}
	
	tmpl, err := template.New("types").Funcs(funcMap).Parse(templates.TypesTemplate)
	if err != nil {
		return nil, err
	}

	var buf bytes.Buffer
	
	// Merge all schemas into a single data structure
	data := struct {
		SimpleType      []*parser.XSDSimpleType
		ComplexTypes    []*parser.XSDComplexType
		Elements        []*parser.XSDElement
		Schemas         []*parser.XSDSchema
		Messages        []*parser.WSDLMessage
		InlineTypes     map[string]*InlineComplexType
		TargetNamespace string
	}{}
	
	// Add RPC-style messages for WSDL 1.1
	if g.wsdl != nil {
		data.Messages = g.wsdl.Messages
	}
	
	data.Schemas = schemas
	data.InlineTypes = inlineTypes
	
	// Merge all schemas and get first non-empty target namespace
	for _, schema := range schemas {
		if data.TargetNamespace == "" && schema.TargetNamespace != "" {
			data.TargetNamespace = schema.TargetNamespace
		}
		if schema.SimpleType != nil {
			data.SimpleType = append(data.SimpleType, schema.SimpleType...)
		}
		if schema.ComplexTypes != nil {
			data.ComplexTypes = append(data.ComplexTypes, schema.ComplexTypes...)
		}
		if schema.Elements != nil {
			data.Elements = append(data.Elements, schema.Elements...)
		}
	}
	
	if err := tmpl.Execute(&buf, data); err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

func (g *Generator) genOperations() ([]byte, error) {
	funcMap := g.createFuncMap()
	
	// Choose template based on WSDL version
	var tmplText string
	var data interface{}
	
	if g.wsdlVersion == "2.0" && g.wsdl2 != nil {
		tmplText = templates.OperationsWSDL2Template
		data = g.wsdl2.Interfaces
	} else if g.wsdl != nil {
		tmplText = templates.OperationsWSDL1Template
		data = g.wsdl.PortTypes
	} else {
		return []byte{}, nil
	}
	
	tmpl, err := template.New("operations").Funcs(funcMap).Parse(tmplText)
	if err != nil {
		return nil, err
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

func (g *Generator) genServer() ([]byte, error) {
	funcMap := g.createFuncMap()
	
	// Choose template based on WSDL version
	var tmplText string
	var data interface{}
	
	if g.wsdlVersion == "2.0" && g.wsdl2 != nil {
		tmplText = templates.ServerWSDL2Template
		data = g.wsdl2.Interfaces
	} else if g.wsdl != nil {
		tmplText = templates.ServerTemplate
		data = g.wsdl.PortTypes
	} else {
		return nil, errors.New("no WSDL data available")
	}
	
	tmpl, err := template.New("server").Funcs(funcMap).Parse(tmplText)
	if err != nil {
		return nil, err
	}
	
	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return nil, err
	}
	
	return buf.Bytes(), nil
}

func (g *Generator) genServerHeader() ([]byte, error) {
	funcMap := g.createFuncMap()
	tmpl, err := template.New("server_header").Funcs(funcMap).Parse(templates.ServerHeaderTemplate)
	if err != nil {
		return nil, err
	}
	
	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, g.pkg); err != nil {
		return nil, err
	}
	
	return buf.Bytes(), nil
}

func (g *Generator) genServerWSDL(ctx context.Context) ([]byte, error) {
	// Read the original WSDL file content
	wsdlContent, err := g.fetchFile(ctx, g.loc)
	if err != nil {
		return nil, fmt.Errorf("failed to read WSDL file: %w", err)
	}
	
	funcMap := g.createFuncMap()
	tmpl, err := template.New("server_wsdl").Funcs(funcMap).Parse(templates.ServerWSDLTemplate)
	if err != nil {
		return nil, err
	}
	
	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, string(wsdlContent)); err != nil {
		return nil, err
	}
	
	return buf.Bytes(), nil
}

// createFuncMap creates the template function map
func (g *Generator) createFuncMap() template.FuncMap {
	return template.FuncMap{
		"toGoType":                g.typeMapper.MapXSDTypeToGoType,
		"replaceReservedWords":    g.typeMapper.EscapeReservedWord,
		"replaceAttrReservedWords": g.typeMapper.EscapeReservedWord,
		"makePublic":              utils.MakePublic,
		"makePrivate":             makePrivate,
		"comment":                 comment,
		"removePointerFromType":   removePointerFromType,
		"removeNamespacePrefix":   removeNamespacePrefix,
		"removeNS":                removeNamespacePrefix,
		"findType":                g.findType,
		"findSOAPAction":          g.findSOAPAction,
		"makeValidXmlTag":         makeValidXmlTag,
		"goString":                goString,
		"sanitizeEnumValue":       sanitizeEnumValue,
		"getTargetNamespace":      g.getTargetNamespace,
		"isRPCStyleMessage":       g.isRPCStyleMessage,
	}
}

// Helper functions
func makePrivate(identifier string) string {
	if identifier == "" {
		return ""
	}
	return strings.ToLower(identifier[:1]) + identifier[1:]
}

func comment(text string) string {
	lines := strings.Split(text, "\n")
	for i, line := range lines {
		lines[i] = "// " + strings.TrimSpace(line)
	}
	return strings.Join(lines, "\n")
}

func removePointerFromType(typ string) string {
	return strings.TrimPrefix(typ, "*")
}

func removeNamespacePrefix(name string) string {
	parts := strings.Split(name, ":")
	if len(parts) > 1 {
		return parts[len(parts)-1]
	}
	return name
}

func makeValidXmlTag(xmlName, goName string) string {
	if xmlName == goName {
		return xmlName
	}
	return xmlName + ",omitempty"
}

func goString(s string) string {
	return strings.ReplaceAll(s, "\"", "\\\"")
}

// sanitizeEnumValue converts a string to a valid Go identifier for enum constants
func sanitizeEnumValue(s string) string {
	// For backward compatibility, convert the way the original generator did
	if s == "" {
		return "EmptyString"
	}
	
	// Remove spaces entirely (don't convert to underscores)
	s = strings.ReplaceAll(s, " ", "")
	// Remove hyphens
	s = strings.ReplaceAll(s, "-", "")
	// Replace dots with nothing
	s = strings.ReplaceAll(s, ".", "")
	// Remove other special characters
	s = strings.ReplaceAll(s, ":", "")
	s = strings.ReplaceAll(s, "/", "")
	s = strings.ReplaceAll(s, "\\", "")
	s = strings.ReplaceAll(s, "(", "")
	s = strings.ReplaceAll(s, ")", "")
	s = strings.ReplaceAll(s, "[", "")
	s = strings.ReplaceAll(s, "]", "")
	s = strings.ReplaceAll(s, "{", "")
	s = strings.ReplaceAll(s, "}", "")
	s = strings.ReplaceAll(s, ",", "")
	s = strings.ReplaceAll(s, ";", "")
	s = strings.ReplaceAll(s, "'", "")
	s = strings.ReplaceAll(s, "\"", "")
	
	// If the result is empty after cleanup, return "Empty"
	if s == "" {
		return "Empty"
	}
	
	return s
}

// isRPCStyleMessage determines if a message requires a wrapper type
// This is true for messages with multiple parts or single primitive-typed parts
func (g *Generator) isRPCStyleMessage(msg *parser.WSDLMessage) bool {
	if msg == nil {
		return false
	}
	
	// Multiple parts always require a wrapper
	if len(msg.Parts) > 1 {
		return true
	}
	
	// Single part with primitive type (no element reference) requires a wrapper
	if len(msg.Parts) == 1 && msg.Parts[0].Element == "" && msg.Parts[0].Type != "" {
		return true
	}
	
	return false
}

// findType finds the actual type name from a message reference
func (g *Generator) findType(message string) string {
	// Remove namespace prefix from message name
	message = removeNamespacePrefix(message)
	
	// For WSDL 1.1, messages typically reference elements
	// Try to find the message and get its part's element/type
	if g.wsdl != nil {
		for _, msg := range g.wsdl.Messages {
			if msg.Name == message {
				// Check if this is an RPC-style operation
				if g.isRPCStyleMessage(msg) {
					// For RPC-style operations, use the message name as the type
					// This will generate a wrapper type in the types template
					return msg.Name
				}
				
				if len(msg.Parts) > 0 {
					// Get the first part's element or type
					part := msg.Parts[0]
					if part.Element != "" {
						return removeNamespacePrefix(part.Element)
					}
					if part.Type != "" {
						// For document-style with a complex type reference
						return removeNamespacePrefix(part.Type)
					}
				}
			}
		}
	}
	
	// For WSDL 2.0, check interfaces
	if g.wsdl2 != nil {
		// In WSDL 2.0, operations directly reference elements
		return message
	}
	
	// Fallback: return the message name without namespace
	return message
}

func (g *Generator) findSOAPAction(operation, portType string) string {
	if g.wsdl == nil {
		return ""
	}

	// Find the binding that references the port type
	for _, binding := range g.wsdl.Binding {
		// Check if this binding is for the port type we're looking for
		bindingPortType := binding.Type
		if idx := strings.LastIndex(bindingPortType, ":"); idx != -1 {
			bindingPortType = bindingPortType[idx+1:]
		}
		
		// Normalize names by removing common suffixes to provide a more robust comparison
		normalizedBindingPortType := bindingPortType
		normalizedBindingPortType = strings.TrimSuffix(normalizedBindingPortType, "PortType")
		normalizedBindingPortType = strings.TrimSuffix(normalizedBindingPortType, "Port")
		normalizedBindingPortType = strings.TrimSuffix(normalizedBindingPortType, "Type")
		
		normalizedPortType := portType
		normalizedPortType = strings.TrimSuffix(normalizedPortType, "PortType")
		normalizedPortType = strings.TrimSuffix(normalizedPortType, "Port")
		normalizedPortType = strings.TrimSuffix(normalizedPortType, "Type")
		
		if strings.EqualFold(normalizedBindingPortType, normalizedPortType) {
			// Find the operation in this binding
			for _, op := range binding.Operations {
				if op.Name == operation {
					return op.SOAPOperation.SOAPAction
				}
			}
		}
	}
	
	return ""
}

func (g *Generator) getTargetNamespace() string {
	// Get the target namespace from the first schema that has one
	var schemas []*parser.XSDSchema
	if g.wsdlVersion == "2.0" && g.wsdl2 != nil {
		schemas = g.wsdl2.Types.Schemas
	} else if g.wsdl != nil {
		schemas = g.wsdl.Types.Schemas
	}
	
	for _, schema := range schemas {
		if schema.TargetNamespace != "" {
			return schema.TargetNamespace
		}
	}
	return ""
}