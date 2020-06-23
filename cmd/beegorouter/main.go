package main

import (
	"bytes"
	"flag"
	"go/ast"
	"go/format"
	"go/parser"
	"go/token"
	"io"
	"io/ioutil"
	"os"
	"strconv"
	"strings"
)

const beegoroutablePkgName = "github.com/icattlecoder/beegoroutable"

type Visitor struct {
}

func getNode(pkg, obj, method string) ast.Expr {

	expr := ast.CallExpr{
		Fun: &ast.Ident{
			Name: "MappingMethods",
		},
	}

	var args []ast.Expr
	method = strings.Trim(method, `"`)
	for _, m := range strings.Split(method, ";") {
		ss := strings.Split(m, ":")
		httpMethod, funcName := ss[0], ss[1]
		args = append(args, &ast.CallExpr{
			Fun: &ast.Ident{
				Name: strings.ToUpper(httpMethod),
			},

			Args: []ast.Expr{
				&ast.SelectorExpr{
					X: &ast.SelectorExpr{
						X: &ast.Ident{
							Name: pkg,
						},
						Sel: &ast.Ident{
							Name: "Default" + obj,
						},
					},
					Sel: &ast.Ident{
						Name: funcName,
					},
				},
			},
		})
	}
	expr.Args = args
	return &expr
}

func (v *Visitor) addImport(genDecl *ast.GenDecl) {
	hasImported := false
	for _, v := range genDecl.Specs {
		imptSpec := v.(*ast.ImportSpec)
		if imptSpec.Path.Value == strconv.Quote(beegoroutablePkgName) {
			hasImported = true
		}
	}
	if !hasImported {
		genDecl.Specs = append(genDecl.Specs, &ast.ImportSpec{
			Name: &ast.Ident{
				Name: ".",
			},
			Path: &ast.BasicLit{
				Kind:  token.STRING,
				Value: strconv.Quote(beegoroutablePkgName),
			},
		})
	}
}

func (v *Visitor) Visit(node ast.Node) ast.Visitor {

	switch n := node.(type) {

	case *ast.GenDecl:
		genDecl := node.(*ast.GenDecl)
		if genDecl.Tok == token.IMPORT {
			v.addImport(genDecl)
			return nil
		}
	case *ast.CallExpr:
		if e, ok := n.Fun.(*ast.SelectorExpr); ok && e.Sel.Name == "NSRouter" {

			if len(n.Args) < 3 {
				return v
			}

			cl := n.Args[1].(*ast.UnaryExpr).X.(*ast.CompositeLit)
			clse := cl.Type.(*ast.SelectorExpr)

			pkg := clse.X.(*ast.Ident)
			obj := clse.Sel.Name

			bl, ok := n.Args[2].(*ast.BasicLit)
			if !ok {
				return v
			}
			n.Args[2] = getNode(pkg.Name, obj, bl.Value)
			return nil
		}
	}

	return v
}

var (
	input     = flag.String("input", "", "input file")
	overwrite = flag.Bool("o", false, "overwrite source file")
)

func main() {

	flag.Parse()
	filename := *input
	if filename == "" {
		flag.PrintDefaults()
		return
	}

	fset := token.NewFileSet()
	f, err := parser.ParseFile(fset, filename, nil, parser.AllErrors)
	if err != nil {
		panic(err)
	}
	ast.Walk(&Visitor{}, f)

	buf := bytes.Buffer{}

	if err := format.Node(&buf, fset, f); err != nil {
		panic(err)
	}

	if !*overwrite {
		io.Copy(os.Stdout, &buf)
		return
	}

	tempFile, err := ioutil.TempFile(os.TempDir(), "beegoroutable-")
	if err != nil {
		panic(err)
	}
	if _, err := io.Copy(tempFile, &buf); err != nil {
		panic(err)
	}
	if err := os.Rename(tempFile.Name(), filename); err != nil {
		panic(err)
	}
}
