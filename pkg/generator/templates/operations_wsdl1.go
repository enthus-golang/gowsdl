// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at http://mozilla.org/MPL/2.0/.

package templates

// OperationsWSDL1Template is the template for WSDL 1.1 operations
const OperationsWSDL1Template = `
{{range .}}
	{{$portType := .Name | makePublic}}

	type {{$portType}} interface {
		{{range .Operations}}
			{{$requestType := findType .Input.Message | replaceReservedWords | makePublic}}
			{{$responseType := findType .Output.Message | replaceReservedWords | makePublic}}

			{{if .Faults}}
			// Error can be either of the following types:
			{{range .Faults}}
			//   - {{.Name}}{{if .Doc}} {{.Doc}}{{end}}{{end}}
			{{end}}
			{{if .Doc}}{{comment .Doc}}{{end}}
			{{.Name | makePublic}}(ctx context.Context, request *{{$requestType}}) (*{{$responseType}}, error)
		{{end}}
	}

	type {{$portType | makePrivate}} struct {
		client *soap.Client
	}

	func New{{$portType}}(client *soap.Client) {{$portType}} {
		return &{{$portType | makePrivate}}{
			client: client,
		}
	}

	{{range .Operations}}
		{{$requestType := findType .Input.Message | replaceReservedWords | makePublic}}
		{{$responseType := findType .Output.Message | replaceReservedWords | makePublic}}
		{{$soapAction := findSOAPAction .Name $portType}}

		func (service *{{$portType | makePrivate}}) {{.Name | makePublic}}(ctx context.Context, request *{{$requestType}}) (*{{$responseType}}, error) {
			response := new({{$responseType}})
			err := service.client.CallContext(ctx, "{{$soapAction}}", request, response)
			if err != nil {
				return nil, err
			}
			return response, nil
		}
	{{end}}
{{end}}
`