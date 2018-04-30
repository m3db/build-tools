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

package lib

import (
	"bytes"
	"fmt"
	"go/build"

	"golang.org/x/tools/go/buildutil"
)

// DanglingDeleteError indicates an otherwise un-imported package was deleted
// from the repository.
type DanglingDeleteError struct {
	pkgName string
}

func (d DanglingDeleteError) Error() string {
	return fmt.Sprintf("dangling package deleted: %v", d.pkgName)
}

// ImportSet is a set of imports.
type ImportSet map[string]struct{}

// ImportGraph is a map from a Go package, `p` to the set of all packages which import `p`.
type ImportGraph map[string]ImportSet

// String representation of the Import Graph.
func (g ImportGraph) String() string {
	var buffer bytes.Buffer
	for pkg, importers := range g {
		for importer := range importers {
			buffer.WriteString(fmt.Sprintf("%s imported by %s\n", pkg, importer))
		}
	}
	return buffer.String()
}

func NewImportGraph(ctx *build.Context, basePkg string) (ImportGraph, error) {
	pkgs := buildutil.ExpandPatterns(ctx, []string{basePkg + "/..."})
	g := make(ImportGraph, len(pkgs))

	for pkg := range pkgs {
		buildPkg, err := ctx.Import(pkg, "", 0)
		if err != nil {
			if _, ok := err.(*build.NoGoError); ok {
				// The package doesn't contain any buildable go files. This can be caused, for
				// example, when the only go files in a package contain build flags which are
				// not present in our build context.
				continue
			}
			return nil, fmt.Errorf("could not import package %v: %v", pkg, err)
		}

		visited := make(map[string]string)

		// ensure we add the pkg to the graph
		if _, ok := g[pkg]; !ok {
			g[pkg] = make(ImportSet)
		}

		for _, path := range buildPkg.Imports {
			g.maybeAddEdge(ctx, buildPkg, visited, pkg, path)
		}
		for _, path := range buildPkg.TestImports {
			g.maybeAddEdge(ctx, buildPkg, visited, pkg, path)
		}
		for _, path := range buildPkg.XTestImports {
			g.maybeAddEdge(ctx, buildPkg, visited, pkg, path)
		}
	}

	return g, nil
}

func (g ImportGraph) maybeAddEdge(
	ctx *build.Context, buildPkg *build.Package, visited map[string]string, pkg, path string,
) {
	if path == "C" {
		// Not a real package.
		return
	}

	importedPkg, ok := visited[path]
	if !ok {
		// It's okay for Import to return an error as not all packages that can be found in
		// a package will necessarily be present. For example, packages imported only by test
		// files in vendored packages will not be installed. In the case of an error, Import
		// always returns a non-nil *Package that contains partial information.
		importBuildPkg, _ := ctx.Import(path, buildPkg.Dir, build.FindOnly)
		if importBuildPkg != nil {
			importedPkg = importBuildPkg.ImportPath
		} else {
			importedPkg = path
		}
		visited[path] = importedPkg
		g.addEdge(importedPkg, pkg)
	}
}

func (g ImportGraph) addEdge(from, to string) {
	edges, ok := g[from]
	if !ok {
		edges = make(map[string]struct{})
		g[from] = edges
	}
	edges[to] = struct{}{}
}

// Vertices returns all known vertices in the graph.
func (g ImportGraph) Vertices() ImportSet {
	visited := make(ImportSet)
	for node, impSet := range g {
		visited[node] = struct{}{}
		for imp := range impSet {
			visited[imp] = struct{}{}
		}
	}
	return visited
}

// Closure return the transitive closure of all packages reachable
// by starting at the provided paths in the ImportGraph.
func (g ImportGraph) Closure(paths ...string) (ImportSet, error) {
	// recursive walk
	closure := make(ImportSet)
	for _, p := range paths {
		if err := g.walk(p, closure); err != nil {
			return nil, err
		}
	}

	// done
	return closure, nil
}

func (g ImportGraph) walk(node string, visited ImportSet) error {
	if _, ok := visited[node]; ok {
		return nil
	}
	if _, ok := g[node]; !ok {
		// NB(prateek): this happens in the case of un-used deleted packages. Look at testcase7.
		return DanglingDeleteError{node}
	}
	visited[node] = struct{}{}
	for to := range g[node] {
		g.walk(to, visited)
	}
	return nil
}
