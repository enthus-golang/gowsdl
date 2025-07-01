// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at http://mozilla.org/MPL/2.0/.

package templates

// OperationsWSDL2Template is the template for WSDL 2.0 operations
const OperationsWSDL2Template = `
{{range .}}
	{{$interface := .Name | makePublic}}

	type {{$interface}} interface {
		{{range .Operations}}
			{{$requestType := ""}}
			{{$responseType := ""}}
			
			{{if .Input}}
				{{$requestType = findType .Input.Element | replaceReservedWords | makePublic}}
			{{end}}
			{{if .Output}}
				{{$responseType = findType .Output.Element | replaceReservedWords | makePublic}}
			{{end}}

			{{$hasFaults := false}}
			{{if .InFault}}{{$hasFaults = gt (len .InFault) 0}}{{end}}
			{{if .OutFault}}{{$hasFaults = gt (len .OutFault) 0}}{{end}}
			
			{{if $hasFaults}}
			// Error can be either of the following types:
			{{range .InFault}}
			//   - {{.Ref}} (input fault){{end}}
			{{range .OutFault}}
			//   - {{.Ref}} (output fault){{end}}
			{{end}}
			
			{{if .Doc}}{{comment .Doc}}{{end}}
			{{.Name | makePublic}}(ctx context.Context, request *{{$requestType}}) (*{{$responseType}}, error)
			{{.Name | makePublic}}Context(ctx context.Context, request *{{$requestType}}) (*{{$responseType}}, error)
		{{end}}
	}

	type {{$interface | makePrivate}} struct {
		client *soap.Client
	}

	func New{{$interface}}(client *soap.Client) {{$interface}} {
		return &{{$interface | makePrivate}}{
			client: client,
		}
	}

	{{range .Operations}}
		{{$requestType := ""}}
		{{$responseType := ""}}
		
		{{if .Input}}
			{{$requestType = findType .Input.Element | replaceReservedWords | makePublic}}
		{{end}}
		{{if .Output}}
			{{$responseType = findType .Output.Element | replaceReservedWords | makePublic}}
		{{end}}

		func (service *{{$interface | makePrivate}}) {{.Name | makePublic}}Context(ctx context.Context, request *{{$requestType}}) (*{{$responseType}}, error) {
			response := new({{$responseType}})
			err := service.client.CallContext(ctx, "", request, response)
			if err != nil {
				return nil, err
			}
			return response, nil
		}

		func (service *{{$interface | makePrivate}}) {{.Name | makePublic}}(ctx context.Context, request *{{$requestType}}) (*{{$responseType}}, error) {
			return service.{{.Name | makePublic}}Context(ctx, request)
		}
	{{end}}
{{end}}
`