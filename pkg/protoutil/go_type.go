package protoutil

import (
	"fmt"

	"google.golang.org/protobuf/compiler/protogen"
	"google.golang.org/protobuf/reflect/protoreflect"
)

type Q func(protogen.GoIdent) string

type Option func(q Q, f *protogen.Field) (ok bool, result string)

func FieldGoType(q Q, f *protogen.Field, options ...Option) string {
	goType := ""
	ok := true
	for _, option := range options {
		ok, goType = option(q, f)
	}
	if !ok {
		return goType
	}

	if f.Desc.IsMap() {
		keyType := FieldGoType(q, f.Message.Fields[0], options...)
		valueType := FieldGoType(q, f.Message.Fields[1], options...)
		return fmt.Sprintf("map[%s]%s", keyType, valueType)
	}

	switch f.Desc.Kind() {
	case protoreflect.BoolKind:
		goType = "bool"
	case protoreflect.StringKind:
		goType = "string"
	case protoreflect.Int32Kind, protoreflect.Sint32Kind, protoreflect.Sfixed32Kind:
		goType = "int32"
	case protoreflect.Uint32Kind, protoreflect.Fixed32Kind:
		goType = "uint32"
	case protoreflect.Int64Kind, protoreflect.Sint64Kind, protoreflect.Sfixed64Kind:
		goType = "int64"
	case protoreflect.Uint64Kind, protoreflect.Fixed64Kind:
		goType = "uint64"
	case protoreflect.FloatKind:
		goType = "float32"
	case protoreflect.DoubleKind:
		goType = "float64"
	case protoreflect.BytesKind:
		goType = "[]byte"
	case protoreflect.EnumKind:
		goType = q(f.Enum.GoIdent)
	case protoreflect.MessageKind, protoreflect.GroupKind:
		goType = "*" + q(f.Message.GoIdent)
	}

	if f.Desc.IsList() {
		return "[]" + goType
	}

	return goType
}
