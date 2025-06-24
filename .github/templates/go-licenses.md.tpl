## [go-licenses](https://github.com/google/go-licenses) report

{{- range . }}
- {{ .Name }} ({{ .Version }}) [{{ .LicenseName }}]({{ .LicenseURL }})
{{- end }}
