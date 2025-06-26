// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at http://mozilla.org/MPL/2.0/.

package generator

import (
	"bytes"
	"context"
	"encoding/xml"
	"fmt"
	"strings"
	"sync"
	"text/template"

	"github.com/hooklift/gowsdl/pkg/core"
	"github.com/hooklift/gowsdl/pkg/generator/templates"
	"github.com/hooklift/gowsdl/pkg/http"
	"github.com/hooklift/gowsdl/pkg/parser"
	"github.com/hooklift/gowsdl/pkg/types"
	"github.com/hooklift/gowsdl/pkg/utils"
)

// Generator handles WSDL code generation
type Generator struct {
	loc                   *http.Location
	rawWSDL               []byte
	pkg                   string
	ignoreTLS             bool
	makePublicFn          func(string) string
	wsdl                  *parser.WSDL
	wsdl2                 *parser.WSDL2
	wsdlVersion           string // "1.1" or "2.0"
	resolvedXSDExternals  map[string]bool
	currentRecursionLevel uint8
	currentNamespace      string
	httpConfig            http.HTTPConfig
	useGenerics           bool
	namespaceManager      *core.NamespaceManager
	typeMapper            *types.TypeMapper
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

	return g.generateCode()
}

func (g *Generator) generateCode() (map[string][]byte, error) {
	gocode := make(map[string][]byte)
	
	var wg sync.WaitGroup
	var mu sync.Mutex
	errChan := make(chan error, 5)
	
	// Generate all components in parallel
	wg.Add(5)
	
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
	tmpl, err := template.New("types").Funcs(funcMap).Parse(templates.TypesTemplate)
	if err != nil {
		return nil, err
	}

	var buf bytes.Buffer
	
	// Get types based on version
	var data interface{}
	if g.wsdlVersion == "2.0" && g.wsdl2 != nil {
		data = g.wsdl2.Types.Schemas
	} else if g.wsdl != nil {
		data = g.wsdl.Types.Schemas
	}
	
	if err := tmpl.Execute(&buf, data); err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

func (g *Generator) genOperations() ([]byte, error) {
	funcMap := g.createFuncMap()
	tmpl, err := template.New("operations").Funcs(funcMap).Parse(templates.UnifiedOperationsTemplate)
	if err != nil {
		return nil, err
	}

	var buf bytes.Buffer
	
	// Get operations based on version
	var data interface{}
	if g.wsdlVersion == "2.0" && g.wsdl2 != nil {
		data = g.wsdl2.Bindings
	} else if g.wsdl != nil {
		data = g.wsdl.Binding
	}
	
	if err := tmpl.Execute(&buf, data); err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

func (g *Generator) genServer() ([]byte, error) {
	// TODO: Implement server generation
	return []byte{}, nil
}

func (g *Generator) genServerHeader() ([]byte, error) {
	// TODO: Implement server header generation
	return []byte{}, nil
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
		"findType":                g.findType,
		"findSOAPAction":          g.findSOAPAction,
		"makeValidXmlTag":         makeValidXmlTag,
		"goString":                goString,
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

// findType and findSOAPAction are placeholder implementations
func (g *Generator) findType(message string) string {
	// TODO: Implement type finding logic
	return message
}

func (g *Generator) findSOAPAction(operation, portType string) string {
	// TODO: Implement SOAP action finding logic
	return ""
}