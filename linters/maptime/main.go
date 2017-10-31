package main

import (
	"flag"
	"fmt"
	"go/ast"
	"go/build"
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
	tags := flag.String("tags", "", "List of build tags to take into account when linting.")
	skipVendor := flag.Bool("skip-vendor", true, "Skip vendor directors.")

	flag.Parse()
	importPaths := gotool.ImportPaths(flag.Args())
	if len(importPaths) == 0 {
		flag.Usage()
		return
	}

	var filteredPaths []string
	if *skipVendor {
		filteredPaths = filterOutVendor(importPaths)
	} else {
		filteredPaths = importPaths
	}

	handleImportPaths(filteredPaths, strings.Fields(*tags), func(position token.Position, keyStr, valStr string) {
		fmt.Printf(
			"%s: Reconsider use of map[%s]%s . Storing an instance of time.Time as part of a map key is not recommended.\n",
			position.String(),
			keyStr,
			valStr,
		)
	})
}

func filterOutVendor(importPaths []string) []string {
	filteredStrings := []string{}
	for _, importPath := range importPaths {
		if !strings.Contains(importPath, "/vendor/") {
			filteredStrings = append(filteredStrings, importPath)
		}
	}
	return filteredStrings
}

func handleImportPaths(importPaths []string, buildTags []string, callback callbackFunc) {
	fs := token.NewFileSet()

	ctx := build.Default
	ctx.BuildTags = buildTags

	conf := loader.Config{
		Fset:  fs,
		Build: &ctx,
	}

	for _, importPath := range importPaths {
		conf.Import(importPath)
	}

	prog, err := conf.Load()
	if err != nil {
		log.Fatal(err)
	}

	for _, pkg := range prog.InitialPackages() {
		for _, file := range pkg.Files {
			ast.Walk(newMapVisitor(fs, pkg.Types, callback), file)
		}
	}
}

func newMapVisitor(fs *token.FileSet, types map[ast.Expr]types.TypeAndValue, callback callbackFunc) mapVisitor {
	return mapVisitor{
		fs:    fs,
		types: types,
		callback: func(position token.Position, keyStr, valStr string) {
			callback(position, path.Base(keyStr), path.Base(valStr))
		},
	}
}

type mapVisitor struct {
	fs       *token.FileSet
	types    map[ast.Expr]types.TypeAndValue
	callback callbackFunc
}

func (v mapVisitor) Visit(node ast.Node) ast.Visitor {
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
	if ok && structType.NumFields() == 3 {
		// VERSION <= go 1.8.X
		if structType.Field(0).Name() == "sec" &&
			structType.Field(0).Type().String() == "int64" &&
			structType.Field(1).Name() == "nsec" &&
			structType.Field(1).Type().String() == "int32" &&
			structType.Field(2).Name() == "loc" &&
			structType.Field(2).Type().String() == "*time.Location" {
			v.callback(position, keyStr, valStr)
			return nil
		}

		// VERSION >= go 1.9.X
		if structType.Field(0).Name() == "wall" &&
			structType.Field(0).Type().String() == "uint64" &&
			structType.Field(1).Name() == "ext" &&
			structType.Field(1).Type().String() == "int64" &&
			structType.Field(2).Name() == "loc" &&
			structType.Field(2).Type().String() == "*time.Location" {
			v.callback(position, keyStr, valStr)
			return nil
		}
	}

	// Detects objects with nested time.Time I.E map[{inner: time.Time}]<T>
	underlyingType := mapType.Key().Underlying().String()
	if strings.Contains(underlyingType, "time.Time") {
		v.callback(position, keyStr, valStr)
		return nil
	}

	return v
}
