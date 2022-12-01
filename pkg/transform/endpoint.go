package transform

import (
	"fmt"
	"github.com/nullc4t/og/internal/types"
	"github.com/nullc4t/og/pkg/names"
)

func Interface2ExchangeStructs(iface types.Interface) []types.ExchangeStruct {
	// create structs
	var es []types.ExchangeStruct

	for _, method := range iface.Methods {
		// for request
		requestStruct := types.ExchangeStruct{StructName: fmt.Sprintf("%sRequest", method.Name)}
		for _, arg := range method.Args {
			// do not add context.Context argument
			if ArgIsContext(*arg) {
				requestStruct.HasContext = true
			} else {
				requestStruct.Fields = append(requestStruct.Fields, arg)
			}
		}
		es = append(es, requestStruct)

		// for response
		responseStruct := types.ExchangeStruct{StructName: fmt.Sprintf("%sResponse", method.Name)}
		responseStruct.Fields = append(responseStruct.Fields, method.Results.Args...)

		es = append(es, responseStruct)
	}

	return es
}

func ArgIsContext(arg types.Arg) bool {
	return arg.Type.String() == "context.Context"
}

func ArgIsError(arg types.Arg) bool {
	return arg.Type.String() == "error"
}

func RenameExchangeStruct(exchangeStruct types.ExchangeStruct) types.ExchangeStruct {
	for _, field := range exchangeStruct.Fields {
		RenameArg(field)
	}
	return exchangeStruct
}

func RenameArg(arg *types.Arg) {
	if arg.Name == "" {
		switch arg.Type.Name() {
		case "error":
			arg.Name = "Error"
		case "arg":
			arg.Name = "Ctx"
		default:
			arg.Name = names.GetExportedName(arg.Type.Name())
		}
	}
	arg.Name = names.GetExportedName(arg.Name)
}

func RenameArgsInInterface(iface types.Interface) {
	for _, method := range iface.Methods {
		for _, arg := range method.Args {
			RenameArg(arg)
		}
		for _, arg := range method.Results.Args {
			RenameArg(arg)
		}
	}
}

func RenameEmpty(t types.Type) string {
	switch t.Name() {
	case "error":
		return "Error"
	case "ctx":
		return "Context"
	default:
		return names.GetExportedName(t.Name())
	}
}
