package main

import (
	"flag"
	"fmt"

	"google.golang.org/protobuf/compiler/protogen"
	"google.golang.org/protobuf/reflect/protoreflect"

	"github.com/joesonw/proto-tools/pkg/genutil"
	"github.com/joesonw/proto-tools/pkg/protoutil"
)

func main() {
	var (
		flags        flag.FlagSet
		importPrefix = flags.String("import_prefix", "", "prefix to prepend to import paths")
	)
	importRewriteFunc := func(importPath protogen.GoImportPath) protogen.GoImportPath {
		switch importPath {
		case "context", "fmt", "math":
			return importPath
		}
		if *importPrefix != "" {
			return protogen.GoImportPath(*importPrefix) + importPath
		}
		return importPath
	}

	protogen.Options{
		ParamFunc:         flags.Set,
		ImportRewriteFunc: importRewriteFunc,
	}.Run(func(gen *protogen.Plugin) error {
		for _, file := range gen.Files {
			if !file.Generate {
				continue
			}
			filename := file.GeneratedFilenamePrefix + "_clone.pb.go"
			g := genutil.New(gen.NewGeneratedFile(filename, file.GoImportPath), func(g *genutil.G) genutil.Generator {
				return &G{
					G:    g,
					file: file,
					gen:  gen,
				}
			})
			if err := g.Generate(); err != nil {
				return err
			}
		}
		return nil
	})
}

type G struct {
	*genutil.G
	file *protogen.File
	gen  *protogen.Plugin
}

func (g *G) Generate() error {
	g.P("// Code generated by protoc-gen-clone. DO NOT EDIT.")
	g.P()
	g.P("package ", g.file.GoPackageName)
	g.P()

	for _, m := range g.file.Messages {
		g.genMessage(m)
	}

	return nil
}

func (g *G) genMessage(m *protogen.Message) {
	if m.Desc.IsMapEntry() {
		return
	}
	for _, subM := range m.Messages {
		g.genMessage(subM)
	}

	g.F("func (z *%s) Clone() *%s{", m.GoIdent, m.GoIdent)
	g.F("if z == nil {")
	g.F("return nil")
	g.F("}")
	g.F("zz := &%s{}", m.GoIdent)
	zIndex := 0
	for _, f := range m.Fields {
		target := ""
		switch {
		case f.Desc.IsList():
			{
				index := zIndex
				zIndex++
				target = fmt.Sprintf("zz%d", index)
				g.F("%s := make(%s, len(z.%s))", target, protoutil.FieldGoType(g.Q, f), f.GoName)
				g.F("for i := range z.%s {", f.GoName)
				switch f.Desc.Kind() {
				case protoreflect.MessageKind, protoreflect.GroupKind:
					g.F("%s[i] = z.%s[i].Clone()", target, f.GoName)
				default:
					g.F("%s[i] = z.%s[i]", target, f.GoName)
				}
				g.F("}")
			}
		case f.Desc.IsMap():
			{
				index := zIndex
				zIndex++
				target = fmt.Sprintf("zz%d", index)
				g.F("%s := make(%s)", target, protoutil.FieldGoType(g.Q, f))
				g.F("for k, v := range z.%s {", f.GoName)
				switch f.Message.Fields[1].Desc.Kind() {
				case protoreflect.MessageKind, protoreflect.GroupKind:
					g.F("%s[k] = v.Clone()", target)
				default:
					g.F("%s[k] = v", target)
				}
				g.F("}")
			}
		default:
			switch f.Desc.Kind() {
			case protoreflect.MessageKind, protoreflect.GroupKind:
				target = fmt.Sprintf("z.%s.Clone()", f.GoName)
			default:
				target = fmt.Sprintf("z.%s", f.GoName)
			}
		}
		g.F("zz.%s = %s", f.GoName, target)
	}
	g.F("	return zz")
	g.F("}")
}
