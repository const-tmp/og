package transform

import (
	"fmt"
	"github.com/nullc4t/og/internal/types"
	"github.com/nullc4t/og/pkg/names"
)

func Args2ProtoFields(args types.Args) []types.ProtoField {
	var res []types.ProtoField
	var i int

	for _, arg := range args {
		if arg.Type.Name() == "Context" {
			continue
		}
		var name string
		if arg.Name != "" {
			name = arg.Name
		} else {
			name = RenameEmpty(arg.Type)
		}
		fmt.Println(arg)
		res = append(res, types.ProtoField{
			Type:   Go2ProtobufType(arg.Type.String()),
			Name:   names.Camel2Snake(name),
			Number: uint(i + 1),
			OneOf:  arg.Type.IsInterface(),
		})
		i++
	}

	return res
}

func Fields2ProtoFields(args []types.Field) []types.ProtoField {
	var res []types.ProtoField
	var i int

	for _, arg := range args {
		if arg.Type.Name() == "Context" {
			continue
		}
		var name string
		if arg.Name != "" {
			name = arg.Name
		} else {
			name = RenameEmpty(arg.Type)
		}
		res = append(res, types.ProtoField{
			Type:   Go2ProtobufType(arg.Type.String()),
			Name:   names.Camel2Snake(name),
			Number: uint(i + 1),
		})
		i++
	}

	return res
}

func Interface2ProtoService(iface types.Interface) types.ProtoService {
	var proto = types.ProtoService{Name: iface.Name}

	for _, method := range iface.Methods {
		fmt.Println("converting ", method.Name)
		proto.Fields = append(proto.Fields, types.ProtoRPC{
			Name: method.Name,
			Request: types.ProtoMessage{
				Name:   fmt.Sprintf("%sRequest", method.Name),
				Fields: Args2ProtoFields(method.Args),
			},
			Response: types.ProtoMessage{
				Name:   fmt.Sprintf("%sResponse", method.Name),
				Fields: Args2ProtoFields(method.Results.Args),
			},
		})
	}

	return proto
}

func Struct2ProtoMessage(s types.Struct) types.ProtoMessage {
	var fields []types.ProtoField
	for i, field := range s.Fields {
		fields = append(fields, types.ProtoField{
			Type:   Go2ProtobufType(field.Type.String()),
			Name:   names.Camel2Snake(field.Name),
			Number: uint(i + 1),
		})
	}
	return types.ProtoMessage{
		Name:   s.Name,
		Fields: fields,
	}
}
