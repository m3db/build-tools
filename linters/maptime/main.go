// Copyright (c) 2017 Uber Technologies, Inc.
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in
// all copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
// THE SOFTWARE.

package main

import (
	"flag"
	"fmt"
	"go/ast"
	"go/build"
	"go/token"
	"go/types"
	"log"
	"strings"

	"golang.org/x/tools/go/loader"

	"github.com/kisielk/gotool"
)

type mapCallbackFunc func(loc token.Position, keyStr, valStr string)
type comparisonCallbackFunc func(loc token.Position, xStr, yStr string)

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
	}, nil)
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

func handleImportPaths(
	importPaths []string,
	buildTags []string,
	mapCallback mapCallbackFunc,
	comparisonCallback comparisonCallbackFunc,
) {
	fs := token.NewFileSet()

	ctx := build.Default
	ctx.BuildTags = buildTags

	conf := loader.Config{
		Fset:  fs,
		Build: &ctx,
	}

	for _, importPath := range importPaths {
		conf.ImportWithTests(importPath)
	}

	prog, err := conf.Load()
	if err != nil {
		log.Fatal(err)
	}

	for _, pkg := range prog.InitialPackages() {
		for _, file := range pkg.Files {
			ast.Walk(mapVisitor{
				fs:                 fs,
				types:              pkg.Types,
				mapCallback:        mapCallback,
				comparisonCallback: comparisonCallback,
			}, file)
		}
	}
}

type mapVisitor struct {
	fs                 *token.FileSet
	types              map[ast.Expr]types.TypeAndValue
	mapCallback        mapCallbackFunc
	comparisonCallback comparisonCallbackFunc
}

func (v mapVisitor) Visit(node ast.Node) ast.Visitor {
	binary, ok := node.(*ast.BinaryExpr)
	if ok {
		xType := v.types[binary.X].Type
		yType := v.types[binary.Y].Type
		if isTimeOrContainsTime(xType) && isTimeOrContainsTime(yType) {
			v.comparisonCallback(v.fs.Position(binary.Pos()), xType.String(), yType.String())
		}
		return nil
	}

	mapNode, ok := node.(*ast.MapType)
	if ok {
		mapType := v.types[mapNode].Type.(*types.Map)
		position := v.fs.Position(mapNode.Map)
		keyStr := mapType.Key().String()
		valStr := mapType.Elem().String()

		containsTime := isTimeOrContainsTime(mapType.Key())
		if containsTime {
			v.mapCallback(position, keyStr, valStr)
			return nil
		}
	}

	return v
}

// isTypeOrContainsTime returns whether the type x represents an instance of
// time.Time or contains a nested time.Time
func isTimeOrContainsTime(x types.Type) bool {
	typeStr := x.String()
	typeUnderlying := x.Underlying()

	// Detects map[time.Time]<T>
	if strings.Contains(typeStr, "time.Time") {
		return true
	}

	// Detects map[timeAlias]<T>
	structType, ok := typeUnderlying.(*types.Struct)
	if ok && structType.NumFields() == 3 {
		// VERSION <= go 1.8.X
		if structType.Field(0).Name() == "sec" &&
			structType.Field(0).Type().String() == "int64" &&
			structType.Field(1).Name() == "nsec" &&
			structType.Field(1).Type().String() == "int32" &&
			structType.Field(2).Name() == "loc" &&
			structType.Field(2).Type().String() == "*time.Location" {
			return true
		}

		// VERSION >= go 1.9.X
		if structType.Field(0).Name() == "wall" &&
			structType.Field(0).Type().String() == "uint64" &&
			structType.Field(1).Name() == "ext" &&
			structType.Field(1).Type().String() == "int64" &&
			structType.Field(2).Name() == "loc" &&
			structType.Field(2).Type().String() == "*time.Location" {
			return true
		}
	}

	// Detects objects with nested time.Time I.E map[{inner: time.Time}]<T>
	if strings.Contains(typeUnderlying.String(), "time.Time") {
		return true
	}

	return false
}
