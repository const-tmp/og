package templates

var (
	StructTemplate = `
type {{ .StructName }} struct {
{{- range .Fields }}
	{{ .Name }} {{ .Type.String }}
{{- else -}}
{{- end -}}
}
`
	ProtocolTemplate = `package {{ .Package }}

{{ range .Structs }}
{{ template "struct" . }}
{{ end }}
`
)
