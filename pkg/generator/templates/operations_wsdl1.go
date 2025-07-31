// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at http://mozilla.org/MPL/2.0/.

package templates

// OperationsWSDL1Template is the template for WSDL 1.1 operations
const OperationsWSDL1Template = `
{{range .PortTypes}}
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
			{{if isRPCPortType $portType}}
				// RPC-style operation: wrap request in operation element
				{{$opName := .Name}}
				{{$inputMsgName := .Input.Message | removeNamespacePrefix}}
				// Create operation wrapper directly
				α := Operation{{$opName | makePublic}}In{
					{{/* Convert request fields to operation wrapper */}}
					{{range $.Messages}}
						{{if eq .Name $inputMsgName}}
							{{range .Parts}}
								{{.Name | replaceReservedWords | makePublic}}: &request.{{.Name | replaceReservedWords | makePublic}},
							{{end}}
						{{end}}
					{{end}}
				}
				
				var γ Operation{{.Name | makePublic}}Out
				
				err := service.client.CallContext(ctx, "{{$soapAction}}", α, &γ)
				if err != nil {
					return nil, err
				}
				
				// Convert response wrapper to output type
				{{$outputMsgName := .Output.Message | removeNamespacePrefix}}
				{{range $.Messages}}
					{{if eq .Name $outputMsgName}}
						{{if eq (len .Parts) 1}}
							{{range .Parts}}
								{{$partName := .Name | replaceReservedWords | makePublic}}
								{{if eq .Name "return"}}
									if γ.Return == nil {
										return nil, nil
									}
									result := &{{$responseType}}{
										Result: *γ.Return,
									}
									return result, nil
								{{else}}
									if γ.{{$partName}} == nil {
										return nil, nil
									}
									result := &{{$responseType}}{
										{{$partName}}: *γ.{{$partName}},
									}
									return result, nil
								{{end}}
							{{end}}
						{{else}}
							// Multiple parts - need to handle each field
							result := &{{$responseType}}{
								{{range .Parts}}
									{{$partName := .Name | replaceReservedWords | makePublic}}
									{{$partName}}: γ.{{$partName}},
								{{end}}
							}
							return result, nil
						{{end}}
					{{end}}
				{{end}}
			{{else}}
				// Document-style operation
				response := new({{$responseType}})
				err := service.client.CallContext(ctx, "{{$soapAction}}", request, response)
				if err != nil {
					return nil, err
				}
				return response, nil
			{{end}}
		}
	{{end}}
{{end}}
`