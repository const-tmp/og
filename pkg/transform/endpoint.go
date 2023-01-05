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
			if types.ArgIsError(*arg) {
				continue
			}
			// do not add context.Context argument
			if types.ArgIsContext(*arg) {
				requestStruct.HasContext = true
			} else {
				fmt.Println(arg.String())
				requestStruct.Fields = append(requestStruct.Fields, arg)
			}
		}
		es = append(es, requestStruct)

		// for response
		responseStruct := types.ExchangeStruct{StructName: fmt.Sprintf("%sResponse", method.Name)}
		for _, arg := range method.Results.Args {
			if types.ArgIsError(*arg) {
				continue
			}
			responseStruct.Fields = append(responseStruct.Fields, arg)
		}

		es = append(es, responseStruct)
	}

	return es
}

func RenameExchangeStruct(exchangeStruct types.ExchangeStruct) types.ExchangeStruct {
	for _, field := range exchangeStruct.Fields {
		RenameArg(field)
	}
	return exchangeStruct
}

func RenameArg(arg *types.Arg) {
	switch arg.Name {
	case "":
		switch arg.Type.Name() {
		case "error":
			arg.Name = "Error"
		case "Context":
			arg.Name = "Context"
		default:
			arg.Name = names.GetExportedName(arg.Type.Name())
		}
	case "err", "Err", "Error", "error":
		arg.Name = "Error"
	default:
		arg.Name = names.GetExportedName(arg.Name)
	}
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

func NameEmptyArgsInInterface(iface *types.Interface) {
	for _, method := range iface.Methods {
		for _, arg := range method.Args {
			if arg.Name == "" {
				arg.Name = renameEmpty(arg.Type)
			}
		}
		for _, arg := range method.Results.Args {
			if arg.Name == "" {
				arg.Name = renameEmpty(arg.Type)
			}
		}
	}
}

func renameEmpty(t types.Type) string {
	switch t.Name() {
	case "error":
		return "err"
	case "Context":
		return "ctx"
	default:
		if isTypeSlice(t) {
			return names.Unexported(t.Name() + "List")
		}
		return names.Unexported(t.Name())
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
