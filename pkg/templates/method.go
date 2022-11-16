package templates

const (
	NamedArgs = "{{ define \"named_args\" }}{{ join (args .) \", \"}}{{ end }}"
	Method    = `{{ define "method" }}
func ({{ .receiver }}) {{ .method.Name }}({{ template "named_args" .method.Args }}) ({{ template "named_args" .method.Results }}) {
	{{- block "body" . }}
	// TODO
	panic("unimplemented")
	{{ end }}
}
{{ end }}`
)
