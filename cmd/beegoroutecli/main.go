package main

import (
	"errors"
	"flag"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strconv"

	"github.com/icattlecoder/beegoroutable"
)

type Handler struct {
	Method string
	Func   string
}

type Router struct {
	Path    string
	Handler []Handler
}

type Visitor struct {
	Prefix  string
	Routers []Router
	apis    []beegoroutable.Api
}

func (v *Visitor) HandleFound() {

	defer func() {
		v.Prefix = ""
		v.Routers = nil
	}()

	var apis []beegoroutable.Api

	for _, r := range v.Routers {
		for _, h := range r.Handler {

			api := beegoroutable.Api{
				Name:       h.Func,
				Path:       v.Prefix + r.Path,
				PathParams: nil,
				Params:     nil,
				Method:     h.Method,
				Body:       "",
			}
			apis = append(apis, api)
		}
	}
	v.apis = append(v.apis, apis...)
}

func (v *Visitor) Visit(node ast.Node) ast.Visitor {

	switch n := node.(type) {
	case *ast.CallExpr:
		if e, ok := n.Fun.(*ast.SelectorExpr); ok {
			ident, ok := e.X.(*ast.Ident)
			if !ok {
				break
			}
			if ident.Name == "beego" && e.Sel.Name == "NewNamespace" {
				v.HandleFound()
				s, _ := strconv.Unquote(n.Args[0].(*ast.BasicLit).Value)
				v.Prefix += s
			} else if ident.Name == "beego" && e.Sel.Name == "NSNamespace" {
				s, _ := strconv.Unquote(n.Args[0].(*ast.BasicLit).Value)
				v.Prefix += s
			} else if ident.Name == "beego" && e.Sel.Name == "NSRouter" {
				r := getRouter(n.Args)
				v.Routers = append(v.Routers, r)
				return nil
			}
		}
	}

	return v
}

func getRouter(args []ast.Expr) Router {

	r := Router{}

	r.Path, _ = strconv.Unquote(args[0].(*ast.BasicLit).Value)

	for _, node := range args[1:] {

		switch n := node.(type) {

		case *ast.CallExpr:
			if n.Fun.(*ast.Ident).Name != "MappingMethods" {
				break
			}
			for _, arg := range n.Args {
				callExpr, ok := arg.(*ast.CallExpr)
				if ok {
					h := Handler{}
					h.Method = callExpr.Fun.(*ast.Ident).Name
					h.Func = callExpr.Args[0].(*ast.SelectorExpr).Sel.Name
					r.Handler = append(r.Handler, h)
				}
			}
		}
	}
	return r
}

var (
	input  = flag.String("input", "", "input file")
	output = flag.String("output", "", "output file, default is stdout")
	pkg    = flag.String("pkg", "", "pkg name")
	name   = flag.String("name", "", "client name, like DeckJob")
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

	v := &Visitor{}
	ast.Walk(v, f)

	v.HandleFound()
	code, err := beegoroutable.GenerateCode(*pkg, *name, v.apis)
	if err != nil {
		log.Fatalln(err)
	}
	if *output == "" {
		fmt.Println(code)
		return
	}

	tempFile, err := ioutil.TempFile("", "")
	if err != nil {
		log.Fatalln(err)
	}
	if _, err := tempFile.WriteString(code); err != nil {
		log.Fatalln("WriteString err:", err)
	}

	dir, err := os.Getwd()
	if err != nil {
		log.Fatalln("Getwd err:", err)
	}
	outputFile := filepath.Join(dir, *output)
	if ensureDir(outputFile) != nil {
		log.Fatalln("ensure Dir err:", err)
	}

	if err := os.Rename(tempFile.Name(), outputFile); err != nil {
		log.Fatalln("rename err:", err)
	}
}

func ensureDir(f string) error {
	dir, _ := filepath.Split(f)
	s, err := os.Stat(dir)
	if err != nil {
		if !os.IsNotExist(err) {
			return err
		}
		if err := os.MkdirAll(dir, 0777); err != nil {
			return fmt.Errorf("mkdir %s err:%v", dir, err)
		}
		return nil
	}
	if !s.IsDir() {
		return errors.New(f + "is not dir")
	}
	return nil
}
