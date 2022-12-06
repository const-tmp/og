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
	ProtoConverter struct {
		Type    types.Struct
		Proto   types.Struct
		Require []ProtoConverter
	}

	Encoder struct {
		StructName    string
		Converters    map[types.Field]Converter
		SubConverters []SubConverterCall
		Type          types.Struct
		Proto         types.Struct
		IsSlice       bool
		IsPointer     bool
		Deps          map[DepIndex]Dependency
		Imports       utils.Set[types.Import]
	}

	Converter struct {
		Funcs []TypeConverter
		Field types.Field
	}

	SubConverterCall struct {
		FieldName     string
		ConverterName string
		Converter     Converter
	}

	Dependency struct {
		Type      types.Struct
		Proto     types.Struct
		IsSlice   bool
		IsPointer bool
	}

	DepIndex struct {
		Type    string
		Proto   string
		IsSlice bool
	}
)

func (c Converter) Convert() string {
	s := c.Field.Name
	for _, converter := range c.Funcs {
		s = converter(s)
	}
	return s
}

func Structs2ProtoConverter(ctx *extract.Context, ty, pb *types.Struct) Encoder {
	ret := Encoder{
		StructName: ty.Name,
		Type:       *ty,
		Proto:      *pb,
		Converters: make(map[types.Field]Converter),
		Deps:       make(map[DepIndex]Dependency),
		Imports:    utils.NewSet[types.Import](),
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
	// we have to fill in all struct fields
	for _, pbField := range pb.Fields {
		// find matching field in second struct
		tyIdx := indexField(ty.Fields, pbField)
		if tyIdx == -1 {
			fmt.Printf("proto field %s.%s not found in %s\n", pb.Name, pbField.Name, ty.Name)
			continue
		}
		tyField := ty.Fields[tyIdx]

		switch {
		// slice encoder
		case isTypeSlice(tyField.Type) && isTypeSlice(pbField.Type):
			isPointer := isSliceTypePointer(tyField.Type)

			ret.Converters[pbField] = Converter{
				Funcs: []TypeConverter{names.GetUnexportedName},
				Field: tyField,
			}
			var prefix string
			if isPointer {
				prefix += "Pointer"
			}
			prefix += "Slice"

			ret.SubConverters = append(ret.SubConverters, SubConverterCall{
				FieldName:     pbField.Name,
				ConverterName: tyField.Type.Name() + prefix + "2Proto",
				Converter: Converter{
					Funcs: []TypeConverter{SelectorFactory("v"), addressFactory(tyField.Type, pbField.Type)},
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
			ret.Converters[pbField] = Converter{
				Funcs: []TypeConverter{SelectorFactory("v"), Error2String},
				Field: tyField,
			}
		case tyField.Type.Name() == "int" && pbField.Type.Name() == "int32":
			ret.Converters[pbField] = Converter{
				Funcs: []TypeConverter{SelectorFactory("v"), Int2Int32},
				Field: tyField,
			}
		case tyField.Type.String() == "time.Time" && pbField.Type.String() == "*timestamppb.Timestamp":
			ret.Converters[pbField] = Converter{
				Funcs: []TypeConverter{SelectorFactory("v"), Time2Proto},
				Field: tyField,
			}
		case tyField.Type.Name() == pbField.Type.Name():
			if tyField.Type.IsBuiltin() {
				ret.Converters[pbField] = Converter{
					Funcs: []TypeConverter{SelectorFactory("v"), addressFactory(tyField.Type, pbField.Type)},
					Field: tyField,
				}
			} else {
				ret.Converters[pbField] = Converter{
					Funcs: []TypeConverter{names.GetUnexportedName},
					Field: tyField,
				}
				ret.SubConverters = append(ret.SubConverters, SubConverterCall{
					FieldName:     pbField.Name,
					ConverterName: tyField.Type.Name() + "2Proto",
					Converter: Converter{
						Funcs: []TypeConverter{SelectorFactory("v"), addressFactory(tyField.Type, pbField.Type)},
						Field: tyField,
					},
				})
			}
		default:
			ret.Converters[pbField] = Converter{
				Funcs: []TypeConverter{names.GetUnexportedName},
				Field: tyField,
			}
			ret.SubConverters = append(ret.SubConverters, SubConverterCall{
				FieldName:     pbField.Name,
				ConverterName: tyField.Type.Name() + "2" + pbField.Type.Name(),
				Converter: Converter{
					Funcs: []TypeConverter{SelectorFactory("v"), addressFactory(tyField.Type, pbField.Type)},
					Field: tyField,
				},
			})
		}
	}

	return ret
}

func addressFactory(t, pb types.Type) TypeConverter {
	var prefix string
	if !isTypePointer(t) && isTypePointer(pb) {
		prefix = "&"
	}
	return func(s string) string {
		return prefix + names.GetUnexportedName(s)
	}
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
		return names.MatchProto(t.Name, pb.Name)
	})
)

func indexStruct(a []*types.Struct, s *types.Struct) int {
	return structSliceUtil.Index(a, s)
}

func indexField(a []types.Field, s types.Field) int {
	return fieldSliceUtil.Index(a, s)
}

type TypeConverter func(s string) string

func NoOpConverter(s string) string {
	return s
}

func SelectorFactory(sel string) TypeConverter {
	return func(s string) string {
		return fmt.Sprintf("%s.%s", sel, s)
	}
}

func Error2String(s string) string {
	return fmt.Sprintf("%s.Error()", s)
}

func Int2Int32(s string) string {
	return fmt.Sprintf("int32(%s)", s)
}

func Time2Proto(s string) string {
	return fmt.Sprintf("timestamppb.New(%s)", s)
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

func NewEncoderFactory(from, to string) TypeConverter {
	return func(s string) string {
		return fmt.Sprintf("%s2%s(%s)", from, to, s)
	}
}

func ToValue(s string) string {
	return names.GetUnexportedName(s)
}

func NewDecoder(s string) string {
	return fmt.Sprintf("Proto2%s", s)
}
