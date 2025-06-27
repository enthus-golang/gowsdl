// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at http://mozilla.org/MPL/2.0/.

package templates

// TypesTemplate is the template for generating type definitions
const TypesTemplate = `
{{define "SimpleType"}}
	{{$type := replaceReservedWords .Name | makePublic}}
	{{if ne .List.ItemType ""}}
		type {{$type}} []{{toGoType .List.ItemType false}}
	{{else if ne .Union.MemberTypes ""}}
		type {{$type}} string
	{{else if .Restriction.Base}}
		{{if .Restriction.Enumeration}}
			type {{$type}} {{toGoType .Restriction.Base false}}

			const (
				{{range .Restriction.Enumeration}}
					{{$enumName := .Value | sanitizeEnumValue | replaceReservedWords | makePublic}}
					{{$type}}{{$enumName}} {{$type}} = "{{goString .Value}}" {{end}}
			)
		{{else}}
			type {{$type}} {{toGoType .Restriction.Base false}}
		{{end}}
	{{end}}
{{end}}

{{define "ComplexTypeInline"}}
	{{with .ComplexContent}}
		{{if ne .Extension.Base ""}}
			{{if eq .Extension.Base "soap:Header"}}
				soap.Header
			{{else}}
				{{removePointerFromType .Extension.Base | removeNamespacePrefix}}
			{{end}}
		{{end}}
	{{end}}

	{{range .Sequence}}
		{{$memberName := .Name | replaceReservedWords | makePublic}}
		{{if ne .Ref ""}}
			{{$memberName = .Ref | removeNamespacePrefix | replaceReservedWords | makePublic}}
			{{$typeName := .Ref | removeNamespacePrefix}}
			{{$memberType := toGoType $typeName .Nillable}}
			{{if eq .MaxOccurs "unbounded"}}
				{{$memberName}} []{{$memberType}} ` + "`" + `xml:"{{.Name}},omitempty" json:"{{.Name}},omitempty"` + "`" + `
			{{else}}
				{{$memberName}} {{$memberType}} ` + "`" + `xml:"{{.Name}},omitempty" json:"{{.Name}},omitempty"` + "`" + `
			{{end}}
		{{else if .ComplexType}}
			{{/* Handle inline complex types */}}
			{{if .ComplexType.SimpleContent}}
				{{/* Simple content with attributes - generate inline struct */}}
				{{if eq .MaxOccurs "unbounded"}}
					{{$memberName}} []struct {
						Value {{toGoType .ComplexType.SimpleContent.Extension.Base false}} ` + "`" + `xml:",chardata" json:"-,"` + "`" + `
						{{range .ComplexType.SimpleContent.Extension.Attributes}}
							{{$attrName := .Name | replaceReservedWords | makePublic}}
							{{$attrName}} {{toGoType .Type false}} ` + "`" + `xml:"{{getTargetNamespace}} {{.Name}},attr,omitempty" json:"{{.Name}},omitempty"` + "`" + `
						{{end}}
					} ` + "`" + `xml:"{{.Name}},omitempty" json:"{{.Name}},omitempty"` + "`" + `
				{{else}}
					{{$memberName}} struct {
						Value {{toGoType .ComplexType.SimpleContent.Extension.Base false}} ` + "`" + `xml:",chardata" json:"-,"` + "`" + `
						{{range .ComplexType.SimpleContent.Extension.Attributes}}
							{{$attrName := .Name | replaceReservedWords | makePublic}}
							{{$attrName}} {{toGoType .Type false}} ` + "`" + `xml:"{{getTargetNamespace}} {{.Name}},attr,omitempty" json:"{{.Name}},omitempty"` + "`" + `
						{{end}}
					} ` + "`" + `xml:"{{.Name}},omitempty" json:"{{.Name}},omitempty"` + "`" + `
				{{end}}
			{{else}}
				{{/* Regular complex type */}}
				{{$typeName := .Name}}
				{{$memberType := toGoType $typeName .Nillable}}
				{{if eq .MaxOccurs "unbounded"}}
					{{$memberName}} []{{$memberType}} ` + "`" + `xml:"{{.Name}},omitempty" json:"{{.Name}},omitempty"` + "`" + `
				{{else}}
					{{$memberName}} {{$memberType}} ` + "`" + `xml:"{{.Name}},omitempty" json:"{{.Name}},omitempty"` + "`" + `
				{{end}}
			{{end}}
		{{else if eq .Type ""}}
			{{$typeName := .Name}}
			{{$memberType := toGoType $typeName .Nillable}}
			{{if eq .MaxOccurs "unbounded"}}
				{{$memberName}} []{{$memberType}} ` + "`" + `xml:"{{.Name}},omitempty" json:"{{.Name}},omitempty"` + "`" + `
			{{else}}
				{{$memberName}} {{$memberType}} ` + "`" + `xml:"{{.Name}},omitempty" json:"{{.Name}},omitempty"` + "`" + `
			{{end}}
		{{else}}
			{{$typeName := .Type | removeNamespacePrefix}}
			{{$memberType := toGoType $typeName .Nillable}}
			{{if eq .MaxOccurs "unbounded"}}
				{{$memberName}} []{{$memberType}} ` + "`" + `xml:"{{.Name}},omitempty" json:"{{.Name}},omitempty"` + "`" + `
			{{else}}
				{{$memberName}} {{$memberType}} ` + "`" + `xml:"{{.Name}},omitempty" json:"{{.Name}},omitempty"` + "`" + `
			{{end}}
		{{end}}
	{{end}}

	{{range .Attributes}}
		
		{{$name := .Name}}
		{{if eq .Name "Id"}}
			{{$name = "IdAttr"}}
		{{end}}
		{{if eq .Name "id"}}
			{{$name = "idAttr"}}
		{{end}}
		{{replaceAttrReservedWords $name | makePublic}} {{toGoType .Type false}} ` + "`" + `xml:"{{getTargetNamespace}} {{.Name}},attr,omitempty" json:"{{.Name}},omitempty"` + "`" + `
	{{end}}

	{{if ne (len .Any) 0}}
		Items []string ` + "`" + `xml:",any"` + "`" + `
	{{end}}
{{end}}

{{define "ComplexType"}}
	{{$type := replaceReservedWords .Name | makePublic}}
	{{if .Abstract}}
		// {{$type}} is abstract
	{{end}}
	type {{$type}} struct {
		{{template "ComplexTypeInline" .}}
		{{with .ComplexContent}}
			{{if ne .Extension.Base ""}}
				{{if eq .Extension.Base "soap:Header"}}
					soap.Header
				{{else}}
					{{removePointerFromType .Extension.Base | removeNamespacePrefix}}
				{{end}}
			{{end}}
			{{template "Elements" .Extension.Sequence}}
			{{range .Extension.Attributes}}
				{{replaceAttrReservedWords .Name | makePublic}} {{toGoType .Type false}} ` + "`" + `xml:"{{.Name}},attr,omitempty"` + "`" + `
			{{end}}
		{{end}}
		{{with .SimpleContent}}
			{{if ne .Extension.Base ""}}
				Value {{toGoType .Extension.Base false}} ` + "`" + `xml:",chardata"` + "`" + `
			{{end}}
			{{range .Extension.Attributes}}
				
				{{replaceAttrReservedWords .Name | makePublic}} {{toGoType .Type false}} ` + "`" + `xml:"{{.Name}},attr,omitempty"` + "`" + `
			{{end}}
		{{end}}
	}
{{end}}

{{range .SimpleType}}
	{{template "SimpleType" .}}
{{end}}

{{range .ComplexTypes}}
	{{template "ComplexType" .}}
{{end}}

{{range .Elements}}
	{{$type := replaceReservedWords .Name | makePublic}}
	{{if ne .Type ""}}
		{{$goType := toGoType .Type .Nillable | removePointerFromType}}
		type {{$type}} {{$goType}}
		{{if eq $goType "soap.XSDDateTime"}}
			func (xdt {{$type}}) MarshalXML(e *xml.Encoder, start xml.StartElement) error {
				return soap.XSDDateTime(xdt).MarshalXML(e, start)
			}

			func (xdt *{{$type}}) UnmarshalXML(d *xml.Decoder, start xml.StartElement) error {
				return (*soap.XSDDateTime)(xdt).UnmarshalXML(d, start)
			}
		{{else if eq $goType "soap.XSDDate"}}
			func (xd {{$type}}) MarshalXML(e *xml.Encoder, start xml.StartElement) error {
				return soap.XSDDate(xd).MarshalXML(e, start)
			}

			func (xd *{{$type}}) UnmarshalXML(d *xml.Decoder, start xml.StartElement) error {
				return (*soap.XSDDate)(xd).UnmarshalXML(d, start)
			}
		{{else if eq $goType "soap.XSDTime"}}
			func (xt {{$type}}) MarshalXML(e *xml.Encoder, start xml.StartElement) error {
				return soap.XSDTime(xt).MarshalXML(e, start)
			}

			func (xt *{{$type}}) UnmarshalXML(d *xml.Decoder, start xml.StartElement) error {
				return (*soap.XSDTime)(xt).UnmarshalXML(d, start)
			}
		{{end}}
	{{else}}
		{{if .ComplexType}}
			{{/* ComplexTypeLocal handles global elements with nested complex types */}}
			{{$outerType := .Name | makePublic | replaceReservedWords}}
			{{if .ComplexType.Name}}
				{{$innerType := .ComplexType.Name | makePublic | replaceReservedWords}}
				type {{$outerType}} struct {
					{{$innerType}}
				}
				{{template "ComplexType" .ComplexType}}
			{{else}}
				type {{$outerType}} struct {
					XMLName xml.Name ` + "`" + `xml:"{{$.TargetNamespace}} {{.Name}}"` + "`" + `
					{{template "ComplexTypeInline" .ComplexType}}
				}
			{{end}}
		{{else}}
			{{if .SimpleType}}
				{{/* SimpleType local handles global elements with nested simple types */}}
				{{with .SimpleType}}
					{{if ne .Union.MemberTypes ""}}
						type {{$type}} string
					{{else if .Restriction.Base}}
						{{if .Restriction.Enumeration}}
							type {{$type}} {{toGoType .Restriction.Base false}}
							const (
								{{range .Restriction.Enumeration}}
									{{$enumName := .Value | sanitizeEnumValue | replaceReservedWords | makePublic}}
									{{$type}}{{$enumName}} {{$type}} = "{{goString .Value}}"
								{{end}}
							)
						{{else}}
							type {{$type}} {{toGoType .Restriction.Base false}}
						{{end}}
					{{else}}
						type {{$type}} interface{}
					{{end}}
				{{end}}
			{{end}}
		{{end}}
	{{end}}
{{end}}

{{define "Elements"}}
	{{range .}}
		{{if ne .Ref ""}}
			{{removeNamespacePrefix .Ref | replaceReservedWords | makePublic}} {{if eq .MaxOccurs "unbounded"}}[]{{end}}{{toGoType .Ref .Nillable}} ` + "`" + `xml:"{{.Ref | removeNamespacePrefix}},omitempty"` + "`" + `
		{{else}}
			{{if not .Type}}
				{{if .ComplexType}}
					{{template "ComplexTypeInline" .}}
				{{end}}
			{{else}}
				{{replaceReservedWords .Name | makePublic}} {{if eq .MaxOccurs "unbounded"}}[]{{end}}{{toGoType .Type .Nillable}} ` + "`" + `xml:"{{.Name}},omitempty"` + "`" + `
			{{end}}
		{{end}}
	{{end}}
{{end}}
`