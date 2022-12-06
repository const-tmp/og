package types

import (
	"fmt"
)

type (
	ProtoFile struct {
		GoPackage     string
		GoPackagePath string
		Package       string
		Services      []ProtoService
		Messages      []ProtoMessage
		Imports       []ProtoImport
	}

	ProtoMessage struct {
		Name   string
		Fields []ProtoField
	}

	ProtoRPC struct {
		Name     string
		Request  ProtoMessage
		Response ProtoMessage
	}

	ProtoService struct {
		Name   string
		Fields []ProtoRPC
	}

	ProtoField struct {
		Type   string
		Name   string
		Number uint
		OneOf  bool
	}

	ProtoImport struct {
		Path string
	}
)

func (p ProtoField) String() string {
	if !p.OneOf {
		return fmt.Sprintf("%s %s = %d;", p.Type, p.Name, p.Number)
	}
	return fmt.Sprintf(`oneof %s {
    // TODO
  }`, p.Name)
}
