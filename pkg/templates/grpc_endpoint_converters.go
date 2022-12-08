package templates

var GRPCEnpointConverters = `package {{ .Package }}

import (
	"context"
	"fmt"
	"errors"
)

{{ range .Exchanges }}
func Encode{{ .Name }}(_ context.Context, r interface{}) (interface{}, error) {
	switch v := r.(type) {
	case *endpoints.{{ .Name }}:
		return {{ .Name }}2Proto(v)
	case endpoints.{{ .Name }}:
		return {{ .Name }}2Proto(&v)
	case nil:
		return nil, errors.New("nil {{ .Name }}")
	default:
		return nil, fmt.Errorf("unexpected type %T", r)
	}
}

func Decode{{ .Name }}(_ context.Context, r interface{}) (interface{}, error) {
	switch v := r.(type) {
	case *proto.{{ .Name }}:
		return Proto2{{ .Name }}(v)
	case nil:
		return nil, errors.New("nil {{ .Name }}")
	default:
		return nil, fmt.Errorf("unexpected type %T", r)
	}
}
{{ end }}
`
