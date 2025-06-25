// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at http://mozilla.org/MPL/2.0/.

package templates

import (
	"text/template"
)

// Manager manages all code generation templates
type Manager struct {
	funcMap template.FuncMap
}

// NewManager creates a new template manager
func NewManager(funcMap template.FuncMap) *Manager {
	return &Manager{
		funcMap: funcMap,
	}
}

// GetOperationsTemplate returns the operations template
func (m *Manager) GetOperationsTemplate() (*template.Template, error) {
	return template.New("operations").Funcs(m.funcMap).Parse(UnifiedOperationsTemplate)
}

// GetTypesTemplate returns the types template
func (m *Manager) GetTypesTemplate() (*template.Template, error) {
	// TODO: Add TypesTemplate constant
	return template.New("types").Funcs(m.funcMap).Parse("")
}

// GetHeaderTemplate returns the header template
func (m *Manager) GetHeaderTemplate() (*template.Template, error) {
	// TODO: Add HeaderTemplate constant
	return template.New("header").Funcs(m.funcMap).Parse("")
}

// GetServerTemplate returns the server template
func (m *Manager) GetServerTemplate() (*template.Template, error) {
	// TODO: Add UnifiedServerTemplate constant
	return template.New("server").Funcs(m.funcMap).Parse("")
}

// GetGenericOperationsTemplate returns the generic operations template
func (m *Manager) GetGenericOperationsTemplate() (*template.Template, error) {
	// TODO: Add GenericOperationsTemplate constant
	return template.New("generic_operations").Funcs(m.funcMap).Parse("")
}