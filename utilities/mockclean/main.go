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
	"bytes"
	"flag"
	"fmt"
	"go/ast"
	"go/format"
	"go/parser"
	"go/printer"
	"go/token"
	"io/ioutil"
	"log"
	"os"
	"regexp"
	"sort"
	"strconv"
	"strings"

	xlog "github.com/m3db/m3x/log"

	"golang.org/x/tools/go/ast/astutil"
)

var (
	pkg            = flag.String("pkg", "", "full package mock is being generated for, e.g. github.com/m3db/m3db/client")
	in             = flag.String("in", defaultInputStdin, "input path for mock being read, '-' for stdin")
	out            = flag.String("out", "-", `file path for mock being written, "-" for stdout`)
	perm           = flag.String("perm", "666", "permissions to write file with")
	groupPrefixes  = flag.String("prefixes", defaultGroupPrefixes, "prefixes to group imports by")
	selfRefCleanup = flag.Bool("cleanup-selfref", true, "cleanup self referrential imports")
	importCleanup  = flag.Bool("cleanup-import", true, "cleanup import aliasing and ordering")
)

const (
	defaultInputStdin    = "-"
	defaultGroupPrefixes = "github.com/m3db"
)

func main() {
	logger, err := xlog.Configuration{}.BuildLogger()
	if err != nil {
		log.Fatalf("unable to build logger: %v", err)
	}

	flag.Parse()

	newFileMode, err := parseNewFileMode(*perm)
	if err != nil {
		logger.Errorf("perm: %v", err)
	}

	if len(*pkg) == 0 || len(*in) == 0 || len(*out) == 0 || err != nil {
		flag.Usage()
		os.Exit(1)
	}

	var inputData []byte
	if *in == defaultInputStdin {
		inputData, err = ioutil.ReadAll(os.Stdin)
	} else {
		inputData, err = ioutil.ReadFile(*in)
	}
	if err != nil {
		logger.Fatalf("unable to read input: %v", err)
	}

	if *selfRefCleanup {
		inputData, err = removeSelfReferrentialImports(inputData, *pkg)
		if err != nil {
			logger.Fatalf("unable to cleanup self referrential imports: %v", err)
		}
	}

	if *importCleanup {
		inputData, err = cleanupImports(inputData)
		if err != nil {
			logger.Fatalf("unable to cleanup imports: %v", err)
		}

		prefixes := strings.Fields(*groupPrefixes)
		inputData, err = reorderImports(inputData, prefixes)
		if err != nil {
			logger.Fatalf("unable to reorder imports: %v", err)
		}
	}

	if *out == "-" {
		_, err = fmt.Printf("%s\n", string(inputData))
	} else {
		err = ioutil.WriteFile(*out, inputData, newFileMode)
	}
	if err != nil {
		logger.Fatalf("unable to write output to %s: %v", *out, err)
	}
}

func removeSelfReferrentialImports(src []byte, packageName string) ([]byte, error) {
	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, "", src, parser.PackageClauseOnly)
	if err != nil {
		return nil, err
	}
	if file.Name == nil {
		return nil, fmt.Errorf("unable to parse package")
	}
	basePkg := file.Name.Name

	strs := []string{
		fmt.Sprintf("%s \"%s\"", basePkg, packageName), "",
		fmt.Sprintf("%s.", basePkg), "",
		fmt.Sprintf("\"%s\"", *pkg), "",
	}
	for i := 0; i+1 < len(strs); i += 2 {
		fmt.Printf("replace:\n%s\n\n", strs[i])
	}

	replacer := strings.NewReplacer(
		// Replace any self referential imports
		fmt.Sprintf("%s \"%s\"", basePkg, packageName), "",
		fmt.Sprintf("%s.", basePkg), "",
		fmt.Sprintf("\"%s\"", *pkg), "")
	return []byte(replacer.Replace(string(src))), nil
}

// reorderImports re-orders imports into groups following the convention below:
// import (
// 	 stdlib
//
//   userPrefixes[0]
//   ...
//   userPrefixes[n]
//
//   third-party
//  )
func reorderImports(src []byte, userPrefixes []string) ([]byte, error) {
	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, "", src, parser.ParseComments)
	if err != nil {
		return nil, err
	}

	// extract the import into required groups
	standardLibImports := []ast.ImportSpec{}
	thirdPartyImports := []ast.ImportSpec{}
	userPrefixImports := make(map[string][]ast.ImportSpec)
	for _, im := range file.Imports {
		if isStdlibPackage(im.Path.Value) {
			standardLibImports = append(standardLibImports, *im)
			continue
		}

		userImport := false
		for _, pre := range userPrefixes {
			if strings.Contains(im.Path.Value, pre) {
				userPrefixImports[pre] = append(userPrefixImports[pre], *im)
				userImport = true
				break
			}
		}

		// default to third-party import if others are not found
		if !userImport {
			thirdPartyImports = append(thirdPartyImports, *im)
		}
	}

	// sort all the imports
	sort.Sort(importSpecs(standardLibImports))
	sort.Sort(importSpecs(thirdPartyImports))
	for _, im := range userPrefixImports {
		sort.Sort(importSpecs(im))
	}

	// now we can replace all the import statements in the original source with
	// new single import block and associated decls.
	removeExistingImports := func(c *astutil.Cursor) bool {
		node := c.Node()
		decl, ok := node.(*ast.GenDecl)
		if !ok {
			return true
		}

		if decl.Tok != token.IMPORT {
			return true
		}

		c.Delete()
		return false
	}
	cleaned := astutil.Apply(file, removeExistingImports, nil)

	// convert import decls into correct structure
	generateImports := func(packagePos token.Pos) []byte {
		var buff bytes.Buffer
		buff.WriteString("import (\n")
		insertNewLineBeforeUsage := false
		writeImport := func(im ast.ImportSpec) {
			insertNewLineBeforeUsage = true
			if im.Name != nil {
				buff.WriteString(fmt.Sprintf("\t%s %s\n", im.Name.Name, im.Path.Value))
			} else {
				buff.WriteString(fmt.Sprintf("\t%s\n", im.Path.Value))
			}
		}
		for _, im := range standardLibImports {
			writeImport(im)
		}

		for _, pre := range userPrefixes {
			imports := userPrefixImports[pre]
			if len(imports) > 0 && insertNewLineBeforeUsage {
				buff.WriteString("\n")
				insertNewLineBeforeUsage = false
			}
			for _, im := range userPrefixImports[pre] {
				writeImport(im)
			}
		}
		if len(thirdPartyImports) > 0 && insertNewLineBeforeUsage {
			buff.WriteString("\n")
		}
		for _, im := range thirdPartyImports {
			writeImport(im)
		}
		buff.WriteString(")")

		return buff.Bytes()
	}

	newImports := generateImports(file.Package)
	cleanedFile := cleaned.(*ast.File)

	var buf bytes.Buffer
	format.Node(&buf, fset, cleanedFile)
	re := regexp.MustCompile("package .*")
	cleanedWithImports := re.ReplaceAllStringFunc(buf.String(), func(x string) string {
		return fmt.Sprintf("%s\n\n%s", x, newImports)
	})
	return []byte(cleanedWithImports), nil
}

// cleanupImports makes the following changes:
// - if an import has the same alias as it's base package, it removes the alias
// 	 i.e. converts "import x abc/x" ==> "import abc/x"
// - if two packages share the same base package name, it prefers to alias the non-standard lib
//   i.e. it converts
// import (
//   fmt "x/fmt"
//   fmt0 "fmt"
// )
// into
// import (
//   "fmt"
//   fmt0 "x/fmt"
// )
func cleanupImports(src []byte) ([]byte, error) {
	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, "", src, parser.ParseComments)
	if err != nil {
		return nil, err
	}

	// basePkg -> []importStmt
	importMap := make(map[string][]ast.ImportSpec)
	for _, i := range file.Imports {
		base, err := strconv.Unquote(i.Path.Value)
		if err != nil {
			panic(err)
		}
		basePkg := extractBasePkg(base)
		arr, ok := importMap[basePkg]
		if !ok {
			importMap[basePkg] = []ast.ImportSpec{*i}
		} else {
			importMap[basePkg] = append(arr, *i)
		}
	}

	// track all import alias changes
	var swaps swapIdents

	// rewrite the ast for each import that has an issue
	for basePkg, imports := range importMap {
		// can never have len(imports) == 0, asserting that to be sure
		if len(imports) == 0 {
			return nil, fmt.Errorf("illegal internal state: %+v", importMap)
		}

		// i.e. we only have a single import with this basePkg, and we're still aliasing it
		// can just drop the alias in this case
		if len(imports) == 1 && imports[0].Name != nil && imports[0].Name.Name == basePkg {
			path := mustUnquote(imports[0].Path.Value)
			astutil.DeleteNamedImport(fset, file, basePkg, path)
			astutil.AddImport(fset, file, path)
			continue
		}

		// i.e. we have >= 2 imports with the same basePkg. need to ensure that the standard
		// library version has no alias.
		var (
			stdLibImport, otherPackageWithBaseAlias *ast.ImportSpec
			redundantAlias                          = false
		)
		for _, im := range imports {
			im := im
			// i.e. stdlib package has a redundant alias
			if isStdlibPackage(im.Path.Value) && im.Name != nil && im.Name.Name == basePkg {
				astutil.DeleteNamedImport(fset, file, basePkg, mustUnquote(im.Path.Value))
				astutil.AddImport(fset, file, mustUnquote(im.Path.Value))
				redundantAlias = true
				break
			}
			// i.e. stdlib has an extra alias
			if isStdlibPackage(im.Path.Value) && im.Name != nil && im.Name.Name != basePkg {
				stdLibImport = &im
				continue
			}
			if im.Name != nil && im.Name.Name == basePkg {
				otherPackageWithBaseAlias = &im
				continue
			}
		}
		if redundantAlias || stdLibImport == nil || otherPackageWithBaseAlias == nil {
			continue
		}
		astutil.DeleteNamedImport(fset, file, stdLibImport.Name.Name, mustUnquote(stdLibImport.Path.Value))
		astutil.DeleteNamedImport(fset, file, basePkg, mustUnquote(otherPackageWithBaseAlias.Path.Value))
		astutil.AddImport(fset, file, mustUnquote(stdLibImport.Path.Value))
		astutil.AddNamedImport(
			fset, file, stdLibImport.Name.Name, mustUnquote(otherPackageWithBaseAlias.Path.Value))
		swaps = append(swaps, swapIdent{
			x: stdLibImport.Name.Name,
			y: basePkg,
		})
	}

	// print current state of file, with new imports
	var buf bytes.Buffer
	printer.Fprint(&buf, fset, file)

	// perform all import alias changes
	bytes := swaps.re().ReplaceAllFunc(buf.Bytes(), func(input []byte) []byte {
		padded := func(x string) string { return fmt.Sprintf("%s.", x) }
		str := string(input)
		for _, s := range swaps {
			if padded(s.x) == str {
				return []byte(padded(s.y))
			} else if padded(s.y) == str {
				return []byte(padded(s.x))
			}
		}
		return input
	})
	return bytes, nil
}

func mustUnquote(s string) string {
	o, err := strconv.Unquote(s)
	if err != nil {
		panic(err)
	}
	return o
}

func isThirdParty(importPath string) bool {
	// Third party package import path usually contains "." (".com", ".org", ...)
	// This logic is taken from golang.org/x/tools/imports package.
	return strings.Contains(importPath, ".")
}

func isStdlibPackage(importPath string) bool {
	return !isThirdParty(importPath)
}

func extractBasePkg(s string) string {
	pkgParts := strings.Split(s, "/")
	return pkgParts[len(pkgParts)-1]
}

func parseNewFileMode(str string) (os.FileMode, error) {
	if len(str) != 3 {
		return 0, fmt.Errorf("file mode must be 3 chars long")
	}

	str = "0" + str

	var v uint32
	n, err := fmt.Sscanf(str, "%o", &v)
	if err != nil {
		return 0, fmt.Errorf("unable to parse: %v", err)
	}
	if n != 1 {
		return 0, fmt.Errorf("no value to parse")
	}
	return os.FileMode(v), nil
}

type swapIdent struct {
	x string
	y string
}

func (s swapIdent) re() string {
	return fmt.Sprintf("%s|%s", s.x, s.y)
}

type swapIdents []swapIdent

func (si swapIdents) re() *regexp.Regexp {
	strs := make([]string, 0, len(si))
	for _, s := range si {
		strs = append(strs, s.re())
	}
	chars := strings.Join(strs, "|")
	base := fmt.Sprintf(`\b(%s)\.`, chars)
	return regexp.MustCompile(base)
}

type importSpecs []ast.ImportSpec

func (a importSpecs) Len() int           { return len(a) }
func (a importSpecs) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a importSpecs) Less(i, j int) bool { return a[i].Path.Value < a[j].Path.Value }
