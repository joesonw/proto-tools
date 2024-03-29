package main

import (
	"flag"

	"google.golang.org/protobuf/compiler/protogen"

	"github.com/joesonw/proto-tools/pkg/genutil"
	"github.com/joesonw/proto-tools/pkg/protoutil"
)

var (
	flags         flag.FlagSet
	importPrefix  = flags.String("import_prefix", "", "prefix to prepend to import paths")
	disableReturn = flags.Bool("disable-return", false, "disable return after set")
)

func main() {
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
			filename := file.GeneratedFilenamePrefix + "_setter.pb.go"
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
	g.P("// Code generated by protoc-gen-setter. DO NOT EDIT.")
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

	for _, f := range m.Fields {
		if *disableReturn {
			g.F("func (z *%s) Set%s(v %s) {", m.GoIdent, f.GoName, protoutil.FieldGoType(g.Q, f))
		} else {
			g.F("func (z *%s) Set%s(v %s) *%s {", m.GoIdent, f.GoName, protoutil.FieldGoType(g.Q, f), m.GoIdent)
		}
		g.F("z.%s = v", f.GoName)
		if !*disableReturn {
			g.F("return z")
		}
		g.F("}")
		switch {
		case f.Desc.IsList():
			if *disableReturn {
				g.F("func (z *%s) Append%s(v %s) {", m.GoIdent, f.GoName, protoutil.FieldGoType(g.Q, protoutil.ElemOfListField(f)))
			} else {
				g.F("func (z *%s) Append%s(v %s) *%s {", m.GoIdent, f.GoName, protoutil.FieldGoType(g.Q, protoutil.ElemOfListField(f)), m.GoIdent)
			}
			g.F("z.%s = append(z.%s, v)", f.GoName, f.GoName)
			if !*disableReturn {
				g.F("return z")
			}
			g.F("}")
		case f.Desc.IsMap():
			if *disableReturn {
				g.F("func (z *%s) Put%s(k %s, v %s) {", m.GoIdent, f.GoName, protoutil.FieldGoType(g.Q, f.Message.Fields[0]), protoutil.FieldGoType(g.Q, f.Message.Fields[1]))
			} else {
				g.F("func (z *%s) Put%s(k %s, v %s) *%s {", m.GoIdent, f.GoName, protoutil.FieldGoType(g.Q, f.Message.Fields[0]), protoutil.FieldGoType(g.Q, f.Message.Fields[1]), m.GoIdent)
			}
			g.F("z.%s[k] = v", f.GoName)
			if !*disableReturn {
				g.F("return z")
			}
			g.F("}")
		}
	}
}
