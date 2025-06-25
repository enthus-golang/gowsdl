package gowsdl

var wsdl2ServerTmpl = `

var WSDLUndefinedError = errors.New("Server was unable to process request. --> Object reference not set to an instance of an object.")

type SOAPEnvelopeRequest struct {
	XMLName xml.Name ` + "`" + `xml:"http://schemas.xmlsoap.org/soap/envelope/ Envelope"` + "`" + `
	Body SOAPBodyRequest
}

type SOAPBodyRequest struct {
	XMLName xml.Name ` + "`" + `xml:"http://schemas.xmlsoap.org/soap/envelope/ Body"` + "`" + `
	{{range .}}
		{{range .Operations}}
			{{if .Input}}
				{{$requestType := findType .Input.Element | replaceReservedWords | makePublic}} ` + `
  				{{$requestType}} *{{$requestType}} ` + "`" + `xml:,omitempty` + "`" + `
			{{end}}
		{{end}}
	{{end}}
}

type SOAPEnvelopeResponse struct { ` + `
	XMLName    xml.Name` + "`" + `xml:"soap:Envelope"` + "`" + `
	PrefixSoap string  ` + "`" + `xml:"xmlns:soap,attr"` + "`" + `
	PrefixXsi  string  ` + "`" + `xml:"xmlns:xsi,attr"` + "`" + `
	PrefixXsd  string  ` + "`" + `xml:"xmlns:xsd,attr"` + "`" + `

	Body SOAPBodyResponse
}

func NewSOAPEnvelopResponse() *SOAPEnvelopeResponse {
	return &SOAPEnvelopeResponse{
		PrefixSoap: "http://schemas.xmlsoap.org/soap/envelope/",
		PrefixXsd:  "http://www.w3.org/2001/XMLSchema",
		PrefixXsi:  "http://www.w3.org/2001/XMLSchema-instance",
	}
}

type Fault struct { ` + `
	XMLName xml.Name ` + "`" + `xml:"SOAP-ENV:Fault"` + "`" + `
	Space   string   ` + "`" + `xml:"xmlns:SOAP-ENV,omitempty,attr"` + "`" + `

	Code   string    ` + "`" + `xml:"faultcode,omitempty"` + "`" + `
	String string    ` + "`" + `xml:"faultstring,omitempty"` + "`" + `
	Actor  string 	 ` + "`" + `xml:"faultactor,omitempty"` + "`" + `
	Detail string    ` + "`" + `xml:"detail,omitempty"` + "`" + `
}


type SOAPBodyResponse struct { ` + `
	XMLName xml.Name   ` + "`" + `xml:"soap:Body"` + "`" + `
	Fault   *Fault ` + "`" + `xml:",omitempty"` + "`" + `
{{range .}}
	{{range .Operations}}
		{{$requestType := ""}}
		{{$responseType := ""}}
		{{if .Input}}{{$requestType = findType .Input.Element | replaceReservedWords | makePublic}}{{end}}
		{{if .Output}}{{$responseType = findType .Output.Element | replaceReservedWords | makePublic}}{{end}}
		{{if and .Input .Output}} ` + `
			{{$requestType}} *{{$responseType}} ` + "`" + `xml:",omitempty"` + "`" + `
		{{end}}
	{{end}}
{{end}}

}

{{range .}}
	{{range .Operations}}
		{{$requestType := ""}}
		{{$responseType := ""}}
		{{if .Input}}{{$requestType = findType .Input.Element | replaceReservedWords | makePublic}}{{end}}
		{{if .Output}}{{$responseType = findType .Output.Element | replaceReservedWords | makePublic}}{{end}}
		{{if and .Input .Output}}
			{{$requestTypeSource := findType .Input.Element | replaceReservedWords }}
func (service *SOAPBodyRequest) {{$requestType}}Func(request *{{$requestType}}) (*{{$responseType}}, error) {
	return nil, WSDLUndefinedError
}
		{{end}}
	{{end}}
{{end}}


func (service *SOAPEnvelopeRequest) call(w http.ResponseWriter, r *http.Request) {
	w.Header().Add("Content-Type", "text/xml; charset=utf-8")
	val := reflect.ValueOf(&service.Body).Elem()
	n := val.NumField()
	var field reflect.Value
	var name string
	find := false

	if r.Method == http.MethodGet {
		w.Write([]byte(wsdl))
		return
	}

	resp := NewSOAPEnvelopResponse()
	defer func() {
		if r := recover(); r != nil {
			resp.Body.Fault = &Fault{}
			resp.Body.Fault.Space = "http://schemas.xmlsoap.org/soap/envelope/"
			resp.Body.Fault.Code = "soap:Server"
			resp.Body.Fault.Detail = fmt.Sprintf("%v", r)
			resp.Body.Fault.String = fmt.Sprintf("%v", r)
		}
		xml.NewEncoder(w).Encode(resp)
	}()

	header := r.Header.Get("Content-Type")
	if !strings.Contains(header, "text/xml") && !strings.Contains(header, "application/soap+xml") {
		resp.Body.Fault = &Fault{}
		resp.Body.Fault.Space = "http://schemas.xmlsoap.org/soap/envelope/"
		resp.Body.Fault.Code = "soap:Client"
		resp.Body.Fault.String = "Invalid content type, expected text/xml or application/soap+xml"
		return
	}

	data, err := ioutil.ReadAll(r.Body)
	if err != nil {
		resp.Body.Fault = &Fault{}
		resp.Body.Fault.Space = "http://schemas.xmlsoap.org/soap/envelope/"
		resp.Body.Fault.Code = "soap:Client"
		resp.Body.Fault.String = err.Error()
		return
	}

	xml.Unmarshal(data, service)

	for i := 0; i < n; i++ {
		field = val.Field(i)
		name = val.Type().Field(i).Name

		if !field.IsNil() {
			find = true
			break
		}
	}

	if find {
		ret := reflect.ValueOf(&resp.Body).MethodByName(name + "Func").Call([]reflect.Value{field})

		if !ret[1].IsNil() {
			resp.Body.Fault = &Fault{}
			resp.Body.Fault.Space = "http://schemas.xmlsoap.org/soap/envelope/"
			resp.Body.Fault.Code = "soap:Server"
			resp.Body.Fault.String = ret[1].Interface().(error).Error()
			return
		}

		if ret[0].Type().Name() != "" {
			val = reflect.ValueOf(&resp.Body).Elem()
			f := val.FieldByName(name)
			f.Set(ret[0])
		}
	}
}
`