// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at http://mozilla.org/MPL/2.0/.

package templates

// UnifiedOperationsTemplate is a consolidated template for both WSDL 1.1 and 2.0
const UnifiedOperationsTemplate = `
{{range .}}
	{{$privateType := .Name | makePrivate}}
	{{$exportType := .Name | makePublic}}

	type {{$exportType}} interface {
		{{range .Operations}}
			{{$soapAction := findSOAPAction .Name $privateType}}
			{{$requestType := ""}}
			{{$responseType := ""}}
			
			{{/* Handle WSDL version differences */}}
			{{if .Input}}
				{{if .Input.Message}}
					{{$requestType = findType .Input.Message | replaceReservedWords | makePublic}}
				{{else if .Input.Element}}
					{{$requestType = findType .Input.Element | replaceReservedWords | makePublic}}
				{{end}}
			{{end}}
			
			{{if .Output}}
				{{if .Output.Message}}
					{{$responseType = findType .Output.Message | replaceReservedWords | makePublic}}
				{{else if .Output.Element}}
					{{$responseType = findType .Output.Element | replaceReservedWords | makePublic}}
				{{end}}
			{{end}}

			{{/* Handle fault documentation */}}
			{{$hasFaults := false}}
			{{if .Faults}}{{$hasFaults = gt (len .Faults) 0}}{{end}}
			{{if .InFault}}{{$hasFaults = gt (len .InFault) 0}}{{end}}
			{{if .OutFault}}{{$hasFaults = gt (len .OutFault) 0}}{{end}}
			
			{{if $hasFaults}}
			// Error can be either of the following types:
			{{range .Faults}}
			//   - {{.Name}} {{.Doc}}{{end}}
			{{range .InFault}}
			//   - {{.Ref}} (input fault){{end}}
			{{range .OutFault}}
			//   - {{.Ref}} (output fault){{end}}
			{{end}}
			
			{{if ne .Doc ""}}/* {{.Doc}} */{{end}}
			{{makePublic .Name | replaceReservedWords}} ({{if ne $requestType ""}}request *{{$requestType}}{{end}}) ({{if ne $responseType ""}}*{{$responseType}}, {{end}}error)
			{{makePublic .Name | replaceReservedWords}}Context (ctx context.Context, {{if ne $requestType ""}}request *{{$requestType}}{{end}}) ({{if ne $responseType ""}}*{{$responseType}}, {{end}}error)
		{{end}}
	}

	type {{$privateType}} struct {
		client *soap.Client
	}

	func New{{$exportType}}(client *soap.Client) {{$exportType}} {
		return &{{$privateType}}{
			client: client,
		}
	}

	{{range .Operations}}
		{{$requestType := ""}}
		{{$responseType := ""}}
		
		{{/* Handle WSDL version differences */}}
		{{if .Input}}
			{{if .Input.Message}}
				{{$requestType = findType .Input.Message | replaceReservedWords | makePublic}}
			{{else if .Input.Element}}
				{{$requestType = findType .Input.Element | replaceReservedWords | makePublic}}
			{{end}}
		{{end}}
		
		{{if .Output}}
			{{if .Output.Message}}
				{{$responseType = findType .Output.Message | replaceReservedWords | makePublic}}
			{{else if .Output.Element}}
				{{$responseType = findType .Output.Element | replaceReservedWords | makePublic}}
			{{end}}
		{{end}}
		
		{{$soapAction := findSOAPAction .Name $privateType}}
		
		func (service *{{$privateType}}) {{makePublic .Name | replaceReservedWords}}Context (ctx context.Context, {{if ne $requestType ""}}request *{{$requestType}}{{end}}) ({{if ne $responseType ""}}*{{$responseType}}, {{end}}error) {
			{{if ne $responseType ""}}response := new({{$responseType}}){{end}}
			err := service.client.CallContext(ctx, "{{if ne $soapAction ""}}{{$soapAction}}{{else}}''{{end}}", {{if ne $requestType ""}}request{{else}}nil{{end}}, {{if ne $responseType ""}}response{{else}}struct{}{}{{end}})
			if err != nil {
				return {{if ne $responseType ""}}nil, {{end}}err
			}

			return {{if ne $responseType ""}}response, {{end}}nil
		}

		func (service *{{$privateType}}) {{makePublic .Name | replaceReservedWords}} ({{if ne $requestType ""}}request *{{$requestType}}{{end}}) ({{if ne $responseType ""}}*{{$responseType}}, {{end}}error) {
			return service.{{makePublic .Name | replaceReservedWords}}Context(
				context.Background(),
				{{if ne $requestType ""}}request,{{end}}
			)
		}

	{{end}}
{{end}}
`