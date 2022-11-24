package templates

var (
	StructTemplate = `
{{ .StructName }} struct {
{{- range .Fields }}
	{{ .Name }} {{ .Type.String }}
{{- else -}}
{{- end -}}
}
`
	ProtocolTemplate = `package {{ .Package }}

type (
{{ range .Structs }}
{{ template "struct" . }}
{{ end }}
)

{{ range .Structs }}
{{ if .Fields }}
func New{{ .StructName }} ({{ struct_constructor_args .Fields }}) {{ .StructName }} {
	return {{ .StructName }}{ {{- struct_constructor_return .Fields -}} }
}
{{ end }}
{{ end }}

{{ range .Structs }}
{{ if .Fields }}
func (r {{ .StructName }}) Args(ctx context.Context) (context.Context, {{ struct_return_types .Fields }}) {
	return ctx, {{ struct_return_args .Fields }}
}
{{ end }}
{{ end }}
`
)
