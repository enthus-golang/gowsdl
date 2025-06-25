// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at http://mozilla.org/MPL/2.0/.

package gowsdl

import (
	"bytes"
	"context"
	"encoding/xml"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"sync"
	"text/template"
	"time"
	"unicode"
)

const maxRecursion uint8 = 20

// WSDLError represents an error that occurred during WSDL processing
type WSDLError struct {
	Op   string // operation that failed
	Path string // file path or URL
	Err  error  // underlying error
}

func (e *WSDLError) Error() string {
	if e.Path != "" {
		return fmt.Sprintf("wsdl %s %q: %v", e.Op, e.Path, e.Err)
	}
	return fmt.Sprintf("wsdl %s: %v", e.Op, e.Err)
}

func (e *WSDLError) Unwrap() error {
	return e.Err
}

// SchemaError represents an error that occurred during XSD schema processing
type SchemaError struct {
	Op     string // operation that failed
	Schema string // schema location or reference
	Err    error  // underlying error
}

func (e *SchemaError) Error() string {
	if e.Schema != "" {
		return fmt.Sprintf("schema %s %q: %v", e.Op, e.Schema, e.Err)
	}
	return fmt.Sprintf("schema %s: %v", e.Op, e.Err)
}

func (e *SchemaError) Unwrap() error {
	return e.Err
}

// GoWSDL defines the struct for WSDL generator.
type GoWSDL struct {
	loc                   *Location
	rawWSDL               []byte
	pkg                   string
	ignoreTLS             bool
	makePublicFn          func(string) string
	wsdl                  *WSDL
	wsdl2                 *WSDL2
	wsdlVersion           string // "1.1" or "2.0"
	resolvedXSDExternals  map[string]bool
	currentRecursionLevel uint8
	currentNamespace      string
	httpConfig            *HTTPClientConfig
	useGenerics           bool
	namespaceManager      *NamespaceManager
}

// Method setNS sets (and returns) the currently active XML namespace.
func (g *GoWSDL) setNS(ns string) string {
	g.currentNamespace = ns
	return ns
}

// Method setNS returns the currently active XML namespace.
func (g *GoWSDL) getNS() string {
	return g.currentNamespace
}

var cacheDir = filepath.Join(os.TempDir(), "gowsdl-cache")

func init() {
	err := os.MkdirAll(cacheDir, 0700)
	if err != nil {
		log.Println("Create cache directory", "error", err)
		os.Exit(1)
	}
}

// downloadFile downloads a file from the given URL using the provided HTTP configuration
func downloadFile(ctx context.Context, url string, httpConfig *HTTPClientConfig) ([]byte, error) {
	client := httpConfig.Build()

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Set user agent
	if httpConfig.UserAgent != "" {
		req.Header.Set("User-Agent", httpConfig.UserAgent)
	}

	var resp *http.Response
	var lastErr error

	// Retry logic
	for attempt := 0; attempt <= httpConfig.MaxRetries; attempt++ {
		if attempt > 0 {
			time.Sleep(httpConfig.RetryDelay)
		}

		resp, err = client.Do(req)
		// Check for context cancellation
		if ctx.Err() != nil {
			return nil, ctx.Err()
		}
		if err == nil && resp.StatusCode < 500 {
			break
		}
		lastErr = err
		if resp != nil {
			_ = resp.Body.Close()
		}
	}

	if lastErr != nil {
		return nil, fmt.Errorf("failed after %d retries: %w", httpConfig.MaxRetries+1, lastErr)
	}

	defer func() {
		_ = resp.Body.Close()
	}()

	if resp.StatusCode != 200 {
		return nil, &WSDLError{
			Op:   "download",
			Path: url,
			Err:  fmt.Errorf("received HTTP %d %s", resp.StatusCode, http.StatusText(resp.StatusCode)),
		}
	}

	// Limit response size
	limitedReader := io.LimitReader(resp.Body, httpConfig.MaxResponseSize)
	data, err := io.ReadAll(limitedReader)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	return data, nil
}

// NewGoWSDLWithConfig initializes WSDL generator with custom HTTP configuration.
func NewGoWSDLWithConfig(file, pkg string, httpConfig *HTTPClientConfig, exportAllTypes bool) (*GoWSDL, error) {
	file = strings.TrimSpace(file)
	if file == "" {
		return nil, errors.New("WSDL file is required to generate Go proxy")
	}

	pkg = strings.TrimSpace(pkg)
	if pkg == "" {
		pkg = "myservice"
	}
	makePublicFn := func(id string) string { return id }
	if exportAllTypes {
		makePublicFn = makePublic
	}

	r, err := ParseLocation(file)
	if err != nil {
		return nil, err
	}

	if httpConfig == nil {
		httpConfig = DefaultHTTPClientConfig()
	}

	return &GoWSDL{
		loc:              r,
		pkg:              pkg,
		ignoreTLS:        httpConfig.InsecureSkipVerify,
		makePublicFn:     makePublicFn,
		httpConfig:       httpConfig,
		namespaceManager: NewNamespaceManager(),
	}, nil
}

// NewGoWSDLWithOptions initializes WSDL generator with all available options.
func NewGoWSDLWithOptions(file, pkg string, httpConfig *HTTPClientConfig, exportAllTypes bool, useGenerics bool) (*GoWSDL, error) {
	file = strings.TrimSpace(file)
	if file == "" {
		return nil, errors.New("WSDL file is required to generate Go proxy")
	}

	pkg = strings.TrimSpace(pkg)
	if pkg == "" {
		pkg = "myservice"
	}
	makePublicFn := func(id string) string { return id }
	if exportAllTypes {
		makePublicFn = makePublic
	}

	r, err := ParseLocation(file)
	if err != nil {
		return nil, err
	}

	if httpConfig == nil {
		httpConfig = DefaultHTTPClientConfig()
	}

	return &GoWSDL{
		loc:              r,
		pkg:              pkg,
		ignoreTLS:        httpConfig.InsecureSkipVerify,
		makePublicFn:     makePublicFn,
		httpConfig:       httpConfig,
		useGenerics:      useGenerics,
		namespaceManager: NewNamespaceManager(),
	}, nil
}

// NewGoWSDL initializes WSDL generator.
func NewGoWSDL(file, pkg string, ignoreTLS bool, exportAllTypes bool) (*GoWSDL, error) {
	file = strings.TrimSpace(file)
	if file == "" {
		return nil, errors.New("WSDL file is required to generate Go proxy")
	}

	pkg = strings.TrimSpace(pkg)
	if pkg == "" {
		pkg = "myservice"
	}
	makePublicFn := func(id string) string { return id }
	if exportAllTypes {
		makePublicFn = makePublic
	}

	r, err := ParseLocation(file)
	if err != nil {
		return nil, err
	}

	// Create default HTTP config for backward compatibility
	httpConfig := DefaultHTTPClientConfig()
	httpConfig.InsecureSkipVerify = ignoreTLS

	return &GoWSDL{
		loc:              r,
		pkg:              pkg,
		ignoreTLS:        ignoreTLS,
		makePublicFn:     makePublicFn,
		httpConfig:       httpConfig,
		namespaceManager: NewNamespaceManager(),
	}, nil
}

// StartWithContext initiates the code generation process with context support.
func (g *GoWSDL) StartWithContext(ctx context.Context) (map[string][]byte, error) {
	gocode := make(map[string][]byte)

	err := g.unmarshal(ctx)
	if err != nil {
		return nil, err
	}

	// Process WSDL nodes
	var schemas []*XSDSchema
	if g.wsdlVersion == "2.0" && g.wsdl2 != nil {
		schemas = g.wsdl2.Types.Schemas
	} else if g.wsdl != nil {
		schemas = g.wsdl.Types.Schemas
	}
	
	for _, schema := range schemas {
		newTraverser(schema, schemas).traverse()
	}

	var wg sync.WaitGroup
	var mu sync.Mutex
	errChan := make(chan error, 5) // Buffer for all goroutines
	
	// Generate header
	wg.Add(1)
	go func() {
		defer wg.Done()
		data, err := g.genHeader()
		if err != nil {
			errChan <- fmt.Errorf("generating header: %w", err)
			return
		}
		mu.Lock()
		gocode["header"] = data
		mu.Unlock()
	}()

	// Generate types
	wg.Add(1)
	go func() {
		defer wg.Done()
		data, err := g.genTypes()
		if err != nil {
			errChan <- fmt.Errorf("generating types: %w", err)
			return
		}
		mu.Lock()
		gocode["types"] = data
		mu.Unlock()
	}()

	// Generate operations
	wg.Add(1)
	go func() {
		defer wg.Done()
		data, err := g.genOperations()
		if err != nil {
			errChan <- fmt.Errorf("generating operations: %w", err)
			return
		}
		mu.Lock()
		gocode["operations"] = data
		mu.Unlock()
	}()

	// Generate server
	wg.Add(1)
	go func() {
		defer wg.Done()
		data, err := g.genServer()
		if err != nil {
			errChan <- fmt.Errorf("generating server: %w", err)
			return
		}
		mu.Lock()
		gocode["server"] = data
		mu.Unlock()
	}()

	// Generate server header
	wg.Add(1)
	go func() {
		defer wg.Done()
		data, err := g.genServerHeader()
		if err != nil {
			errChan <- fmt.Errorf("generating server header: %w", err)
			return
		}
		mu.Lock()
		gocode["server_header"] = data
		mu.Unlock()
	}()

	// Wait for all goroutines to finish
	wg.Wait()
	close(errChan)

	// Collect any errors
	var errors []error
	for err := range errChan {
		errors = append(errors, err)
	}
	
	// If any errors occurred, return them
	if len(errors) > 0 {
		// Combine multiple errors into a single error
		if len(errors) == 1 {
			return nil, errors[0]
		}
		return nil, fmt.Errorf("multiple errors during code generation: %v", errors)
	}

	gocode["server_wsdl"] = []byte("var wsdl = `" + string(g.rawWSDL) + "`")

	return gocode, nil
}

// Start initiaties the code generation process by starting two goroutines: one
// to generate types and another one to generate operations.
func (g *GoWSDL) Start() (map[string][]byte, error) {
	return g.StartWithContext(context.Background())
}

func (g *GoWSDL) fetchFile(ctx context.Context, loc *Location) (data []byte, err error) {
	if loc.f != "" {
		log.Println("Reading", "file", loc.f)
		data, err = os.ReadFile(loc.f)
	} else {
		log.Println("Downloading", "file", loc.u.String())
		data, err = downloadFile(ctx, loc.u.String(), g.httpConfig)
	}
	return
}

func (g *GoWSDL) unmarshal(ctx context.Context) error {
	data, err := g.fetchFile(ctx, g.loc)
	if err != nil {
		return &WSDLError{
			Op:   "fetch",
			Path: g.loc.String(),
			Err:  err,
		}
	}
	g.rawWSDL = data

	// Detect WSDL version
	version, err := detectWSDLVersion(data)
	if err != nil {
		return &WSDLError{
			Op:   "detect_version",
			Path: g.loc.String(),
			Err:  err,
		}
	}
	g.wsdlVersion = version

	// Parse based on version
	if version == "2.0" {
		g.wsdl2 = new(WSDL2)
		err = xml.Unmarshal(data, g.wsdl2)
		if err != nil {
			return &WSDLError{
				Op:   "parse",
				Path: g.loc.String(),
				Err:  fmt.Errorf("failed to unmarshal WSDL 2.0: %w", err),
			}
		}

		// Register WSDL 2.0 namespaces
		if g.wsdl2.Xmlns != nil {
			g.namespaceManager.RegisterNamespaces(g.wsdl2.Xmlns)
		}

		// Register namespaces from each schema
		for _, schema := range g.wsdl2.Types.Schemas {
			if schema.Xmlns != nil {
				g.namespaceManager.RegisterNamespaces(schema.Xmlns)
			}
			err = g.resolveXSDExternals(ctx, schema, g.loc)
			if err != nil {
				return &WSDLError{
					Op:   "resolve_schemas",
					Path: g.loc.String(),
					Err:  err,
				}
			}
		}
	} else {
		// Default to WSDL 1.1
		g.wsdl = new(WSDL)
		err = xml.Unmarshal(data, g.wsdl)
		if err != nil {
			return &WSDLError{
				Op:   "parse",
				Path: g.loc.String(),
				Err:  fmt.Errorf("failed to unmarshal WSDL 1.1: %w", err),
			}
		}

		// Register WSDL namespaces
		if g.wsdl.Xmlns != nil {
			g.namespaceManager.RegisterNamespaces(g.wsdl.Xmlns)
		}

		// Register namespaces from each schema
		for _, schema := range g.wsdl.Types.Schemas {
			if schema.Xmlns != nil {
				g.namespaceManager.RegisterNamespaces(schema.Xmlns)
			}
			err = g.resolveXSDExternals(ctx, schema, g.loc)
			if err != nil {
				return &WSDLError{
					Op:   "resolve_schemas",
					Path: g.loc.String(),
					Err:  err,
				}
			}
		}
	}

	return nil
}

func (g *GoWSDL) resolveXSDExternals(ctx context.Context, schema *XSDSchema, loc *Location) error {
	download := func(base *Location, ref string) error {
		location, err := base.Parse(ref)
		if err != nil {
			return &SchemaError{
				Op:     "parse_reference",
				Schema: ref,
				Err:    err,
			}
		}
		schemaKey := location.String()
		if g.resolvedXSDExternals[location.String()] {
			return nil
		}
		if g.resolvedXSDExternals == nil {
			g.resolvedXSDExternals = make(map[string]bool, maxRecursion)
		}
		g.resolvedXSDExternals[schemaKey] = true

		var data []byte
		if data, err = g.fetchFile(ctx, location); err != nil {
			return &SchemaError{
				Op:     "fetch",
				Schema: location.String(),
				Err:    err,
			}
		}

		newschema := new(XSDSchema)

		err = xml.Unmarshal(data, newschema)
		if err != nil {
			return &SchemaError{
				Op:     "parse",
				Schema: location.String(),
				Err:    fmt.Errorf("failed to unmarshal XSD schema: %w", err),
			}
		}

		// Register namespaces from the newly loaded schema
		if newschema.Xmlns != nil {
			g.namespaceManager.RegisterNamespaces(newschema.Xmlns)
		}

		if (len(newschema.Includes) > 0 || len(newschema.Imports) > 0) &&
			maxRecursion > g.currentRecursionLevel {
			g.currentRecursionLevel++

			err = g.resolveXSDExternals(ctx, newschema, location)
			if err != nil {
				return &SchemaError{
					Op:     "resolve_nested",
					Schema: location.String(),
					Err:    err,
				}
			}
		}

		// Append to the appropriate schema collection based on WSDL version
		if g.wsdlVersion == "2.0" && g.wsdl2 != nil {
			g.wsdl2.Types.Schemas = append(g.wsdl2.Types.Schemas, newschema)
		} else if g.wsdl != nil {
			g.wsdl.Types.Schemas = append(g.wsdl.Types.Schemas, newschema)
		}

		return nil
	}

	for _, impts := range schema.Imports {
		// Download the file only if we have a hint in the form of schemaLocation.
		if impts.SchemaLocation == "" {
			log.Printf("[WARN] Don't know where to find XSD for %s", impts.Namespace)
			continue
		}

		if e := download(loc, impts.SchemaLocation); e != nil {
			return e
		}
	}

	for _, incl := range schema.Includes {
		if e := download(loc, incl.SchemaLocation); e != nil {
			return e
		}
	}

	return nil
}

func (g *GoWSDL) genTypes() ([]byte, error) {
	funcMap := template.FuncMap{
		"toGoType":                 toGoType,
		"stripns":                  stripns,
		"replaceReservedWords":     replaceReservedWords,
		"replaceAttrReservedWords": replaceAttrReservedWords,
		"normalize":                normalize,
		"makePublic":               g.makePublicFn,
		"makeFieldPublic":          makePublic,
		"comment":                  comment,
		"removeNS":                 removeNS,
		"goString":                 goString,
		"findNameByType":           g.findNameByType,
		"removePointerFromType":    removePointerFromType,
		"setNS":                    g.setNS,
		"getNS":                    g.getNS,
	}

	data := new(bytes.Buffer)
	tmpl := template.Must(template.New("types").Funcs(funcMap).Parse(typesTmpl))
	err := tmpl.Execute(data, g.wsdl.Types)
	if err != nil {
		return nil, err
	}

	return data.Bytes(), nil
}

func (g *GoWSDL) genOperations() ([]byte, error) {
	funcMap := template.FuncMap{
		"toGoType":             toGoType,
		"stripns":              stripns,
		"replaceReservedWords": replaceReservedWords,
		"normalize":            normalize,
		"makePublic":           g.makePublicFn,
		"makePrivate":          makePrivate,
		"findType":             g.findType,
		"findSOAPAction":       g.findSOAPAction,
		"findServiceAddress":   g.findServiceAddress,
	}

	data := new(bytes.Buffer)
	
	// Choose template based on useGenerics flag
	templateContent := opsTmpl
	if g.useGenerics {
		templateContent = genericOpsTmpl
	}
	
	tmpl := template.Must(template.New("operations").Funcs(funcMap).Parse(templateContent))
	err := tmpl.Execute(data, g.wsdl.PortTypes)
	if err != nil {
		return nil, err
	}

	return data.Bytes(), nil
}

func (g *GoWSDL) genServer() ([]byte, error) {
	funcMap := template.FuncMap{
		"toGoType":             toGoType,
		"stripns":              stripns,
		"replaceReservedWords": replaceReservedWords,
		"makePublic":           g.makePublicFn,
		"findType":             g.findType,
		"findSOAPAction":       g.findSOAPAction,
		"findServiceAddress":   g.findServiceAddress,
	}

	data := new(bytes.Buffer)
	tmpl := template.Must(template.New("server").Funcs(funcMap).Parse(serverTmpl))
	err := tmpl.Execute(data, g.wsdl.PortTypes)
	if err != nil {
		return nil, err
	}

	return data.Bytes(), nil
}

func (g *GoWSDL) genHeader() ([]byte, error) {
	funcMap := template.FuncMap{
		"toGoType":             toGoType,
		"stripns":              stripns,
		"replaceReservedWords": replaceReservedWords,
		"normalize":            normalize,
		"makePublic":           g.makePublicFn,
		"findType":             g.findType,
		"comment":              comment,
	}

	data := new(bytes.Buffer)
	tmpl := template.Must(template.New("header").Funcs(funcMap).Parse(headerTmpl))
	err := tmpl.Execute(data, g.pkg)
	if err != nil {
		return nil, err
	}

	return data.Bytes(), nil
}

func (g *GoWSDL) genServerHeader() ([]byte, error) {
	funcMap := template.FuncMap{
		"toGoType":             toGoType,
		"stripns":              stripns,
		"replaceReservedWords": replaceReservedWords,
		"makePublic":           g.makePublicFn,
		"findType":             g.findType,
		"comment":              comment,
	}

	data := new(bytes.Buffer)
	tmpl := template.Must(template.New("server_header").Funcs(funcMap).Parse(serverHeaderTmpl))
	err := tmpl.Execute(data, g.pkg)
	if err != nil {
		return nil, err
	}

	return data.Bytes(), nil
}

var reservedWords = map[string]string{
	"break":       "break_",
	"default":     "default_",
	"func":        "func_",
	"interface":   "interface_",
	"select":      "select_",
	"case":        "case_",
	"defer":       "defer_",
	"go":          "go_",
	"map":         "map_",
	"struct":      "struct_",
	"chan":        "chan_",
	"else":        "else_",
	"goto":        "goto_",
	"package":     "package_",
	"switch":      "switch_",
	"const":       "const_",
	"fallthrough": "fallthrough_",
	"if":          "if_",
	"range":       "range_",
	"type":        "type_",
	"continue":    "continue_",
	"for":         "for_",
	"import":      "import_",
	"return":      "return_",
	"var":         "var_",
}

var reservedWordsInAttr = map[string]string{
	"break":       "break_",
	"default":     "default_",
	"func":        "func_",
	"interface":   "interface_",
	"select":      "select_",
	"case":        "case_",
	"defer":       "defer_",
	"go":          "go_",
	"map":         "map_",
	"struct":      "struct_",
	"chan":        "chan_",
	"else":        "else_",
	"goto":        "goto_",
	"package":     "package_",
	"switch":      "switch_",
	"const":       "const_",
	"fallthrough": "fallthrough_",
	"if":          "if_",
	"range":       "range_",
	"type":        "type_",
	"continue":    "continue_",
	"for":         "for_",
	"import":      "import_",
	"return":      "return_",
	"var":         "var_",
	"string":      "astring",
}

var specialCharacterMapping = map[string]string{
	"+": "Plus",
	"@": "At",
}

// Replaces Go reserved keywords to avoid compilation issues
func replaceReservedWords(identifier string) string {
	value := reservedWords[identifier]
	if value != "" {
		return value
	}
	return normalize(identifier)
}

// Replaces Go reserved keywords to avoid compilation issues
func replaceAttrReservedWords(identifier string) string {
	value := reservedWordsInAttr[identifier]
	if value != "" {
		return value
	}
	return normalize(identifier)
}

// Normalizes value to be used as a valid Go identifier, avoiding compilation issues
func normalize(value string) string {
	for k, v := range specialCharacterMapping {
		value = strings.ReplaceAll(value, k, v)
	}

	mapping := func(r rune) rune {
		if r == '.' || r == '-' {
			return '_'
		}
		if unicode.IsLetter(r) || unicode.IsDigit(r) || r == '_' {
			return r
		}
		return -1
	}

	return strings.Map(mapping, value)
}

func goString(s string) string {
	return strings.ReplaceAll(s, "\"", "\\\"")
}

var xsd2GoTypes = map[string]string{
	"string":             "string",
	"token":              "string",
	"float":              "float32",
	"double":             "float64",
	"decimal":            "float64",
	"integer":            "int32",
	"int":                "int32",
	"short":              "int16",
	"byte":               "int8",
	"long":               "int64",
	"boolean":            "bool",
	"datetime":           "soap.XSDDateTime",
	"date":               "soap.XSDDate",
	"time":               "soap.XSDTime",
	"base64binary":       "[]byte",
	"hexbinary":          "[]byte",
	"unsignedint":        "uint32",
	"nonnegativeinteger": "uint32",
	"unsignedshort":      "uint16",
	"unsignedbyte":       "byte",
	"unsignedlong":       "uint64",
	"anytype":            "AnyType",
	"ncname":             "NCName",
	"anyuri":             "AnyURI",
}

func removeNS(xsdType string) string {
	// Handles name space, ie. xsd:string, xs:string
	r := strings.Split(xsdType, ":")

	if len(r) == 2 {
		return r[1]
	}

	return r[0]
}

func toGoType(xsdType string, nillable bool) string {
	// Handles name space, ie. xsd:string, xs:string
	r := strings.Split(xsdType, ":")

	t := r[0]

	if len(r) == 2 {
		t = r[1]
	}

	value := xsd2GoTypes[strings.ToLower(t)]

	if value != "" {
		if nillable {
			value = "*" + value
		}
		return value
	}

	return "*" + replaceReservedWords(makePublic(t))
}

func removePointerFromType(goType string) string {
	return regexp.MustCompile(`^\s*\*`).ReplaceAllLiteralString(goType, "")
}

// Given a message, finds its type.
//
// I'm not very proud of this function but
// it works for now and performance doesn't
// seem critical at this point
func (g *GoWSDL) findType(message string) string {
	message = stripns(message)

	for _, msg := range g.wsdl.Messages {
		if msg.Name != message {
			continue
		}

		// Assumes document/literal wrapped WS-I
		if len(msg.Parts) == 0 {
			// Message does not have parts. This could be a Port
			// with HTTP binding or SOAP 1.2 binding, which are not currently
			// supported.
			log.Printf("[WARN] %s message doesn't have any parts, ignoring message...", msg.Name)
			continue
		}

		part := msg.Parts[0]
		if part.Type != "" {
			return stripns(part.Type)
		}

		elRef := stripns(part.Element)

		for _, schema := range g.wsdl.Types.Schemas {
			for _, el := range schema.Elements {
				if strings.EqualFold(elRef, el.Name) {
					if el.Type != "" {
						return stripns(el.Type)
					}
					return el.Name
				}
			}
		}
	}
	return ""
}

// Given a type, check if there's an Element with that type, and return its name.
func (g *GoWSDL) findNameByType(name string) string {
	return newTraverser(nil, g.wsdl.Types.Schemas).findNameByType(name)
}

// TODO(c4milo): Add support for namespaces instead of striping them out
// TODO(c4milo): improve runtime complexity if performance turns out to be an issue.
func (g *GoWSDL) findSOAPAction(operation, portType string) string {
	for _, binding := range g.wsdl.Binding {
		if !strings.EqualFold(stripns(binding.Type), portType) {
			continue
		}

		for _, soapOp := range binding.Operations {
			if soapOp.Name == operation {
				return soapOp.SOAPOperation.SOAPAction
			}
		}
	}
	return ""
}

func (g *GoWSDL) findServiceAddress(name string) string {
	for _, service := range g.wsdl.Service {
		for _, port := range service.Ports {
			if port.Name == name {
				return port.SOAPAddress.Location
			}
		}
	}
	return ""
}

// stripns extracts the local part of a qualified name (for backward compatibility)
// Deprecated: Use GoWSDL.resolveType instead for proper namespace handling
func stripns(xsdType string) string {
	r := strings.Split(xsdType, ":")
	t := r[0]

	if len(r) == 2 {
		t = r[1]
	}

	return t
}

// resolveType resolves a qualified type name to its local name and namespace
func (g *GoWSDL) resolveType(qname string) (localName, namespace string) {
	if g.namespaceManager == nil {
		// Fallback for backward compatibility
		return stripns(qname), ""
	}
	
	return g.namespaceManager.ResolveQName(qname, g.currentNamespace)
}

func makePublic(identifier string) string {
	if isBasicType(identifier) {
		return identifier
	}
	if identifier == "" {
		return "EmptyString"
	}
	field := []rune(identifier)
	if len(field) == 0 {
		return identifier
	}

	field[0] = unicode.ToUpper(field[0])
	return string(field)
}

var basicTypes = map[string]string{
	"string":      "string",
	"float32":     "float32",
	"float64":     "float64",
	"int":         "int",
	"int8":        "int8",
	"int16":       "int16",
	"int32":       "int32",
	"int64":       "int64",
	"bool":        "bool",
	"time.Time":   "time.Time",
	"[]byte":      "[]byte",
	"byte":        "byte",
	"uint16":      "uint16",
	"uint32":      "uint32",
	"uinit64":     "uint64",
	"interface{}": "interface{}",
}

func isBasicType(identifier string) bool {
	if _, exists := basicTypes[identifier]; exists {
		return true
	}
	return false
}

func makePrivate(identifier string) string {
	field := []rune(identifier)
	if len(field) == 0 {
		return identifier
	}

	field[0] = unicode.ToLower(field[0])
	return string(field)
}

func comment(text string) string {
	lines := strings.Split(text, "\n")

	var output string
	if len(lines) == 1 && lines[0] == "" {
		return ""
	}

	// Helps to determine if there is an actual comment without screwing newlines
	// in real comments.
	hasComment := false

	for _, line := range lines {
		line = strings.TrimLeftFunc(line, unicode.IsSpace)
		if line != "" {
			hasComment = true
		}
		output += "\n// " + line
	}

	if hasComment {
		return output
	}
	return ""
}
