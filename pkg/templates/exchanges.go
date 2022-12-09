package templates

var (
	TransportExchanges = `// Code generated by og. DO NOT EDIT.
package {{ .Package }}

type (
{{- range .Structs }}
// {{ .StructName }} is an exchange struct
{{ .StructName }} struct {
{{- range .Fields }}
{{ .Name }} {{ .Type.String }} {{ .Name | camel2snake | jsonTag }}
{{- else -}}
{{- end -}}
}
{{ end }}
)

{{ range .Structs }}

// New{{ .StructName }} is a constructor for {{ .StructName }}
func New{{ .StructName }} ({{ struct_constructor_args .Fields }}) {{ .StructName }} {
	return {{ .StructName }}{ {{- struct_constructor_return .Fields -}} }
}

// Args is a shortcut returning args to original interface's method
{{- if .HasContext }}
func (r {{ .StructName }}) Args(ctx context.Context) (context.Context, {{ struct_types .Fields }}) {
{{ if .Fields }}
	return ctx, {{ struct_args .Fields }}
{{ else }}
	return ctx
{{ end }}
}
{{ else }}
func (r {{ .StructName }}) Args() ({{ struct_types .Fields }}) {
	return {{ struct_args .Fields }}
}
{{ end }}
{{ end }}
`
)
