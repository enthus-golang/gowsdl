// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at http://mozilla.org/MPL/2.0/.

package templates

// TypesTemplate is the template for generating type definitions
const TypesTemplate = `
{{define "SimpleType"}}
	{{$type := replaceReservedWords .Name | makePublic}}
	{{if .Doc}} {{comment .Doc}} {{end}}
	{{if ne .List.ItemType ""}}
		type {{$type}} []{{toGoType .List.ItemType false}}
	{{else if ne .Union.MemberTypes ""}}
		type {{$type}} string
	{{else if .Restriction.Base}}
		{{if .Restriction.Enumeration}}
			type {{$type}} {{toGoType .Restriction.Base false}}

			const (
				{{range .Restriction.Enumeration}}
					{{with .Doc}} {{comment .}} {{end}}
					{{$type}}_{{$value := replaceReservedWords .Value}}{{$value | makePublic}} {{$type}} = "{{goString .Value}}" {{end}}
			)
		{{else}}
			type {{$type}} {{toGoType .Restriction.Base false}}
		{{end}}
	{{end}}
{{end}}

{{define "ComplexTypeInline"}}
	{{$type := replaceReservedWords .Name | makePublic}}
	{{with .ComplexContent}}
		{{if ne .Extension.Base ""}}
			{{if eq .Extension.Base "soap:Header"}}
				soap.Header
			{{else}}
				{{removePointerFromType .Extension.Base | removeNamespacePrefix}}
			{{end}}
		{{end}}
	{{end}}

	{{if ne (len .Sequence) 0}}
		{{range .Sequence}}
			{{$memberName := .Name | makePublic | replaceReservedWords}}
			{{$typeName := ""}}
			{{if ne .Ref ""}}
				{{$memberName = .Ref | makePublic | replaceReservedWords | removeNamespacePrefix}}
				{{$typeName = .Ref | removeNamespacePrefix}}
			{{else if eq .Type ""}}
				{{$typeName = .Name}}
			{{else}}
				{{$typeName = .Type | removeNamespacePrefix}}
			{{end}}
			{{$memberType := toGoType $typeName .Nillable}}
			{{if .Doc}} {{comment .Doc}} {{end}}
			{{if ne .Min ""}}
				// min: {{.Min}}
			{{end}}
			{{if ne .Max ""}}
				{{if eq .Max "unbounded"}}
					{{if eq $memberType "string"}}
						XMLName xml.Name ` + "`" + `xml:"{{.Name}}"` + "`" + ` {{$memberName}} []{{$memberType}} ` + "`" + `xml:",chardata"` + "`" + `{{if ne .Type ""}} // {{.Type}} {{end}}
					{{else}}
						{{$memberName}} []{{$memberType}} ` + "`" + `xml:"{{.Name}},omitempty"` + "`" + `{{if ne .Type ""}} // {{.Type}} {{end}}
					{{end}}
				{{else}}
					// max: {{.Max}}
				{{end}}
			{{else}}
				{{$xmlTag := makeValidXmlTag .Name $memberName}}
				{{$memberName}} {{$memberType}} ` + "`" + `xml:"{{$xmlTag}},omitempty"` + "`" + `{{if ne .Type ""}} // {{.Type}} {{end}}
			{{end}}
		{{end}}
	{{end}}

	{{range .Attributes}}
		{{with .Doc}} {{comment .}} {{end}}
		{{$name := .Name}}
		{{if eq .Name "Id"}}
			{{$name = "IdAttr"}}
		{{end}}
		{{if eq .Name "id"}}
			{{$name = "idAttr"}}
		{{end}}
		{{replaceAttrReservedWords $name | makePublic}} {{toGoType .Type false}} ` + "`" + `xml:"{{.Name}},attr,omitempty"` + "`" + `
	{{end}}

	{{if ne (len .Any) 0}}
		Items []string ` + "`" + `xml:",any"` + "`" + `
	{{end}}
{{end}}

{{define "ComplexType"}}
	{{$type := replaceReservedWords .Name | makePublic}}
	{{if .Doc}} {{comment .Doc}} {{end}}
	{{if eq .Abstract "true"}}
		// {{$type}} is abstract
	{{end}}
	type {{$type}} struct {
		{{template "ComplexTypeInline" .}}
		{{with .ComplexContent}}
			{{with .Extension}}
				{{template "ComplexTypeInline" .}}
			{{end}}
		{{end}}
		{{with .SimpleContent}}
			{{if ne .Extension.Base ""}}
				Value {{toGoType .Extension.Base false}} ` + "`" + `xml:",chardata"` + "`" + `
			{{end}}
			{{range .Extension.Attributes}}
				{{with .Doc}} {{comment .}} {{end}}
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
		{{if .Doc}} {{comment .Doc}} {{end}}
		type {{$type}} {{toGoType .Type .Nillable | removePointerFromType}}
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
									{{with .Doc}} {{comment .}} {{end}}
									{{$type}}_{{replaceReservedWords .Value | makePublic}} {{$type}} = "{{goString .Value}}"
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
`