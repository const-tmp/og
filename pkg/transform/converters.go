package transform

import (
	"fmt"
	"github.com/nullc4t/og/internal/types"
	"github.com/nullc4t/og/pkg/extract"
	"github.com/nullc4t/og/pkg/names"
	"github.com/nullc4t/og/pkg/utils"
	"strings"
)

type (
	Converter struct {
		Type                types.Struct
		Proto               types.Struct
		StructName          string
		Expressions         map[types.Field]FieldExpression
		ConverterCalls      []ConverterCall
		IsSlice             bool
		IsPointer           bool
		IsInterface         bool
		Deps                map[DepIndex]Dependency
		Imports             utils.Set[types.Import]
		InterfaceConverters map[struct{ t, p string }]InterfaceConverter
		ErrorHandler        bool
	}

	InterfaceConverter struct {
		Name  string
		Type  types.Interface
		Proto types.Struct
	}

	FieldExpressionFunc func(s string) string

	FieldExpression struct {
		FieldExpressions []FieldExpressionFunc
		Field            types.Field
	}

	ConverterCall struct {
		FieldName     string
		ConverterName string
		Converter     FieldExpression
	}

	Dependency struct {
		Type        types.Struct
		Proto       types.Struct
		IsSlice     bool
		IsPointer   bool
		IsInterface bool
	}

	DepIndex struct {
		Type    string
		Proto   string
		IsSlice bool
	}
)

func NewConverter(t types.Struct, pb types.Struct, structName string) *Converter {
	return &Converter{
		Type:                t,
		Proto:               pb,
		StructName:          structName,
		Expressions:         make(map[types.Field]FieldExpression),
		Deps:                make(map[DepIndex]Dependency),
		Imports:             utils.NewSet[types.Import](),
		InterfaceConverters: map[struct{ t, p string }]InterfaceConverter{},
	}
}

func (c FieldExpression) Render() string {
	s := c.Field.Name
	for _, expression := range c.FieldExpressions {
		s = expression(s)
	}
	return s
}

func Structs2ProtoEncoder(ctx *extract.Context, ty, pb *types.Struct) Converter {
	ret := Converter{
		StructName:          ty.Name,
		Type:                *ty,
		Proto:               *pb,
		Expressions:         make(map[types.Field]FieldExpression),
		Deps:                make(map[DepIndex]Dependency),
		Imports:             utils.NewSet[types.Import](),
		InterfaceConverters: map[struct{ t, p string }]InterfaceConverter{},
	}
	// get all imported types to add it go generated file
	for _, field := range ty.Fields {
		if field.Type.IsImported() {
			ret.Imports.Add(
				types.Import{Path: field.Type.ImportPath(), Name: field.Type.Package()},
			)
		}
	}
	for _, pbField := range pb.Fields {
		if pbField.Type.IsImported() {
			ret.Imports.Add(
				types.Import{Path: pbField.Type.ImportPath(), Name: pbField.Type.Package()},
			)
		}
	}
	// fill in all fields of returning struct
	for _, pbField := range pb.Fields {
		// find matching field in second struct
		tyIdx := indexField(ty.Fields, pbField)
		if tyIdx == -1 {
			fmt.Printf("proto field %s.%s not found in %s\n", pb.Name, pbField.Name, ty.Name)
			ret.Expressions[pbField] = FieldExpression{
				FieldExpressions: []FieldExpressionFunc{TODOField},
			}
			continue
		}
		tyField := ty.Fields[tyIdx]

		tyi := ctx.GetInterface(tyField.Type)

		if tyi != nil {
			pbs := ctx.GetStruct(pbField.Type)
			ret.Expressions[pbField] = FieldExpression{
				FieldExpressions: []FieldExpressionFunc{
					names.Unexported,
					//AddressFactory(tyField.Type, pbField.Type),
				},
				Field: tyField,
			}
			ret.ConverterCalls = append(ret.ConverterCalls, ConverterCall{
				FieldName:     pbField.Name,
				ConverterName: tyField.Type.Name() + "2Proto",
				Converter: FieldExpression{
					FieldExpressions: []FieldExpressionFunc{SelectorFactory("v")},
					Field:            tyField,
				},
			})
			ret.InterfaceConverters[struct{ t, p string }{t: tyi.Name, p: pbs.Name}] = InterfaceConverter{
				Name:  tyi.Name + "2Proto",
				Type:  *tyi,
				Proto: *pbs,
			}
			continue
		}

		switch {
		// slice encoder
		case isTypeSlice(tyField.Type) && isTypeSlice(pbField.Type):
			isPointer := isSliceTypePointer(tyField.Type)

			ret.Expressions[pbField] = FieldExpression{
				FieldExpressions: []FieldExpressionFunc{names.Unexported},
				Field:            tyField,
			}
			var prefix string
			if isPointer {
				prefix += "Pointer"
			}
			prefix += "Slice"

			ret.ConverterCalls = append(ret.ConverterCalls, ConverterCall{
				FieldName:     pbField.Name,
				ConverterName: tyField.Type.Name() + prefix + "2Proto",
				Converter: FieldExpression{
					FieldExpressions: []FieldExpressionFunc{
						SelectorFactory("v"),
						AddressFactory(tyField.Type, pbField.Type),
					},
					Field: tyField,
				},
			})

			ret.Deps[DepIndex{Type: tyField.Type.Name(), Proto: pbField.Type.Name(), IsSlice: true}] = Dependency{
				Type:      *ctx.GetStruct(tyField.Type),
				Proto:     *ctx.GetStruct(pbField.Type),
				IsSlice:   true,
				IsPointer: isPointer,
			}
			ret.Deps[DepIndex{Type: tyField.Type.Name(), Proto: pbField.Type.Name(), IsSlice: false}] = Dependency{
				Type:    *ctx.GetStruct(tyField.Type),
				Proto:   *ctx.GetStruct(pbField.Type),
				IsSlice: false,
			}

		case tyField.Type.Name() == "error" && pbField.Type.Name() == "string":
			ret.ErrorHandler = true
			ret.Expressions[pbField] = FieldExpression{
				FieldExpressions: []FieldExpressionFunc{ErrorString},
			}
		case tyField.Type.Name() == "int" && pbField.Type.Name() == "int32":
			ret.Expressions[pbField] = FieldExpression{
				FieldExpressions: []FieldExpressionFunc{SelectorFactory("v"), Int2Int32},
				Field:            tyField,
			}
		case tyField.Type.String() == "time.Time" && pbField.Type.String() == "*timestamppb.Timestamp":
			ret.Expressions[pbField] = FieldExpression{
				FieldExpressions: []FieldExpressionFunc{SelectorFactory("v"), Time2Proto},
				Field:            tyField,
			}
		case tyField.Type.String() == "decimal.Decimal" && pbField.Type.String() == "string":
			ret.Expressions[pbField] = FieldExpression{
				FieldExpressions: []FieldExpressionFunc{SelectorFactory("v"), Decimal2String},
				Field:            tyField,
			}
		case tyField.Type.Name() == pbField.Type.Name():
			if tyField.Type.IsBuiltin() {
				ret.Expressions[pbField] = FieldExpression{
					FieldExpressions: []FieldExpressionFunc{SelectorFactory("v"), AddressFactory(tyField.Type, pbField.Type)},
					Field:            tyField,
				}
			} else {
				ret.Expressions[pbField] = FieldExpression{
					FieldExpressions: []FieldExpressionFunc{names.Unexported},
					Field:            tyField,
				}
				ret.ConverterCalls = append(ret.ConverterCalls, ConverterCall{
					FieldName:     pbField.Name,
					ConverterName: tyField.Type.Name() + "2Proto",
					Converter: FieldExpression{
						FieldExpressions: []FieldExpressionFunc{SelectorFactory("v"), AddressFactory(tyField.Type, pbField.Type)},
						Field:            tyField,
					},
				})
			}
		default:
			ret.Expressions[pbField] = FieldExpression{
				FieldExpressions: []FieldExpressionFunc{names.Unexported},
				Field:            tyField,
			}
			ret.ConverterCalls = append(ret.ConverterCalls, ConverterCall{
				FieldName:     pbField.Name,
				ConverterName: tyField.Type.Name() + "2" + pbField.Type.Name(),
				Converter: FieldExpression{
					FieldExpressions: []FieldExpressionFunc{SelectorFactory("v"), AddressFactory(tyField.Type, pbField.Type)},
					Field:            tyField,
				},
			})
		}
	}

	return ret
}

func Structs2ProtoDecoder(ctx *extract.Context, ty, pb *types.Struct) Converter {
	ret := NewConverter(*ty, *pb, ty.Name)

	// get all imported types to add it go generated file
	for _, field := range ty.Fields {
		if field.Type.IsImported() {
			ret.Imports.Add(
				types.Import{Path: field.Type.ImportPath(), Name: field.Type.Package()},
			)
		}
	}
	for _, pbField := range pb.Fields {
		if pbField.Type.IsImported() {
			ret.Imports.Add(
				types.Import{Path: pbField.Type.ImportPath(), Name: pbField.Type.Package()},
			)
		}
	}
	// fill in all fields of returning struct
	for _, tyField := range ty.Fields {
		// find matching field in second struct
		pbIdx := indexField(pb.Fields, tyField)
		if pbIdx == -1 {
			fmt.Printf("proto field %s.%s not found in %s\n", pb.Name, tyField.Name, ty.Name)
			if tyField.Name == "" {
				ret.Expressions[tyField] = FieldExpression{
					FieldExpressions: []FieldExpressionFunc{EmbeddedStructFactory(tyField.Type)},
				}
			} else {
				ret.Expressions[tyField] = FieldExpression{
					FieldExpressions: []FieldExpressionFunc{TODOField},
				}
			}
			continue
		}
		pbField := pb.Fields[pbIdx]

		fmt.Printf("type: %s.%s (%s)\tproto: %s.%s (%s)\n", ty.Name, tyField.Name, tyField.Type, pb.Name, pbField.Name, pbField.Type)

		tyi := ctx.GetInterface(tyField.Type)
		if tyi != nil {
			pbs := ctx.GetStruct(pbField.Type)
			ret.Expressions[tyField] = FieldExpression{
				FieldExpressions: []FieldExpressionFunc{
					names.Unexported,
					//DereferenceFactory(tyField.Type, pbField.Type),
				},
				Field: pbField,
			}
			ret.ConverterCalls = append(ret.ConverterCalls, ConverterCall{
				FieldName:     tyField.Name,
				ConverterName: "Proto2" + pbField.Type.Name(),
				Converter: FieldExpression{
					FieldExpressions: []FieldExpressionFunc{SelectorFactory("v")},
					Field:            pbField,
				},
			})
			ret.InterfaceConverters[struct{ t, p string }{t: tyi.Name, p: pbs.Name}] = InterfaceConverter{
				Name:  tyi.Name,
				Type:  *tyi,
				Proto: *pbs,
			}
			continue
		}

		switch {
		// slice encoder
		case isTypeSlice(pbField.Type) && isTypeSlice(tyField.Type):
			isPointer := isSliceTypePointer(tyField.Type)

			ret.Expressions[tyField] = FieldExpression{
				FieldExpressions: []FieldExpressionFunc{names.Unexported},
				Field:            pbField,
			}

			var suffix string
			if isPointer {
				suffix += "Pointer"
			}
			suffix += "Slice"

			ret.ConverterCalls = append(ret.ConverterCalls, ConverterCall{
				FieldName:     tyField.Name,
				ConverterName: "Proto2" + pbField.Type.Name() + suffix,
				Converter: FieldExpression{
					FieldExpressions: []FieldExpressionFunc{
						SelectorFactory("v"),
						DereferenceFactory(tyField.Type, pbField.Type),
					},
					Field: pbField,
				},
			})

			ret.Deps[DepIndex{Type: tyField.Type.Name(), Proto: pbField.Type.Name(), IsSlice: true}] = Dependency{
				Type:      *ctx.GetStruct(tyField.Type),
				Proto:     *ctx.GetStruct(pbField.Type),
				IsSlice:   true,
				IsPointer: isPointer,
			}
			ret.Deps[DepIndex{Type: tyField.Type.Name(), Proto: pbField.Type.Name(), IsSlice: false}] = Dependency{
				Type:    *ctx.GetStruct(tyField.Type),
				Proto:   *ctx.GetStruct(pbField.Type),
				IsSlice: false,
			}

		case pbField.Type.Name() == "error" && tyField.Type.Name() == "string":
			ret.Expressions[tyField] = FieldExpression{
				FieldExpressions: []FieldExpressionFunc{SelectorFactory("v"), Error2String},
				Field:            pbField,
			}
		case pbField.Type.Name() == "string" && tyField.Type.Name() == "error":
			ret.ErrorHandler = true
			ret.Expressions[tyField] = FieldExpression{
				FieldExpressions: []FieldExpressionFunc{ServiceError},
			}
			ret.Imports.Add(types.Import{Path: "errors"})
		case pbField.Type.Name() == "int" && tyField.Type.Name() == "int32":
			ret.Expressions[tyField] = FieldExpression{
				FieldExpressions: []FieldExpressionFunc{SelectorFactory("v"), Int2Int32},
				Field:            pbField,
			}
		case pbField.Type.Name() == "int32" && tyField.Type.Name() == "int":
			ret.Expressions[tyField] = FieldExpression{
				FieldExpressions: []FieldExpressionFunc{SelectorFactory("v"), Int322Int},
				Field:            pbField,
			}
		case pbField.Type.String() == "time.Time" && tyField.Type.String() == "*timestamppb.Timestamp":
			ret.Expressions[tyField] = FieldExpression{
				FieldExpressions: []FieldExpressionFunc{SelectorFactory("v"), Time2Proto},
				Field:            pbField,
			}
		case pbField.Type.String() == "*timestamppb.Timestamp" && tyField.Type.String() == "time.Time":
			ret.Expressions[tyField] = FieldExpression{
				FieldExpressions: []FieldExpressionFunc{SelectorFactory("v"), Proto2Time},
				Field:            pbField,
			}
		case pbField.Type.String() == "string" && tyField.Type.String() == "decimal.Decimal":
			ret.Expressions[tyField] = FieldExpression{
				FieldExpressions: []FieldExpressionFunc{names.Unexported},
				Field:            pbField,
			}
			ret.ConverterCalls = append(ret.ConverterCalls, ConverterCall{
				FieldName:     tyField.Name,
				ConverterName: "decimal.NewFromString",
				Converter: FieldExpression{
					FieldExpressions: []FieldExpressionFunc{SelectorFactory("v")},
					Field:            pbField,
				},
			})
			ret.Imports.Add(types.Import{Path: tyField.Type.ImportPath()})

		case pbField.Type.Name() == tyField.Type.Name():
			if pbField.Type.IsBuiltin() {
				ret.Expressions[tyField] = FieldExpression{
					FieldExpressions: []FieldExpressionFunc{SelectorFactory("v"), DereferenceFactory(tyField.Type, pbField.Type)},
					Field:            pbField,
				}
			} else {
				ret.Expressions[tyField] = FieldExpression{
					FieldExpressions: []FieldExpressionFunc{names.Unexported, DereferenceFactory(tyField.Type, pbField.Type)},
					Field:            pbField,
				}
				ret.ConverterCalls = append(ret.ConverterCalls, ConverterCall{
					FieldName:     tyField.Name,
					ConverterName: "Proto2" + tyField.Type.Name(),
					Converter: FieldExpression{
						FieldExpressions: []FieldExpressionFunc{SelectorFactory("v")},
						Field:            pbField,
					},
				})
			}
		default:
			ret.Expressions[tyField] = FieldExpression{
				FieldExpressions: []FieldExpressionFunc{names.Unexported},
				Field:            pbField,
			}
			ret.ConverterCalls = append(ret.ConverterCalls, ConverterCall{
				FieldName:     tyField.Name,
				ConverterName: pbField.Type.Name() + "2" + tyField.Type.Name(),
				Converter: FieldExpression{
					FieldExpressions: []FieldExpressionFunc{SelectorFactory("v"), DereferenceFactory(tyField.Type, pbField.Type)},
					Field:            pbField,
				},
			})
		}
	}

	return *ret
}

func AddressFactory(t, pb types.Type) FieldExpressionFunc {
	var prefix string
	if !isTypePointer(t) && isTypePointer(pb) {
		prefix = "&"
	}
	return func(s string) string {
		return prefix + names.Unexported(s)
	}
}

func DereferenceFactory(t, pb types.Type) FieldExpressionFunc {
	var prefix string
	if !isTypePointer(t) && isTypePointer(pb) {
		prefix = "*"
	}
	return func(s string) string {
		return prefix + s
	}
}

func EmbeddedStructFactory(t types.Type) FieldExpressionFunc {
	return func(_ string) string {
		return fmt.Sprintf(`%s.%s{
	// TODO
}`, t.Package(), t.Name())
	}
}

func TODOField(_ string) string {
	return "todo /* TODO */"
}

func isTypePointer(p types.Type) bool {
	return strings.HasPrefix(p.String(), "*")
}

func isSliceTypePointer(p types.Type) bool {
	return strings.HasPrefix(strings.Replace(p.String(), "[]", "", 1), "*")
}

func isTypeSlice(p types.Type) bool {
	return strings.HasPrefix(p.String(), "[]")
}

var (
	structSliceUtil = utils.NewSlice[*types.Struct](func(t, pb *types.Struct) bool {
		return t.Name == pb.Name
	})
	fieldSliceUtil = utils.NewSlice[types.Field](func(t, pb types.Field) bool {
		return names.MatchProto(t.Name, pb.Name) || names.MatchProto(pb.Name, t.Name)
	})
)

func indexStruct(a []*types.Struct, s *types.Struct) int {
	return structSliceUtil.Index(a, s)
}

func indexField(a []types.Field, s types.Field) int {
	return fieldSliceUtil.Index(a, s)
}

func NoOpConverter(s string) string {
	return s
}

func SelectorFactory(sel string) FieldExpressionFunc {
	return func(s string) string {
		return fmt.Sprintf("%s.%s", sel, s)
	}
}

func Error2String(s string) string {
	return fmt.Sprintf("%s.Error()", s)
}

func ErrorString(s string) string {
	return "errorString"
}

func ServiceError(s string) string {
	return "serviceError"
}

func String2Error(s string) string {
	return fmt.Sprintf("errors.New(%s)", s)
}

func Int2Int32(s string) string {
	return fmt.Sprintf("int32(%s)", s)
}

func Int322Int(s string) string {
	return fmt.Sprintf("int(%s)", s)
}

func Time2Proto(s string) string {
	return fmt.Sprintf("timestamppb.New(%s)", s)
}

func Proto2Time(s string) string {
	return fmt.Sprintf("%s.AsTime()", s)
}

func Decimal2String(s string) string {
	return fmt.Sprintf("%s.String()", s)
}

func ValueOf(s string) string {
	return fmt.Sprintf("*%s", s)
}

func AddressOf(s string) string {
	return fmt.Sprintf("&%s", s)
}

func NewEncoder(s string) string {
	return fmt.Sprintf("%s2Proto", s)
}

func NewEncoderFactory(from, to string) FieldExpressionFunc {
	return func(s string) string {
		return fmt.Sprintf("%s2%s(%s)", from, to, s)
	}
}

func ToValue(s string) string {
	return names.Unexported(s)
}

func NewDecoder(s string) string {
	return fmt.Sprintf("Proto2%s", s)
}
