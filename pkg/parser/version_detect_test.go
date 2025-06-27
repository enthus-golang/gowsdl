package parser

import (
	"testing"
)

func TestDetectWSDLVersion(t *testing.T) {
	tests := []struct {
		name    string
		wsdl    string
		want    string
		wantErr bool
	}{
		{
			name: "WSDL 1.1 with namespace",
			wsdl: `<?xml version="1.0" encoding="UTF-8"?>
<definitions xmlns="http://schemas.xmlsoap.org/wsdl/" 
             targetNamespace="http://example.com/test">
</definitions>`,
			want:    "1.1",
			wantErr: false,
		},
		{
			name: "WSDL 1.1 with prefix",
			wsdl: `<?xml version="1.0" encoding="UTF-8"?>
<wsdl:definitions xmlns:wsdl="http://schemas.xmlsoap.org/wsdl/" 
                  targetNamespace="http://example.com/test">
</wsdl:definitions>`,
			want:    "1.1",
			wantErr: false,
		},
		{
			name: "WSDL 2.0 with namespace",
			wsdl: `<?xml version="1.0" encoding="UTF-8"?>
<description xmlns="http://www.w3.org/ns/wsdl/" 
             targetNamespace="http://example.com/test">
</description>`,
			want:    "2.0",
			wantErr: false,
		},
		{
			name: "WSDL 2.0 with prefix",
			wsdl: `<?xml version="1.0" encoding="UTF-8"?>
<wsdl2:description xmlns:wsdl2="http://www.w3.org/ns/wsdl/" 
                   targetNamespace="http://example.com/test">
</wsdl2:description>`,
			want:    "2.0",
			wantErr: false,
		},
		{
			name: "Invalid WSDL",
			wsdl: `<?xml version="1.0" encoding="UTF-8"?>
<root xmlns="http://example.com/unknown">
</root>`,
			want:    "",
			wantErr: true,
		},
		{
			name:    "Empty input",
			wsdl:    "",
			want:    "",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := DetectWSDLVersion([]byte(tt.wsdl))
			if (err != nil) != tt.wantErr {
				t.Errorf("detectWSDLVersion() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("detectWSDLVersion() = %v, want %v", got, tt.want)
			}
		})
	}
}