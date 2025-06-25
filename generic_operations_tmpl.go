// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at http://mozilla.org/MPL/2.0/.

package gowsdl

var genericOpsTmpl = `
{{range .}}
	{{$privateType := .Name | makePrivate}}
	{{$exportType := .Name | makePublic}}

	type {{$exportType}} interface {
		{{range .Operations}}
			{{$faults := len .Faults}}
			{{$soapAction := findSOAPAction .Name $privateType}}
			{{$requestType := findType .Input.Message | replaceReservedWords | makePublic}}
			{{$responseType := findType .Output.Message | replaceReservedWords | makePublic}}

			{{/*if ne $soapAction ""*/}}
			{{if gt $faults 0}}
			// Error can be either of the following types:
			// {{range .Faults}}
			//   - {{.Name}} {{.Doc}}{{end}}{{end}}
			{{if ne .Doc ""}}/* {{.Doc}} */{{end}}
			{{makePublic .Name | replaceReservedWords}} ({{if ne $requestType ""}}request *{{$requestType}}{{end}}) ({{if ne $responseType ""}}*{{$responseType}}, {{end}}error)
			{{/*end*/}}
			{{makePublic .Name | replaceReservedWords}}Context (ctx context.Context, {{if ne $requestType ""}}request *{{$requestType}}{{end}}) ({{if ne $responseType ""}}*{{$responseType}}, {{end}}error)
			{{/*end*/}}
			
			// Generic versions
			{{makePublic .Name | replaceReservedWords}}Generic ({{if ne $requestType ""}}request *{{$requestType}}{{end}}) ({{if ne $responseType ""}}soap.Result[{{$responseType}}], {{end}}error)
			{{makePublic .Name | replaceReservedWords}}GenericContext (ctx context.Context, {{if ne $requestType ""}}request *{{$requestType}}{{end}}) ({{if ne $responseType ""}}soap.Result[{{$responseType}}], {{end}}error)
		{{end}}
	}

	type {{$privateType}} struct {
		client *soap.Client
		{{range .Operations}}
		{{$soapAction := findSOAPAction .Name $privateType}}
		{{$requestType := findType .Input.Message | replaceReservedWords | makePublic}}
		{{$responseType := findType .Output.Message | replaceReservedWords | makePublic}}
		{{if and (ne $requestType "") (ne $responseType "")}}
		{{.Name | makePrivate}}Client *soap.GenericClient[{{$requestType}}, {{$responseType}}]
		{{end}}
		{{end}}
	}

	func New{{$exportType}}(client *soap.Client) {{$exportType}} {
		return &{{$privateType}}{
			client: client,
			{{range .Operations}}
			{{$soapAction := findSOAPAction .Name $privateType}}
			{{$requestType := findType .Input.Message | replaceReservedWords | makePublic}}
			{{$responseType := findType .Output.Message | replaceReservedWords | makePublic}}
			{{if and (ne $requestType "") (ne $responseType "")}}
			{{.Name | makePrivate}}Client: soap.NewGenericClient[{{$requestType}}, {{$responseType}}](client.URL(), "{{if ne $soapAction ""}}{{$soapAction}}{{else}}''{{end}}"),
			{{end}}
			{{end}}
		}
	}

	{{range .Operations}}
		{{$requestType := findType .Input.Message | replaceReservedWords | makePublic}}
		{{$soapAction := findSOAPAction .Name $privateType}}
		{{$responseType := findType .Output.Message | replaceReservedWords | makePublic}}
		
		// Standard implementation
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

		// Generic implementation
		{{if and (ne $requestType "") (ne $responseType "")}}
		func (service *{{$privateType}}) {{makePublic .Name | replaceReservedWords}}GenericContext (ctx context.Context, request *{{$requestType}}) (soap.Result[{{$responseType}}], error) {
			return service.{{.Name | makePrivate}}Client.CallAsResult(ctx, *request), nil
		}

		func (service *{{$privateType}}) {{makePublic .Name | replaceReservedWords}}Generic (request *{{$requestType}}) (soap.Result[{{$responseType}}], error) {
			return service.{{makePublic .Name | replaceReservedWords}}GenericContext(context.Background(), request)
		}
		{{else if ne $responseType ""}}
		func (service *{{$privateType}}) {{makePublic .Name | replaceReservedWords}}GenericContext (ctx context.Context) (soap.Result[{{$responseType}}], error) {
			response := new({{$responseType}})
			err := service.client.CallContext(ctx, "{{if ne $soapAction ""}}{{$soapAction}}{{else}}''{{end}}", nil, response)
			if err != nil {
				if fault, ok := err.(*soap.SOAPFault); ok {
					return soap.Result[{{$responseType}}]{Fault: fault}, nil
				}
				return soap.Result[{{$responseType}}]{}, err
			}
			return soap.Result[{{$responseType}}]{Value: *response}, nil
		}

		func (service *{{$privateType}}) {{makePublic .Name | replaceReservedWords}}Generic () (soap.Result[{{$responseType}}], error) {
			return service.{{makePublic .Name | replaceReservedWords}}GenericContext(context.Background())
		}
		{{end}}
	{{end}}
{{end}}
`