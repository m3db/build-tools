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
	"fmt"
	"go/build"
	"io/ioutil"
	"os"
	"strings"

	"github.com/m3db/build-tools/utilities/ggd/lib"

	"github.com/pborman/getopt"
	"github.com/tmc/dot"
)

const usageNotes = `ggd: command line tool to find packages affected by git changes. Examples:
# Assuming CWD is in a git repository directory present in the GOPATH.

# (1) List all the golang packages affected by changes between master and HEAD
ggd

# (2) List all the golang packages affected by changes between branchA and HEAD
ggd branchA

# (3) List all the golang packages affected by between changes between branchA and branchB
ggd branchA..branchB

# (4) Run tests for all the golang packages affected by between master, and head
go test $(ggd)

# (5) List all the golang packages affected by between master and head, and build tags 'integration'
ggd -t integration

# (6) The same as (5), but include debug output (sent to STDERR)
ggd -t integration -d

# (7) The same as (6), but include debug output (sent to STDERR), and
# save the generated DAG in changes.png for visualization
ggd -t integration -d -o change.dot
dot -Tpng change.png change.dot
`

var (
	basePkg          = getopt.StringLong("base-package", 'p', "", "repository package (can leave blank typically)")
	debugMode        = getopt.BoolLong("debug", 'd', "debug mode - useful for understanding output")
	buildTags        = getopt.ListLong("build-tags", 't', "comma separated golang build tags")
	catchallPatterns = getopt.ListLong("catchall-patterns", 'c', "comma separated file/dir names, such that if any of these are changed, all tests are run")
	filterPatterns   = getopt.ListLong("filter-patterns", 'f', "comma separated dir names to be filtered out during change calculation")
	dotOutputFile    = getopt.StringLong("output-file", 'o', "", "optionally, visualize changes in a dot file (only in debug mode)")
)

var (
	defaultCatchallPatterns = []string{
		"/testdata/", // go list ./... is unable to gauge impact of changes to this directory
	}
	defaultFilterPatterns = []string{"/vendor/"}
)

func main() {
	getopt.Parse()
	if err := validateArgs(); err != nil {
		printUsage()
		die("illegal arguments: %v", err)
	}
	sha1Fn := lib.CurrentSHA1

	baseRef, headRef, err := rangeExtractor{sha1Fn}.extractRanges(getopt.Args())
	dieIfErr(err, "unable to extract ranges: %v", err)
	commitRange := fmt.Sprintf("%s..%s", baseRef, headRef)
	debug("retrieving files changed from git in commit-range: %v", commitRange)

	gitChanges, err := lib.ChangedFiles(commitRange, *basePkg)
	dieIfErr(err, "unable to find changed files from git: %v", err)

	debug("files changed from git: %v", gitChanges)
	debug("calculating changed packages")

	debug("default catchall patterns %v", defaultCatchallPatterns)
	catchall := defaultCatchallPatterns
	if catchallPatterns != nil && len(*catchallPatterns) != 0 {
		debug("overriding catchall patterns to %v", *catchallPatterns)
		catchall = *catchallPatterns
	}

	allPackagesAffected := anyMatches(gitChanges, catchall)
	debug("allPackagesAffected %v", allPackagesAffected)

	changedPackages := []string{}
	if !allPackagesAffected {
		changedPackages = lib.ChangedPackages(gitChanges)
		debug("changed packages: %v", changedPackages)
	} else {
		debug("skipping changed packages calculation, know all are affected")
	}

	debug("default filter patterns %v", defaultFilterPatterns)
	filters := defaultFilterPatterns
	if filterPatterns != nil && len(*filterPatterns) != 0 {
		debug("overriding filter patterns to %v", *filterPatterns)
		filters = *filterPatterns
	}

	currentSHA1, err := sha1Fn()
	dieIfErr(err, "unable to extract current sha1: %v", err)
	if currentSHA1 != headRef {
		debug("checking out head-ref: %v", headRef)
		checkoutResult, err := lib.Checkout(headRef)
		dieIfErr(err, "unable to checkout head-ref: %v", err)
		debug("checked out head-ref: %v", headRef)
		debug("git-output: %v", checkoutResult)
	}

	debug("calculating affected packages on head-ref due to changes")
	changeSet, graph, fullChangeSet, err := dagHelper(changedPackages, allPackagesAffected, filters, *basePkg)
	dieIfErr(err, "unable to compute dag: %v", err)
	debug("affected packages (including transitive changes): %v", fullChangeSet)
	debug("change graph: %v", graph)

	// output changedset
	for c := range fullChangeSet {
		fmt.Println(c)
	}

	if *dotOutputFile == "" {
		debug("skipping creation of dot output file")
		return
	}

	debug("creating dot output file")
	dotOutput := computeDot(changeSet, fullChangeSet, graph, *basePkg)
	err = ioutil.WriteFile(*dotOutputFile, []byte(dotOutput), 0644)
	dieIfErr(err, "unable to write output file: %v", err)
	debug("created dot output file at %v", dotOutputFile)
}

type currentSHA1Fn func() (string, error)

type rangeExtractor struct {
	sha1Fn currentSHA1Fn
}

func (r rangeExtractor) extractRanges(args []string) (base string, head string, err error) {
	currentSHA1, err := r.sha1Fn()
	if err != nil {
		return "", "", fmt.Errorf("unable to retrieve current sha1: %v", err)
	}
	if len(args) == 0 {
		return "master", currentSHA1, nil
	}
	const rangeParam = ".."
	arg0 := args[0]
	if strings.Contains(arg0, rangeParam) {
		tokens := strings.Split(arg0, rangeParam)
		if len(tokens) != 2 {
			return "", "", fmt.Errorf("illegal commit range provided: %s", arg0)
		}
		if strings.ToLower(tokens[1]) == "head" {
			return tokens[0], currentSHA1, nil
		}
		return tokens[0], tokens[1], nil
	}
	return arg0, currentSHA1, nil
}

func computeDot(changedSet, fullChangeSet lib.ImportSet, graph lib.ImportGraph, basePkg string) string {
	dGraph := dot.NewGraph("Changes")
	dGraph.Set("rankdir", "LR")
	dGraph.SetType(dot.DIGRAPH)
	nodeMap := map[string]*dot.Node{}
	createNode := func(p string) *dot.Node {
		if n, ok := nodeMap[p]; ok {
			return n
		}
		n := dot.NewNode(withoutPkgName(p, basePkg))
		// change shape for directories differently to visualize ordering more clearly
		if _, ok := changedSet[p]; ok {
			n.Set("shape", "star")
		}
		dGraph.AddNode(n)
		nodeMap[p] = n
		return n
	}
	for p := range fullChangeSet {
		for q := range graph[p] {
			dGraph.AddEdge(dot.NewEdge(createNode(p), createNode(q)))
		}
	}
	return dGraph.String()
}

func dagHelper(
	changedPackages []string,
	allPackagesChanged bool,
	filterPatterns []string,
	basePkg string,
) (
	changedSet lib.ImportSet,
	g lib.ImportGraph,
	fullChangeSet lib.ImportSet,
	err error,
) {
	buildCtx := &build.Default
	if buildTags != nil {
		buildCtx.BuildTags = *buildTags
	}
	g, err = lib.NewImportGraph(buildCtx, basePkg)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("unable to find import dag: %v", err)
	}
	debug("import graph: %v", g)

	allPackagesFn := func() {
		allVerts := g.Vertices()
		debug("graphVertices: %v", g)
		debug("applying filter patterns", filterPatterns)
		for vert := range allVerts {
			if lib.MatchesAny(vert, filterPatterns) {
				debug("%v matched filter pattern, dropping", vert)
				delete(allVerts, vert)
			}
		}
		changedSet = allVerts
		fullChangeSet = allVerts
	}

	if allPackagesChanged {
		allPackagesFn()
		return changedSet, g, fullChangeSet, nil
	}

	changedSet = make(lib.ImportSet)
	fullChangeSet = make(lib.ImportSet)
	changedPackageNames := make([]string, 0, len(changedPackages))
	for _, p := range changedPackages {
		if lib.MatchesAny(p, filterPatterns) {
			debug("%v matched filter pattern, dropping", p)
			continue
		}
		changedSet[p] = struct{}{}
		fullChangeSet[p] = struct{}{}
		changedPackageNames = append(changedPackageNames, p)
	}

	closure, err := g.Closure(changedPackageNames...)
	if err != nil {
		debug("unable to compute closure: %v", err)
		debug("defaulting to all packages")
		allPackagesFn()
		return changedSet, g, fullChangeSet, nil
	}
	debug("change closure: %v", g)

	for p := range closure {
		fullChangeSet[p] = struct{}{}
	}

	return changedSet, g, fullChangeSet, nil
}

func validateArgs() error {
	if *basePkg == "" {
		pkg, err := inferPackage()
		if err != nil {
			return fmt.Errorf("unable to infer package (%v), please use -base-pkg", err)
		}
		*basePkg = pkg
		debug("inferred base package as: %v\n", *basePkg)
	}
	if !*debugMode && *dotOutputFile != "" {
		return fmt.Errorf("can only create dot output file in debug mode")
	}
	return nil
}

func withoutPkgName(p string, basePkg string) string {
	return strings.TrimPrefix(p, basePkg+"/")
}

func inferPackage() (string, error) {
	gopath := os.Getenv("GOPATH")
	if gopath == "" {
		return "", fmt.Errorf("GOPATH is not set")
	}

	d, err := os.Getwd()
	if err != nil {
		return "", err
	}

	path_prefix := fmt.Sprintf("%s/src/", gopath)
	if !strings.HasPrefix(d, path_prefix) {
		return "", fmt.Errorf("unable to infer path [%s, %s]", path_prefix, d)
	}

	currentPkg := strings.TrimPrefix(d, path_prefix)
	return currentPkg, nil
}

func anyMatches(inputs []string, patterns []string) bool {
	for _, in := range inputs {
		if lib.MatchesAny(in, patterns) {
			return true
		}
	}
	return false
}

func printUsage() {
	getopt.PrintUsage(os.Stderr)
	fmt.Println(usageNotes)
}

func dieIfErr(err error, f string, args ...interface{}) {
	dieIf(err != nil, f, args)
}

func dieIf(c bool, f string, args ...interface{}) {
	if c {
		die(f, args)
	}
}

func die(f string, args ...interface{}) {
	printerHelper(f, args)
	os.Exit(1)
}

func debug(f string, args ...interface{}) {
	if *debugMode {
		printerHelper(f, args)
	}
}

func printerHelper(f string, args ...interface{}) {
	if len(args) == 0 {
		fmt.Fprintf(os.Stderr, f)
	} else {
		fmt.Fprintf(os.Stderr, f, args)
	}
	fmt.Fprintf(os.Stderr, "\n")
}
