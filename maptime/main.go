package main

import (
	"flag"
	"fmt"
	"go/ast"
	"go/token"
	"go/types"
	"log"
	"path"
	"strings"

	"golang.org/x/tools/go/loader"

	"github.com/kisielk/gotool"
)

type callbackFunc func(loc token.Position, keyStr, valStr string)

func main() {
	flag.Parse()
	importPaths := gotool.ImportPaths(flag.Args())
	if len(importPaths) == 0 {
		return
	}

	handleImportPaths(importPaths, func(position token.Position, keyStr, valStr string) {
		fmt.Printf("%s: Reconsider use of map[%s]%s . Storing an instance of time.Time as part of a map key is not recommended.\n", position.String(), keyStr, valStr)
	})
}

func handleImportPaths(importPaths []string, callback callbackFunc) {
	fs := token.NewFileSet()
	var conf loader.Config
	conf.Fset = fs

	for _, importPath := range importPaths {
		conf.Import(importPath)
	}

	prog, err := conf.Load()
	if err != nil {
		log.Fatal(err)
	}

	for _, pkg := range prog.InitialPackages() {
		for _, file := range pkg.Files {
			ast.Walk(astVisitor{fs: fs, types: pkg.Types, callback: callback}, file)
		}
	}
}

type astVisitor struct {
	fs       *token.FileSet
	types    map[ast.Expr]types.TypeAndValue
	callback callbackFunc
}

func (v astVisitor) Visit(node ast.Node) ast.Visitor {
	mapNode, ok := node.(*ast.MapType)
	if !ok {
		return v
	}

	mapType := v.types[mapNode].Type.(*types.Map)
	position := v.fs.Position(mapNode.Map)
	keyStr := mapType.Key().String()
	valStr := mapType.Elem().String()

	// Detects map[time.Time]<T>
	if strings.Contains(keyStr, "time.Time") {
		v.callback(position, keyStr, valStr)
		return nil
	}

	// Detects map[timeAlias]<T>
	structType, ok := mapType.Key().Underlying().(*types.Struct)
	if ok {
		if structType.Field(0).Name() == "sec" &&
			structType.Field(0).Type().String() == "int64" &&
			structType.Field(1).Name() == "nsec" &&
			structType.Field(1).Type().String() == "int32" &&
			structType.Field(2).Name() == "loc" &&
			structType.Field(2).Type().String() == "*time.Location" {
			v.callback(position, path.Base(keyStr), valStr)
			return nil
		}
	}

	// Detects objects with nested time.Time I.E map[{inner: time.Time}]<T>
	underlyingType := mapType.Key().Underlying().String()
	if strings.Contains(underlyingType, "time.Time") {
		v.callback(position, path.Base(keyStr), valStr)
		return nil
	}

	return v
}
