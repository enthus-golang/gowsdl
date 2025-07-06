// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at http://mozilla.org/MPL/2.0/.

// Example showing how to use gowsdl with unqualified schema elements
package example

import (
	"fmt"
	
	// Import your generated package
	// "your-module/generated"
	"github.com/enthus-golang/gowsdl/soap"
)

// DemonstrateNamespaceFix demonstrates the fix for namespace handling when elementFormDefault is unqualified
func DemonstrateNamespaceFix() {
	// Create SOAP client
	_ = soap.NewClient("https://example.com/soap/service")
	
	// Create service from generated code
	// service := generated.NewGeneralOvernightService(client)
	
	// Create request - the generated types now have correct namespace handling
	/*
	request := &generated.GOWebService_SendungsErstellung{
		Versender:      "2611570",
		Benutzername:   "21480MCL",
		Status:         "1",
		Kundenreferenz: "AU201061_LS300999",
		Ansprechpartner: &generated.GOWebService_SendungsErstellungAnsprechpartnerType{
			Telefon: generated.GOWebService_SendungsErstellungAnsprechpartnerTypeTelefonType{
				LaenderPrefix:  "49",
				Vorwahl:        "89", 
				Telefonnummer:  "12345678",
			},
		},
	}
	*/
	
	// The generated code now:
	// 1. Uses namespace prefix (tns:) for operation elements when elementFormDefault is unqualified
	// 2. Implements TargetNamespace() method that the SOAP client uses
	// 3. Generates proper SOAP envelope with namespace declarations
	
	// Call the service
	// response, err := service.SendungsErstellung(request)
	
	// The SOAP envelope will now be correctly formatted as:
	/*
	<soap:Envelope xmlns:soap="http://schemas.xmlsoap.org/soap/envelope/" 
	               xmlns:tns="https://wsdemo.ax4.com/ws/GeneralOvernight"
	               xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance"
	               xmlns:xsd="http://www.w3.org/2001/XMLSchema">
		<soap:Body>
			<tns:GOWebService_SendungsErstellung>
				<Versender>2611570</Versender>
				<Benutzername>21480MCL</Benutzername>
				<Status>1</Status>
				<Kundenreferenz>AU201061_LS300999</Kundenreferenz>
				<Ansprechpartner>
					<Telefon>
						<LaenderPrefix>49</LaenderPrefix>
						<Vorwahl>89</Vorwahl>
						<Telefonnummer>12345678</Telefonnummer>
					</Telefon>
				</Ansprechpartner>
			</tns:GOWebService_SendungsErstellung>
		</soap:Body>
	</soap:Envelope>
	*/
	
	// Note: The child elements (Versender, Benutzername, etc.) are unqualified (no namespace prefix)
	// This is correct behavior when elementFormDefault="unqualified" (the default)
	
	fmt.Println("Example showing namespace fix for unqualified schemas")
	
	// Additional notes:
	// - If your WSDL has elementFormDefault="qualified", the generated code will use 
	//   the full namespace format: xml:"namespace localname"
	// - If elementFormDefault is not specified or is "unqualified", the generated code
	//   will use namespace prefixes: xml:"tns:localname"
}