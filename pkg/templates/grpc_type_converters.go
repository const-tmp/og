package templates

var GRPCTypeConverters = `package {{ .Package }}

{{ range .Exchanges }}
func {{ .Name }}2Proto(v *endpoints.{{ .Name }}) (*proto.{{ .Name }}, error) {
	panic("unimplemented") // TODO
}

func Proto2{{ .Name }}(v *proto.{{ .Name }}) (*endpoints.{{ .Name }}, error) {
	panic("unimplemented") // TODO
}
{{ end }}
`
