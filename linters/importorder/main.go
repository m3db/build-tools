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
	"errors"
	"flag"
	"fmt"
	"go/ast"
	"go/build"
	"go/parser"
	"go/token"
	"log"
	"os"
	"regexp"
	"strings"

	"github.com/kisielk/gotool"
	"golang.org/x/tools/go/loader"
)

const (
	standard = "STDLIB"
	external = "EXTERNAL"
)

var (
	errMultipleImport       = errors.New("more than one import declaration found")
	errOutOfOrder           = errors.New("import is out of order")
	errImportMatchedAlready = errors.New("import already matched pattern")
	errGroupMachedAlready   = errors.New("import group already matches previously seen pattern")
	errNoMatch              = errors.New("import does not match any of the provided patterns or number of groups exceeds number of patterns provided")
)

type lintError struct {
	fileName    string
	importName  string
	line        int
	patternSeen string
	err         error
}

type lintErrors []lintError

func main() {
	tags := flag.String("tags", "", "List of build tags to take into account when linting.")
	skipVendor := flag.Bool("skip-vendor", true, "Skip vendor directors.")
	rawPatterns := flag.String("patterns", "STDLIB EXTERNAL", "Specify the patterns of each group in order. If checking for Go standard imports write `STDLIB`, if checking for a wildard group write `EXTERNAL`.")

	flag.Parse()
	importPaths := gotool.ImportPaths(flag.Args())
	if len(importPaths) == 0 {
		flag.Usage()
		return
	}

	patterns := strings.Fields(*rawPatterns)
	if len(patterns) < 1 {
		fmt.Fprint(os.Stderr, "List of patterns must be greater than 0\n")
		os.Exit(1)
	}

	var filteredPaths []string
	if *skipVendor {
		filteredPaths = filterOutVendor(importPaths)
	} else {
		filteredPaths = importPaths
	}

	groupedErrors := handleImportPaths(filteredPaths, strings.Fields(*tags), patterns)
	printErrors(groupedErrors)
}

func printErrors(groupedErrors []lintErrors) {
	if len(groupedErrors) != 0 {
		for _, group := range groupedErrors {
			for _, imp := range group {
				fmt.Printf("File: %s\nImport: %s\nLine: %d\nPattern: %s\nError: %v\n\n", imp.fileName, imp.importName, imp.line, imp.patternSeen, imp.err)
			}
		}
	}
}

func handleImportPaths(importPaths []string, buildTags, patterns []string) []lintErrors {
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
	}

	for _, importPath := range importPaths {
		conf.ImportWithTests(importPath)
	}

	prog, err := conf.Load()
	if err != nil {
		log.Fatal(err)
	}

	var groupedLintErrors []lintErrors
	for _, pkg := range prog.InitialPackages() {
		for _, file := range pkg.Files {
			imports := imports(fs, file)
			if len(imports) == 0 {
				continue
			}
			fileLintErr := findErrors(imports, patterns, file)
			if fileLintErr != nil {
				groupedLintErrors = append(groupedLintErrors, fileLintErr)
			}
		}
	}
	return groupedLintErrors
}

// importDecl is the collection of importGroups contained in a single import block.
type importDecl struct {
	Groups []importGroup
}

// importGroup is a collection of imports
type importGroup struct {
	Imports []importSpec
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
			break
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
		path string
	)

	pathLit := is.Path
	if pathLit != nil {
		path = pathLit.Value
	}

	return importSpec{
		Name: filename,
		Path: path,
		Line: line,
	}
}

func findErrors(importDecls []importDecl, patterns []string, file *ast.File) lintErrors {
	var (
		lintErrors lintErrors

		// seenPatterns is used to keep track of patterns that have been seen already
		seenPatterns = make([]string, 0)
	)

	// check if there is more than one import declaration
	if len(importDecls) > 1 {
		lintErr := lintError{
			fileName: importDecls[0].Groups[0].Imports[0].Name,
			err:      errMultipleImport,
		}
		lintErrors = append(lintErrors, lintErr)
		return lintErrors
	}

	if len(importDecls) > 0 {
		return findRecErr(importDecls[0].Groups, patterns, lintErrors, seenPatterns)
	}

	return lintErrors
}

func findRecErr(importGroup []importGroup, currentPatterns []string, lintErrors lintErrors, seenPatterns []string) lintErrors {
	for _, group := range importGroup {
		var checking bool
		for i, importSpec := range group.Imports {
			// if there are still groups to check, but no more patterns left, throw an error
			if len(currentPatterns) == 0 && len(importGroup) > 0 {
				lintErrors = addLintError(importSpec, lintErrors, errNoMatch, "")
				return findRecErr(importGroup[1:], currentPatterns, lintErrors, seenPatterns)
			}
			// checking against "STDLIB"
			if currentPatterns[0] == standard {
				match := isThirdParty(importSpec.Path)
				// !match guarantees that the import is part of Go's standard library since it ensure there is no "." in the name
				if !match {
					if patternSeen(currentPatterns[0], seenPatterns) {
						lintErrors = addLintError(importSpec, lintErrors, errGroupMachedAlready, currentPatterns[0])
						return findRecErr(importGroup[1:], currentPatterns[1:], lintErrors, seenPatterns)
					}
					checking = true
				}
				// if we are checking against the standard library pattern, but the next import has a "." in it, we know it's out of order
				if match && checking {
					lintErrors = addLintError(importSpec, lintErrors, errOutOfOrder, currentPatterns[0])
				}
				// if we haven't seen a standard library import then we know that we should just skip to next pattern
				if match && !checking {
					return findRecErr(importGroup[0:], currentPatterns[1:], lintErrors, seenPatterns)
				}
				// check to see if we are at the last import of the group, then add "STDLIB" to list of seen patterns
				if i == len(group.Imports)-1 {
					seenPatterns = addSeenPattern(currentPatterns[0], seenPatterns)
					return findRecErr(importGroup[1:], currentPatterns[1:], lintErrors, seenPatterns)
				}
			}
			// checking against "EXTERNAL"
			if currentPatterns[0] == external {
				match := isThirdParty(importSpec.Path)
				// if we know there is a "." in the name, we need to make sure that the import doesn't match any of
				// the remaining patterns or any of the seen patterns
				if match {
					for _, seen := range seenPatterns {
						match, err := regexp.MatchString(seen, importSpec.Path)
						if err != nil {
							fmt.Fprintf(os.Stderr, "Unable to match: pattern, got error %v", err)
							return nil
						}
						if match {
							lintErrors = addLintError(importSpec, lintErrors, errImportMatchedAlready, seen)
							return findRecErr(importGroup[1:], currentPatterns[0:], lintErrors, seenPatterns)
						}
					}
					for _, pattern := range currentPatterns {
						match, err := regexp.MatchString(pattern, importSpec.Path)
						if err != nil {
							fmt.Fprintf(os.Stderr, "Unable to match: pattern, got error %v", err)
							return nil
						}
						if match && !checking {
							return findRecErr(importGroup[0:], currentPatterns[1:], lintErrors, seenPatterns)
						}
						if match && checking {
							lintErrors = addLintError(importSpec, lintErrors, errImportMatchedAlready, pattern)
							seenPatterns = addSeenPattern(currentPatterns[0], seenPatterns)
							return findRecErr(importGroup[1:], currentPatterns[1:], lintErrors, seenPatterns)
						}
					}
					checking = true
				}
				if !match && checking {
					lintErrors = addLintError(importSpec, lintErrors, errOutOfOrder, standard)
				}
				if !match && !checking {
					lintErrors = addLintError(importSpec, lintErrors, errOutOfOrder, standard)
					return findRecErr(importGroup[0:], currentPatterns[1:], lintErrors, seenPatterns)
				}
				if i == len(group.Imports)-1 {
					seenPatterns = addSeenPattern(currentPatterns[0], seenPatterns)
					return findRecErr(importGroup[1:], currentPatterns[1:], lintErrors, seenPatterns)
				}
			}
			// checking against all other patterns (i.e. not "STDLIB" or "EXTERNAL")
			if currentPatterns[0] != standard && currentPatterns[0] != external {
				for _, seen := range seenPatterns {
					switch seen {
					case standard:
						match := isThirdParty(importSpec.Path)
						if !match {
							lintErrors = addLintError(importSpec, lintErrors, errImportMatchedAlready, seen)
							return findRecErr(importGroup[1:], currentPatterns[0:], lintErrors, seenPatterns)
						}
					case external:
						match := isThirdParty(importSpec.Path)
						if match {
							allMatch := true
							for _, seen := range seenPatterns {
								match, err := regexp.MatchString(seen, importSpec.Path)
								if err != nil {
									fmt.Fprintf(os.Stderr, "Unable to match: pattern, got error %v", err)
									return nil
								}
								if match {
									allMatch = false
								}
							}
							for _, pattern := range currentPatterns {
								match, err := regexp.MatchString(pattern, importSpec.Path)
								if err != nil {
									fmt.Fprintf(os.Stderr, "Unable to match: pattern, got error %v", err)
									return nil
								}
								if match {
									allMatch = false
								}
							}
							if allMatch {
								lintErrors = addLintError(importSpec, lintErrors, errImportMatchedAlready, seen)
								return findRecErr(importGroup[1:], currentPatterns[0:], lintErrors, seenPatterns)
							}
						}
					default:
						match, err := regexp.MatchString(seen, importSpec.Path)
						if err != nil {
							fmt.Fprintf(os.Stderr, "Unable to match: pattern, got error %v", err)
							return nil
						}
						if match {
							lintErrors = addLintError(importSpec, lintErrors, errImportMatchedAlready, seen)
							return findRecErr(importGroup[1:], currentPatterns[0:], lintErrors, seenPatterns)
						}
					}
				}
				match, err := regexp.MatchString(currentPatterns[0], importSpec.Path)
				if err != nil {
					fmt.Fprintf(os.Stderr, "Unable to match: pattern, got error %v", err)
					return nil
				}
				if match {
					if patternSeen(currentPatterns[0], seenPatterns) {
						lintErrors = addLintError(importSpec, lintErrors, errGroupMachedAlready, currentPatterns[0])
						return findRecErr(importGroup[1:], currentPatterns[1:], lintErrors, seenPatterns)
					}
					checking = true
				}
				if !match && checking {
					lintErrors = addLintError(importSpec, lintErrors, errOutOfOrder, currentPatterns[0])
				}
				if !match && !checking {
					return findRecErr(importGroup[0:], currentPatterns[1:], lintErrors, seenPatterns)
				}
				if i == len(group.Imports)-1 {
					seenPatterns = addSeenPattern(currentPatterns[0], seenPatterns)
					return findRecErr(importGroup[1:], currentPatterns[1:], lintErrors, seenPatterns)
				}
			}
		}
	}
	return lintErrors
}

func patternSeen(pattern string, seenPatterns []string) bool {
	for _, p := range seenPatterns {
		if p == pattern {
			return true
		}
	}
	return false
}

func addSeenPattern(pattern string, seenPatterns []string) []string {
	for _, p := range seenPatterns {
		if p == pattern {
			return seenPatterns
		}
	}
	seenPatterns = append(seenPatterns, pattern)
	return seenPatterns
}

func isThirdParty(path string) bool {
	re := regexp.MustCompile(`\.`)
	return re.MatchString(path)
}

func addLintError(i importSpec, lintErrors lintErrors, err error, pattern string) lintErrors {
	lintErrors = append(lintErrors, lintError{
		fileName:    i.Name,
		importName:  i.Path,
		line:        i.Line,
		patternSeen: pattern,
		err:         err,
	})
	return lintErrors
}
