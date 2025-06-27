// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at http://mozilla.org/MPL/2.0/.

package templates

// Unused template - kept for potential future use
var _ = `
{{range .}}
	{{$privateType := .Name | makePrivate}}
	{{$exportType := .Name | makePublic}}

	type {{$exportType}} interface {
		{{range .Operations}}
			{{$inFaults := len .InFault}}
			{{$outFaults := len .OutFault}}
			{{$soapAction := findSOAPAction .Name $privateType}}
			{{$requestType := ""}}
			{{$responseType := ""}}
			{{if .Input}}{{$requestType = findType .Input.Element | replaceReservedWords | makePublic}}{{end}}
			{{if .Output}}{{$responseType = findType .Output.Element | replaceReservedWords | makePublic}}{{end}}

			{{/*if ne $soapAction ""*/}}
			{{if or (gt $inFaults 0) (gt $outFaults 0)}}
			// Error can be either of the following types:
			// {{range .InFault}}
			//   - {{.Ref}} (input fault){{end}}{{range .OutFault}}
			//   - {{.Ref}} (output fault){{end}}{{end}}
			{{if ne .Doc ""}}/* {{.Doc}} */{{end}}
			{{makePublic .Name | replaceReservedWords}} ({{if ne $requestType ""}}request *{{$requestType}}{{end}}) ({{if ne $responseType ""}}*{{$responseType}}, {{end}}error)
			{{/*end*/}}
			{{makePublic .Name | replaceReservedWords}}Context (ctx context.Context, {{if ne $requestType ""}}request *{{$requestType}}{{end}}) ({{if ne $responseType ""}}*{{$responseType}}, {{end}}error)
			{{/*end*/}}
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
		{{if .Input}}{{$requestType = findType .Input.Element | replaceReservedWords | makePublic}}{{end}}
		{{if .Output}}{{$responseType = findType .Output.Element | replaceReservedWords | makePublic}}{{end}}
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