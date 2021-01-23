package protoutil

import (
	"google.golang.org/protobuf/compiler/protogen"
	"google.golang.org/protobuf/reflect/protoreflect"
)

type shimmedListDesc struct {
	protoreflect.FieldDescriptor
}

func (d *shimmedListDesc) IsList() bool {
	return false
}

func ElemOfListField(f *protogen.Field) *protogen.Field {
	return &protogen.Field{
		Desc:     &shimmedListDesc{FieldDescriptor: f.Desc},
		GoName:   f.GoName,
		GoIdent:  f.GoIdent,
		Parent:   f.Parent,
		Oneof:    f.Oneof,
		Extendee: f.Extendee,
		Enum:     f.Enum,
		Message:  f.Message,
		Location: f.Location,
		Comments: f.Comments,
	}
}
