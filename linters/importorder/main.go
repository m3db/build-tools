// Copyright (c) 2018 Uber Technologies, Inc.
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
	"errors"
	"flag"
	"fmt"
	"go/ast"
	"go/build"
	"go/parser"
	"go/token"
	"go/types"
	"log"
	"sort"
	"strings"

	"github.com/kisielk/gotool"
	"golang.org/x/tools/go/loader"
)

const (
	standardImportGroup = "STDLIB"
	externalImportGroup = "EXTERNAL"
)

var (
	errMultipleImport = errors.New("more than one import declaration found")
	errOutOfOrder     = errors.New("imports are out of order")

	defaultPattern = fmt.Sprintf("%s %s", standardImportGroup, externalImportGroup)
)

type lintError struct {
	fileName                   string
	goldStandard, originalDecl importDecl
	err                        error
}

type lintErrors []lintError

func main() {
	tags := flag.String("tags", "", "List of build tags to take into account when linting.")
	skipVendor := flag.Bool("skip-vendor", true, "Skip vendor directors.")
	rawPatterns := flag.String("patterns", defaultPattern, "Specify the patterns of each group in order. If checking for Go standard imports write `STDLIB`, if checking for a wildard group write `EXTERNAL`.")
	verbose := flag.Bool("verbose", false, "If imports are out of order, determines whether we return just an error (false) or the full comparison list (true).")

	flag.Parse()
	importPaths := gotool.ImportPaths(flag.Args())
	if len(importPaths) == 0 {
		flag.Usage()
		return
	}

	patterns := strings.Fields(*rawPatterns)
	if len(patterns) < 1 {
		log.Fatal("List of patterns must be greater than 0\n")
	}

	filteredPaths := importPaths
	if *skipVendor {
		filteredPaths = filterOutVendor(filteredPaths)
	}

	groupedErrors := handleImportPaths(filteredPaths, strings.Fields(*tags), patterns)
	printErrors(verbose, groupedErrors)
}

func printErrors(verbose *bool, groupedErrors lintErrors) {
	if *verbose {
		for _, imp := range groupedErrors {
			fmt.Printf("%s:%v: import groups should look like:\n%v\n", imp.fileName, imp.err, imp.goldStandard)
		}
	} else {
		for _, imp := range groupedErrors {
			fmt.Printf("%s: %v.\n", imp.fileName, imp.err)
		}
	}
}

func handleImportPaths(importPaths []string, buildTags, patterns []string) lintErrors {
	fs := token.NewFileSet()

	ctx := build.Default
	ctx.BuildTags = buildTags

	conf := loader.Config{
		Fset:  fs,
		Build: &ctx,
		// Since we are not concerned with the entire file, we should only parse the imports
		ParserMode: parser.ImportsOnly,
		// Continue even if type or IO errors are present
		AllowErrors: true,
		TypeChecker: types.Config{
			Error: func(e error) {},
		},
	}

	for _, importPath := range importPaths {
		conf.ImportWithTests(importPath)
	}

	prog, err := conf.Load()
	if err != nil {
		log.Fatal(err)
	}

	var groupedLintErrors lintErrors
	for _, pkg := range prog.InitialPackages() {
		for _, file := range pkg.Files {
			imports := imports(fs, file)
			if len(imports) == 0 {
				continue
			}
			validateImportDecl(imports)
			goldStandard, err := getGoldStandard(imports, patterns)
			if err != nil {
				groupedLintErrors = append(groupedLintErrors, lintError{err: err, fileName: fs.Position(file.Pos()).Filename})
				continue
			}
			if !compareImports(goldStandard, imports[0]) {
				groupedLintErrors = append(groupedLintErrors, lintError{fileName: fs.Position(file.Pos()).Filename,
					originalDecl: imports[0],
					goldStandard: goldStandard,
					err:          errOutOfOrder,
				})
			}
		}
	}
	return groupedLintErrors
}

func compareImports(goldStandard, originalImportDecl importDecl) bool {
	goldGroups := goldStandard.Groups
	originalGroups := originalImportDecl.Groups

	if len(goldGroups) != len(originalGroups) {
		return false
	}

	for i, goldGroup := range goldGroups {
		for j, goldImport := range goldGroup.Imports {
			if goldImport != originalGroups[i].Imports[j] {
				return false
			}
		}
	}
	return true
}

func getGoldStandard(imports []importDecl, patterns []string) (importDecl, error) {
	var emptyImportDecl importDecl
	if len(imports) > 1 {
		return emptyImportDecl, errMultipleImport
	}

	if len(imports) == 0 {
		return emptyImportDecl, nil
	}

	combinedImports := concatenateImports(imports[0])
	goldStandard := createGoldStandard(combinedImports, patterns)

	return goldStandard, nil
}

func createGoldStandard(imports []importSpec, patterns []string) importDecl {
	groups := make([]importGroup, 0, len(patterns))
	importMap := convertToMap(imports)

	for i, pattern := range patterns {
		var tempGroup []importSpec
		for imp := range importMap {
			switch {
			case pattern == standardImportGroup:
				tempGroup = orderStdLib(tempGroup, imp, importMap)
			case pattern == externalImportGroup:
				tempGroup = orderExt(tempGroup, imp, importMap, patterns, i)
			case strings.Contains(imp.Path, pattern):
				tempGroup = append(tempGroup, imp)
			default:
				// no need to do anything special here since an import that doesn't match
				// any pattern should cause the linter to fail
			}
			delete(importMap, imp)
		}
		sort.Sort(importSpecs(tempGroup))
		if len(tempGroup) > 0 {
			importGroup := importGroup{Imports: tempGroup}
			groups = append(groups, importGroup)
		}
	}
	return importDecl{
		Groups: groups,
	}
}

func orderExt(tempGroup []importSpec, imp importSpec, importMap map[importSpec]struct{}, patterns []string, i int) []importSpec {
	if i == len(patterns)-1 {
		if strings.Contains(imp.Path, ".") {
			tempGroup = append(tempGroup, imp)
		}
	} else {
		for _, pattern := range patterns[i+1:] {
			if !strings.Contains(imp.Path, pattern) {
				tempGroup = append(tempGroup, imp)
			}
		}
	}
	return tempGroup
}

func orderStdLib(tempGroup []importSpec, imp importSpec, importMap map[importSpec]struct{}) []importSpec {
	if !isThirdParty(imp.Path) {
		tempGroup = append(tempGroup, imp)
	}
	return tempGroup
}

func convertToMap(imports []importSpec) map[importSpec]struct{} {
	importsMap := make(map[importSpec]struct{})
	for _, imp := range imports {
		importsMap[imp] = struct{}{}
	}
	return importsMap
}

func concatenateImports(imports importDecl) []importSpec {
	var combinedImports []importSpec
	for _, group := range imports.Groups {
		for _, imp := range group.Imports {
			combinedImports = append(combinedImports, imp)
		}
	}
	return combinedImports
}

// importDecl is the collection of importGroups contained in a single import block.
type importDecl struct {
	Groups []importGroup
}

// importGroup is a collection of imports
type importGroup struct {
	Imports importSpecs
}

type importSpecs []importSpec

func (imp importSpecs) Len() int {
	return len(imp)
}

func (imp importSpecs) Swap(i, j int) {
	imp[i], imp[j] = imp[j], imp[i]
}

func (imp importSpecs) Less(i, j int) bool {
	return imp[i].Path < imp[j].Path
}

// importSpec is a single import
type importSpec struct {
	Position token.Position
	Line     int
	Name     string
	Path     string
}

// Imports returns the file imports grouped by paragraph.
func imports(fset *token.FileSet, f *ast.File) []importDecl {
	var importDecls []importDecl

	for _, decl := range f.Decls {
		genDecl, ok := decl.(*ast.GenDecl)
		if !ok || genDecl.Tok != token.IMPORT {
			continue
		}

		var (
			importDecl importDecl
			group      importGroup
		)

		var lastLine int
		for _, spec := range genDecl.Specs {
			importSpec := spec.(*ast.ImportSpec)
			pos := importSpec.Path.ValuePos
			line := fset.Position(pos).Line
			fileName := fset.Position(pos).Filename
			if lastLine > 0 && pos > 0 && line-lastLine > 1 {
				importDecl.Groups = append(importDecl.Groups, group)
				group = importGroup{}
			}
			group.Imports = append(group.Imports, newImportSpec(importSpec, line, fileName))
			lastLine = line
		}
		importDecl.Groups = append(importDecl.Groups, group)
		importDecls = append(importDecls, importDecl)
	}

	return importDecls
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

func newImportSpec(is *ast.ImportSpec, line int, filename string) importSpec {
	var (
		pathLit = is.Path
		path    string
	)

	if pathLit != nil {
		path = pathLit.Value
	}

	return importSpec{
		Name: filename,
		Path: path,
		Line: line,
	}
}

func validateImportDecl(importDecls []importDecl) {
	for _, importDecl := range importDecls {
		if len(importDecl.Groups) == 0 {
			log.Fatal(fmt.Errorf("import group cannot be empty"))
		}
		for _, group := range importDecl.Groups {
			if len(group.Imports) == 0 {
				log.Fatal(fmt.Errorf("imports cannot be empty"))
			}
		}
	}
}

func isThirdParty(path string) bool {
	return strings.Contains(path, ".")
}
