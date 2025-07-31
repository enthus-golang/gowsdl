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

{{define "ComplexTypeInlineNamed"}}
	{{$parentTypeName := .ParentTypeName}}
	{{$complexType := .ComplexType}}
	{{with $complexType.ComplexContent}}
		{{if ne .Extension.Base ""}}
			{{if eq .Extension.Base "soap:Header"}}
				soap.Header
			{{else}}
				{{removePointerFromType .Extension.Base | removeNamespacePrefix}}
			{{end}}
		{{end}}
	{{end}}

	{{range $complexType.Sequence}}
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
			{{$inlineTypeName := getInlineTypeName $parentTypeName .Name}}
			{{if eq .MaxOccurs "unbounded"}}
				{{$memberName}} []{{$inlineTypeName}} ` + "`" + `xml:"{{.Name}},omitempty" json:"{{.Name}},omitempty"` + "`" + `
			{{else if and (eq .MinOccurs "0") (ne .MaxOccurs "unbounded")}}
				{{$memberName}} *{{$inlineTypeName}} ` + "`" + `xml:"{{.Name}},omitempty" json:"{{.Name}},omitempty"` + "`" + `
			{{else}}
				{{$memberName}} {{$inlineTypeName}} ` + "`" + `xml:"{{.Name}},omitempty" json:"{{.Name}},omitempty"` + "`" + `
			{{end}}
		{{else if .Type}}
			{{$typeName := .Type | removeNamespacePrefix}}
			{{$memberType := toGoType $typeName .Nillable}}
			{{if eq .MaxOccurs "unbounded"}}
				{{$memberName}} []{{$memberType}} ` + "`" + `xml:"{{.Name}},omitempty" json:"{{.Name}},omitempty"` + "`" + `
			{{else}}
				{{$memberName}} {{$memberType}} ` + "`" + `xml:"{{.Name}},omitempty" json:"{{.Name}},omitempty"` + "`" + `
			{{end}}
		{{else}}
			{{$memberName}} string ` + "`" + `xml:"{{.Name}},omitempty" json:"{{.Name}},omitempty"` + "`" + `
		{{end}}
	{{end}}

	{{range $complexType.Choice}}
		{{$memberName := .Name | replaceReservedWords | makePublic}}
		{{if .ComplexType}}
			{{$inlineTypeName := getInlineTypeName $parentTypeName .Name}}
			{{if eq .MaxOccurs "unbounded"}}
				{{$memberName}} []{{$inlineTypeName}} ` + "`" + `xml:"{{.Name}},omitempty" json:"{{.Name}},omitempty"` + "`" + `
			{{else if and (eq .MinOccurs "0") (ne .MaxOccurs "unbounded")}}
				{{$memberName}} *{{$inlineTypeName}} ` + "`" + `xml:"{{.Name}},omitempty" json:"{{.Name}},omitempty"` + "`" + `
			{{else}}
				{{$memberName}} {{$inlineTypeName}} ` + "`" + `xml:"{{.Name}},omitempty" json:"{{.Name}},omitempty"` + "`" + `
			{{end}}
		{{else if .Type}}
			{{$typeName := .Type | removeNamespacePrefix}}
			{{$memberType := toGoType $typeName .Nillable}}
			{{if eq .MaxOccurs "unbounded"}}
				{{$memberName}} []{{$memberType}} ` + "`" + `xml:"{{.Name}},omitempty" json:"{{.Name}},omitempty"` + "`" + `
			{{else}}
				{{$memberName}} {{$memberType}} ` + "`" + `xml:"{{.Name}},omitempty" json:"{{.Name}},omitempty"` + "`" + `
			{{end}}
		{{else}}
			{{$memberName}} string ` + "`" + `xml:"{{.Name}},omitempty" json:"{{.Name}},omitempty"` + "`" + `
		{{end}}
	{{end}}

	{{range $complexType.All}}
		{{$memberName := .Name | replaceReservedWords | makePublic}}
		{{if .ComplexType}}
			{{$inlineTypeName := getInlineTypeName $parentTypeName .Name}}
			{{if eq .MaxOccurs "unbounded"}}
				{{$memberName}} []{{$inlineTypeName}} ` + "`" + `xml:"{{.Name}},omitempty" json:"{{.Name}},omitempty"` + "`" + `
			{{else if and (eq .MinOccurs "0") (ne .MaxOccurs "unbounded")}}
				{{$memberName}} *{{$inlineTypeName}} ` + "`" + `xml:"{{.Name}},omitempty" json:"{{.Name}},omitempty"` + "`" + `
			{{else}}
				{{$memberName}} {{$inlineTypeName}} ` + "`" + `xml:"{{.Name}},omitempty" json:"{{.Name}},omitempty"` + "`" + `
			{{end}}
		{{else if .Type}}
			{{$typeName := .Type | removeNamespacePrefix}}
			{{$memberType := toGoType $typeName .Nillable}}
			{{if eq .MaxOccurs "unbounded"}}
				{{$memberName}} []{{$memberType}} ` + "`" + `xml:"{{.Name}},omitempty" json:"{{.Name}},omitempty"` + "`" + `
			{{else}}
				{{$memberName}} {{$memberType}} ` + "`" + `xml:"{{.Name}},omitempty" json:"{{.Name}},omitempty"` + "`" + `
			{{end}}
		{{else}}
			{{$memberName}} string ` + "`" + `xml:"{{.Name}},omitempty" json:"{{.Name}},omitempty"` + "`" + `
		{{end}}
	{{end}}

	{{range $complexType.Attributes}}
		{{$name := .Name}}
		{{if eq .Name "Id"}}
			{{$name = "IdAttr"}}
		{{end}}
		{{if eq .Name "id"}}
			{{$name = "idAttr"}}
		{{end}}
		{{replaceAttrReservedWords $name | makePublic}} {{toGoType .Type false}} ` + "`" + `xml:"{{.Name}},attr,omitempty" json:"{{.Name}},omitempty"` + "`" + `
	{{end}}

	{{if ne (len $complexType.Any) 0}}
		Items []string ` + "`" + `xml:",any"` + "`" + `
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
			{{if and (eq .MinOccurs "0") (ne .MaxOccurs "unbounded")}}
				{{/* Optional complex type - use pointer to named type */}}
				{{$inlineTypeName := getInlineTypeName $.Name .Name}}
				{{$memberName}} *{{$inlineTypeName}} ` + "`" + `xml:"{{.Name}},omitempty" json:"{{.Name}},omitempty"` + "`" + `
			{{else if .ComplexType.SimpleContent.Extension.Base}}
				{{/* Simple content with attributes - use named type */}}
				{{$inlineTypeName := getInlineTypeName $.Name .Name}}
				{{if eq .MaxOccurs "unbounded"}}
					{{$memberName}} []{{$inlineTypeName}} ` + "`" + `xml:"{{.Name}},omitempty" json:"{{.Name}},omitempty"` + "`" + `
				{{else if and (eq .MinOccurs "0") (ne .MaxOccurs "unbounded")}}
					{{$memberName}} *{{$inlineTypeName}} ` + "`" + `xml:"{{.Name}},omitempty" json:"{{.Name}},omitempty"` + "`" + `
				{{else}}
					{{$memberName}} {{$inlineTypeName}} ` + "`" + `xml:"{{.Name}},omitempty" json:"{{.Name}},omitempty"` + "`" + `
				{{end}}
			{{else if or .ComplexType.Sequence .ComplexType.Choice .ComplexType.All}}
				{{/* Complex type with sequence/choice/all - use named type */}}
				{{$inlineTypeName := getInlineTypeName $.Name .Name}}
				{{if eq .MaxOccurs "unbounded"}}
					{{$memberName}} []{{$inlineTypeName}} ` + "`" + `xml:"{{.Name}},omitempty" json:"{{.Name}},omitempty"` + "`" + `
				{{else if and (eq .MinOccurs "0") (ne .MaxOccurs "unbounded")}}
					{{$memberName}} *{{$inlineTypeName}} ` + "`" + `xml:"{{.Name}},omitempty" json:"{{.Name}},omitempty"` + "`" + `
				{{else}}
					{{$memberName}} {{$inlineTypeName}} ` + "`" + `xml:"{{.Name}},omitempty" json:"{{.Name}},omitempty"` + "`" + `
				{{end}}
			{{else}}
				{{/* Regular complex type without sequence - fallback to using element name as type */}}
				{{$typeName := .Name}}
				{{$memberType := toGoType $typeName .Nillable}}
				{{if eq .MaxOccurs "unbounded"}}
					{{$memberName}} []{{$memberType}} ` + "`" + `xml:"{{.Name}},omitempty" json:"{{.Name}},omitempty"` + "`" + `
				{{else}}
					{{$memberName}} {{$memberType}} ` + "`" + `xml:"{{.Name}},omitempty" json:"{{.Name}},omitempty"` + "`" + `
				{{end}}
			{{end}}
		{{else if eq .Type ""}}
			{{if .SimpleType}}
				{{/* Element has inline simple type - use its base type */}}
				{{if .SimpleType.Restriction.Base}}
					{{$baseType := .SimpleType.Restriction.Base | removeNamespacePrefix}}
					{{$memberType := toGoType $baseType .Nillable}}
					{{if eq .MaxOccurs "unbounded"}}
						{{$memberName}} []{{$memberType}} ` + "`" + `xml:"{{.Name}},omitempty" json:"{{.Name}},omitempty"` + "`" + `
					{{else}}
						{{$memberName}} {{$memberType}} ` + "`" + `xml:"{{.Name}},omitempty" json:"{{.Name}},omitempty"` + "`" + `
					{{end}}
				{{else}}
					{{/* Fallback if no restriction base */}}
					{{$memberName}} string ` + "`" + `xml:"{{.Name}},omitempty" json:"{{.Name}},omitempty"` + "`" + `
				{{end}}
			{{else}}
				{{/* No type and no inline simple type - use element name as type */}}
				{{$typeName := .Name}}
				{{$memberType := toGoType $typeName .Nillable}}
				{{if eq .MaxOccurs "unbounded"}}
					{{$memberName}} []{{$memberType}} ` + "`" + `xml:"{{.Name}},omitempty" json:"{{.Name}},omitempty"` + "`" + `
				{{else}}
					{{$memberName}} {{$memberType}} ` + "`" + `xml:"{{.Name}},omitempty" json:"{{.Name}},omitempty"` + "`" + `
				{{end}}
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

	{{range .Choice}}
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
		{{else if .Type}}
			{{$typeName := .Type | removeNamespacePrefix}}
			{{$memberType := toGoType $typeName .Nillable}}
			{{if eq .MaxOccurs "unbounded"}}
				{{$memberName}} []{{$memberType}} ` + "`" + `xml:"{{.Name}},omitempty" json:"{{.Name}},omitempty"` + "`" + `
			{{else}}
				{{$memberName}} {{$memberType}} ` + "`" + `xml:"{{.Name}},omitempty" json:"{{.Name}},omitempty"` + "`" + `
			{{end}}
		{{end}}
	{{end}}

	{{range .All}}
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
		{{else if .Type}}
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
		{{replaceAttrReservedWords $name | makePublic}} {{toGoType .Type false}} ` + "`" + `xml:"{{.Name}},attr,omitempty" json:"{{.Name}},omitempty"` + "`" + `
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
		{{if isMessageElement .Name}}
			{{if isElementFormQualified}}
			XMLName xml.Name ` + "`" + `xml:"{{getTargetNamespace}} {{.Name}}"` + "`" + `
			{{else if isResponseElement .Name}}
			{{/* Response elements use local name only to handle any namespace prefix */}}
			XMLName xml.Name ` + "`" + `xml:"{{.Name}}"` + "`" + `
			{{else}}
			XMLName xml.Name ` + "`" + `xml:"tns:{{.Name}}"` + "`" + `
			{{end}}
		{{end}}
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
	
	{{if isMessageElement .Name}}
		{{if not isElementFormQualified}}
		// TargetNamespace returns the target namespace for SOAP envelope configuration
		func (t {{$type}}) TargetNamespace() string {
			return "{{getTargetNamespace}}"
		}
		{{end}}
	{{end}}
{{end}}

{{range .SimpleType}}
	{{template "SimpleType" .}}
{{end}}

{{range .ComplexTypes}}
	{{template "ComplexType" .}}
{{end}}

{{/* Generate named types for inline complex types */}}
{{range $key, $inline := .InlineTypes}}
	{{$type := $inline.GeneratedName}}
	type {{$type}} struct {
		{{if $inline.ComplexType.SimpleContent.Extension.Base}}
			{{/* Simple content with attributes */}}
			Value {{toGoType $inline.ComplexType.SimpleContent.Extension.Base false}} ` + "`" + `xml:",chardata" json:"-,"` + "`" + `
			{{range $inline.ComplexType.SimpleContent.Extension.Attributes}}
				{{$attrName := .Name | replaceReservedWords | makePublic}}
				{{$attrName}} {{toGoType .Type false}} ` + "`" + `xml:"{{.Name}},attr,omitempty" json:"{{.Name}},omitempty"` + "`" + `
			{{end}}
		{{else}}
			{{/* Pass the parent type name to handle nested inline types */}}
			{{template "ComplexTypeInlineNamed" (dict "ParentTypeName" $inline.GeneratedName "ComplexType" $inline.ComplexType)}}
		{{end}}
	}
{{end}}

{{range .Elements}}
	{{$type := replaceReservedWords .Name | makePublic}}
	{{if ne .Type ""}}
		{{$baseType := .Type | removeNamespacePrefix | replaceReservedWords}}
		{{$goType := toGoType .Type .Nillable | removePointerFromType}}
		{{$goPublicType := $baseType | makePublic}}
		{{/* Check if element name is different from the type name */}}
		{{$createAlias := true}}
		{{if eq .Name $baseType}}
			{{/* Don't create type alias if element name equals the type name */}}
			{{$createAlias = false}}
		{{end}}
		{{if $createAlias}}
			{{/* For complex types, use the public version of the type name */}}
			{{if isComplexType .Type}}
				type {{$type}} struct {
					{{if isElementFormQualified}}
					XMLName xml.Name ` + "`" + `xml:"{{$.TargetNamespace}} {{.Name}}"` + "`" + `
					{{else if isResponseElement .Name}}
					{{/* Response elements use local name only to handle any namespace prefix */}}
					XMLName xml.Name ` + "`" + `xml:"{{.Name}}"` + "`" + `
					{{else}}
					XMLName xml.Name ` + "`" + `xml:"tns:{{.Name}}"` + "`" + `
					{{end}}
					{{$goPublicType}}
				}
				
				{{if not isElementFormQualified}}
				func (t {{$type}}) TargetNamespace() string {
					return "{{$.TargetNamespace}}"
				}
				{{end}}
			{{else}}
				type {{$type}} {{$goType}}
			{{end}}
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
					{{if isElementFormQualified}}
					XMLName xml.Name ` + "`" + `xml:"{{$.TargetNamespace}} {{.Name}}"` + "`" + `
					{{else if isResponseElement .Name}}
					{{/* Response elements use local name only to handle any namespace prefix */}}
					XMLName xml.Name ` + "`" + `xml:"{{.Name}}"` + "`" + `
					{{else}}
					XMLName xml.Name ` + "`" + `xml:"tns:{{.Name}}"` + "`" + `
					{{end}}
					{{template "ComplexTypeInline" .ComplexType}}
				}
				
				{{if not isElementFormQualified}}
				func (t {{$outerType}}) TargetNamespace() string {
					return "{{$.TargetNamespace}}"
				}
				{{end}}
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

{{/* Generate wrapper types for RPC-style messages */}}
{{range .Messages}}
	{{if isRPCStyleMessage .}}
		{{$type := replaceReservedWords .Name | makePublic}}
		type {{$type}} struct {
			{{range .Parts}}
				{{$partName := .Name | replaceReservedWords | makePublic}}
				{{if ne .Type ""}}
					{{$partType := .Type | removeNamespacePrefix}}
					{{$partName}} {{toGoType $partType false}} ` + "`" + `xml:"{{.Name}}" json:"{{.Name}}"` + "`" + `
				{{else if ne .Element ""}}
					{{$elementType := .Element | removeNamespacePrefix}}
					{{$partName}} {{toGoType $elementType false}} ` + "`" + `xml:"{{.Name}}" json:"{{.Name}}"` + "`" + `
				{{end}}
			{{end}}
		}
	{{end}}
{{end}}

{{/* Generate RPC operation wrapper types */}}
{{range .PortTypes}}
	{{if isRPCPortType .Name}}
		{{$portType := .}}
		{{range .Operations}}
			{{$opName := .Name}}
			{{$inputMsgName := .Input.Message | removeNamespacePrefix}}
			{{$outputMsgName := .Output.Message | removeNamespacePrefix}}
			
			{{/* Find the input message and create operation wrapper */}}
			{{range $.Messages}}
				{{if eq .Name $inputMsgName}}
					// Operation wrapper for {{$opName | makePublic}}.
					type Operation{{$opName | makePublic}}In struct {
						XMLName xml.Name ` + "`" + `xml:"tns:{{$opName}}"` + "`" + `
						
						{{range .Parts}}
							{{$partName := .Name | replaceReservedWords | makePublic}}
							{{if ne .Type ""}}
								{{$partType := .Type | removeNamespacePrefix}}
								{{$partName}} *{{toGoType $partType false | removePointerFromType}} ` + "`" + `xml:"{{.Name}},omitempty" json:"{{.Name}},omitempty"` + "`" + `
							{{else if ne .Element ""}}
								{{$elementType := .Element | removeNamespacePrefix}}
								{{$partName}} *{{toGoType $elementType false | removePointerFromType}} ` + "`" + `xml:"{{.Name}},omitempty" json:"{{.Name}},omitempty"` + "`" + `
							{{end}}
						{{end}}
					}
				{{end}}
			{{end}}
			
			{{/* Find the output message and create operation response wrapper */}}
			{{range $.Messages}}
				{{if eq .Name $outputMsgName}}
					// Operation wrapper for {{$opName | makePublic}}.
					type Operation{{$opName | makePublic}}Out struct {
						XMLName xml.Name ` + "`" + `xml:"{{$opName}}Response"` + "`" + `
						
						{{range .Parts}}
							{{$partName := .Name | replaceReservedWords | makePublic}}
							{{if ne .Type ""}}
								{{$partType := .Type | removeNamespacePrefix}}
								{{if eq .Name "return"}}
									Return *{{toGoType $partType false | removePointerFromType}} ` + "`" + `xml:"return,omitempty" json:"return,omitempty"` + "`" + `
								{{else}}
									{{$partName}} *{{toGoType $partType false | removePointerFromType}} ` + "`" + `xml:"{{.Name}},omitempty" json:"{{.Name}},omitempty"` + "`" + `
								{{end}}
							{{else if ne .Element ""}}
								{{$elementType := .Element | removeNamespacePrefix}}
								{{if eq .Name "return"}}
									Return *{{toGoType $elementType false | removePointerFromType}} ` + "`" + `xml:"return,omitempty" json:"return,omitempty"` + "`" + `
								{{else}}
									{{$partName}} *{{toGoType $elementType false | removePointerFromType}} ` + "`" + `xml:"{{.Name}},omitempty" json:"{{.Name}},omitempty"` + "`" + `
								{{end}}
							{{end}}
						{{end}}
					}
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