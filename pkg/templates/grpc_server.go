package templates

var GRPCServer = `package {{ .Package }}

import (
    "context"
    "github.com/go-kit/kit/transport/grpc"
)

// grpcServer is a proto.{{ .ServiceName }}Server
type grpcServer struct {
    proto.Unimplemented{{ .ServiceName }}Server
	{{- range .Endpoints }}
	{{ .Name | unexported }} grpc.Handler
	{{- end }}
}

// NewGRPCServer makes a set of endpoints available as a gRPC AddServer.
func NewGRPCServer(endpoints endpoints.EndpointSet) proto.{{ .ServiceName }}Server {
    // options := []grpc.ServerOption{
        //grpc.ServerErrorHandler(transport.NewLogErrorHandler(logger)),
    // }

    return &grpcServer{
		{{- range .Endpoints }}
        {{ .Name | unexported }}: grpc.NewServer(
        endpoints.{{ .Name }}Endpoint,
        Decode{{ .Name }}Request,
        Encode{{ .Name }}Response,
    ),
		{{- end }}
    }
}

{{ range .Endpoints }}
func (s *grpcServer) {{ .Name }}(ctx context.Context, req *proto.{{ .Name }}Request) (*proto.{{ .Name }}Response, error) {
    _, rep, err := s.{{ .Name | unexported }}.ServeGRPC(ctx, req)
    if err != nil {
        return nil, err
    }
    return rep.(*proto.{{ .Name }}Response), nil
}
{{ end }}
`
