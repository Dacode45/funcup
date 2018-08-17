package main

import (
	"bytes"
	"fmt"
	"go/ast"
	"go/importer"
	"go/parser"
	"go/token"
	"go/types"
	"html/template"
	"log"
	"strings"
)

const testFile = `
package main

import "fmt"

type Strings []string

func main() {
	fmt.Println(Strings(nil))
}
`

func main() {
	fset := token.NewFileSet()

	// Parse input string
	f, err := parser.ParseFile(fset, "hello.go", testFile, 0)
	checkErr(err)

	conf := types.Config{Importer: importer.Default()}

	pkg, err := conf.Check("cmd/test", fset, []*ast.File{f}, nil)
	checkErr(err)
	PrintPackage(pkg)

	target := "Strings"

	obj := pkg.Scope().Lookup(target)
	if obj == nil {
		log.Fatalln("Cannot find target", target)
	}
	PrintObj(obj)

	var sliceOf types.Type
	if asSlice, ok := obj.Type().Underlying().(*types.Slice); !ok {
		log.Fatalln("Only works on slices")
	} else {
		sliceOf = asSlice.Elem()
	}

	g := Generator{
		target,
		pkg,
		obj,
		sliceOf,
	}
	buf, err := g.Run()
	checkErr(err)
	fmt.Println(buf.String())
}

func PrintPackage(pkg *types.Package) {
	fmt.Println("Print Package")
	fmt.Printf("\tPackage  %q\n", pkg.Path())
	fmt.Printf("\tName:    %s\n", pkg.Name())
	fmt.Printf("\tImports: %s\n", pkg.Imports())
	fmt.Printf("\tScope:   %s\n", pkg.Scope())
}

func PrintObj(obj types.Object) {
	fmt.Println("Print Object")
	fmt.Printf("\tName %s\n", obj.Name())
	fmt.Printf("\tType %s\n", obj.Type().Underlying())
}

func checkErr(err error) {
	if err != nil {
		log.Fatal(err)
	}
}

// Generator is used to generate functional programming stuff for the type
type Generator struct {
	target  string
	pkg     *types.Package
	obj     types.Object
	sliceOf types.Type
}

// Package returns the package name
func (g Generator) Package() string {
	return g.pkg.Name()
}

// Recv returns the name of reciver to use for methods
func (g Generator) Recv() string {
	return string(strings.ToLower(g.target)[0])
}

// Type returns the name of the target
func (g Generator) Type() string {
	return g.obj.Name()
}

// SliceOf returns the Type of the elements of the slice
func (g Generator) SliceOf() string {
	return g.sliceOf.String()
}

// Run executes the template
func (g Generator) Run() (bytes.Buffer, error) {
	var buf bytes.Buffer
	if err := funcTmpl.Execute(&buf, g); err != nil {
		return buf, err
	}
	return buf, nil
}

var funcTmpl = template.Must(template.New("func").Parse(
	`//Code generated automatically with github.com/Dacode45/funcup. DO NOT EDIT.
package {{.Package}}

func ({{.Recv}} {{.Type}}) ForEach(cb func(element {{.SliceOf}}, index int, slice {{.Type}}) error {
	for i, element := range {{.Recv}} {
		err := cb(element, i, {{.Recv}})
		if err != nil {
			break;
		}
	}
})`))
