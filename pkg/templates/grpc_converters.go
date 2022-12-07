package templates

var GRPCConverters = `
package {{ .Package }}

import (
    "context"

    "google.golang.org/grpc"

    "github.com/go-kit/kit/endpoint"
    "github.com/go-kit/kit/log"
    "github.com/go-kit/kit/transport"
    grpctransport "github.com/go-kit/kit/transport/grpc"
)

{{ range .Messages }}
// decodeGRPC{{ .Name }} is a transport/grpc.DecodeRequestFunc that converts a
// gRPC sum request to a user-domain sum request. Primarily useful in a server.
func DecodeGRPC{{ .Name }}(_ context.Context, grpcReq interface{}) (interface{}, error) {
    req := grpcReq.(*proto.{{ .Name }})
    return endpoints.{{ .Name }}{/* TODO */}, nil
}

// EncodeGRPC{{ .Name }} is a transport/grpc.EncodeResponseFunc that converts
// a user-domain concat response to a gRPC concat reply. Primarily useful in a
// server.
func encodeGRPC{{ .Name }}(_ context.Context, response interface{}) (interface{}, error) {
    resp := response.(endpoints.{{ .Name }})
    return &proto.{{ .Name }}{/* TODO */}, nil
}
{{ end }}
`

var GRPCEncoder = `package {{ .Package }}
{{ range $k, $conv := .InterfaceConverters }}
func {{ $conv.Name }}2Proto(v {{ $conv.Type.Package }}.{{ $conv.Type.Name }}) (*{{ $conv.Proto.Package }}.{{ $conv.Proto.Name }}, error) {
	panic("unimplemented") // TODO
}

func Proto2{{ $conv.Name }}(v *{{ $conv.Proto.Package }}.{{ $conv.Proto.Name }}) ({{ $conv.Type.Package }}.{{ $conv.Type.Name }}, error) {
	panic("unimplemented") // TODO
}
{{- end }}

{{ range .Encoders }}
{{- if .IsSlice }}
func {{ .StructName }}{{ if .IsPointer }}Pointer{{ end }}{{ if .IsSlice }}Slice{{ end }}2Proto(v []{{ if .IsPointer }}*{{ end }}{{ .Type.Package }}.{{ .Type.Name }}) ([]*{{ .Proto.Package }}.{{ .Proto.Name }}, error) {
	var res []*{{ .Proto.Package }}.{{ .Proto.Name }}
	for _, x := range v {
		p, err := {{ .StructName }}2Proto({{ if not .IsPointer }}&{{ end }}x)
		if err != nil {
			return nil, err
		}
		res = append(res, p)
	}
	return res, nil
}
{{ else }}
func {{ .StructName }}2Proto(v *{{ .Type.Package }}.{{ .Type.Name }}) (*{{ .Proto.Package }}.{{ .Proto.Name }}, error) {
{{- range .ConverterCalls }}
	{{ .FieldName | unexported}}, err := {{ .ConverterName }}({{ .Converter.Render }})
	if err != nil {
		return nil, err
	}
{{ end }}
	return &{{ .Proto.Package }}.{{ .Proto.Name }}{
{{- range $k, $v := .Expressions }}
		{{ $k.Name }}: {{ $v.Render }},
{{- end }}
	}, nil
}
{{ end }}
{{- end }}

{{ range .Decoders }}
{{- if .IsSlice }}
func Proto2{{ .StructName }}{{ if .IsPointer }}Pointer{{ end }}{{ if .IsSlice }}Slice{{ end }}(v []*{{ .Proto.Package }}.{{ .Proto.Name }}) ([]{{ if .IsPointer }}*{{ end }}{{ .Type.Package }}.{{ .Type.Name }}, error) {
	var res []{{ if .IsPointer }}*{{ end }}{{ .Type.Package }}.{{ .Type.Name }}
	for _, x := range v {
		p, err := Proto2{{ .StructName }}(x)
		if err != nil {
			return nil, err
		}
		res = append(res, {{ if not .IsPointer }}*{{ end }}p)
	}
	return res, nil
}
{{ else }}
func Proto2{{ .StructName }}(v *{{ .Proto.Package }}.{{ .Proto.Name }}) (*{{ .Type.Package }}.{{ .Type.Name }}, error) {
{{- range .ConverterCalls }}
	{{ .FieldName | unexported}}, err := {{ .ConverterName }}({{ .Converter.Render }})
	if err != nil {
		return nil, err
	}
{{ end }}
	return &{{ .Type.Package }}.{{ .Type.Name }}{
{{- range $k, $v := .Expressions }}
		{{ if $k.Name }}{{ $k.Name }}{{ else }}{{ $k.Type.Name }}{{ end }}: {{ $v.Render }},
{{- end }}
	}, nil
}
{{ end }}
{{- end }}


`
