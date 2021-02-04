package genutil

import (
	"bytes"
	"fmt"

	"google.golang.org/protobuf/compiler/protogen"
)

type Generator interface {
	Generate() error
}

type G struct {
	*protogen.GeneratedFile
	preBuf  bytes.Buffer
	buf     bytes.Buffer
	postBuf bytes.Buffer
	gen     Generator
}

func New(file *protogen.GeneratedFile, gen func(*G) Generator) *G {
	g := &G{
		GeneratedFile: file,
	}
	g.gen = gen(g)
	return g
}

func (g *G) F(format string, a ...interface{}) {
	for i := range a {
		switch x := a[i].(type) {
		case protogen.GoIdent:
			a[i] = g.Q(x)
		}
	}
	g.P(fmt.Sprintf(format, a...))
}

func (g *G) P(v ...interface{}) {
	for _, x := range v {
		switch x := x.(type) {
		case protogen.GoIdent:
			_, _ = fmt.Fprint(&g.buf, g.QualifiedGoIdent(x))
		default:
			_, _ = fmt.Fprint(&g.buf, x)
		}
	}
	_, _ = fmt.Fprintln(&g.buf)
}

func (g *G) Q(ident protogen.GoIdent) string {
	return g.QualifiedGoIdent(ident)
}

func (g *G) Generate() error {
	if err := g.gen.Generate(); err != nil {
		return err
	}
	_, _ = g.Write(g.preBuf.Bytes())
	_, _ = g.Write(g.buf.Bytes())
	_, _ = g.Write(g.postBuf.Bytes())
	return nil
}

func (g *G) Pre(format string, a ...interface{}) {
	for i := range a {
		switch x := a[i].(type) {
		case protogen.GoIdent:
			a[i] = g.Q(x)
		}
	}
	_, _ = fmt.Fprintf(&g.preBuf, format, a...)
}

func (g *G) Post(format string, a ...interface{}) {
	for i := range a {
		switch x := a[i].(type) {
		case protogen.GoIdent:
			a[i] = g.Q(x)
		}
	}
	_, _ = fmt.Fprintf(&g.postBuf, format, a...)
}
