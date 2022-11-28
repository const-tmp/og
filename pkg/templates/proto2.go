package templates

var Proto2 = `syntax = "proto3";

option go_package = "{{ .GoPackagePath }};{{ .GoPackage }}";

package {{ .Package }};
{{ range .Services }}
// The {{ .Name }} service definition.
service {{ .Name }} {
    {{- range .Fields }}
    rpc {{ .Name }} ({{ .Request.Name }}) returns ({{ .Response.Name }});
    {{- end }}
}
{{- end }}
{{ range .Messages }}
{{- if .Fields }}
message {{ .Name }} {
{{- range .Fields }}
	{{ . }}
{{- end }}
}
{{ else }}
message {{ .Name }} {}
{{ end }}
{{- end }}
`
