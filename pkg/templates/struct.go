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
`
)
