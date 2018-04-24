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
)

type graph map[string]map[string]struct{}

func newGraph(ctx *build.Context, pkgs map[string]bool) (graph, error) {
	g := make(graph, len(pkgs))

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

func (g graph) maybeAddEdge(
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
		// always returns a non-nil *Package. In the case of an error it will only contain
		// partial information.
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

func (g graph) addEdge(from, to string) {
	edges, ok := g[from]
	if !ok {
		edges = make(map[string]struct{})
		g[from] = edges
	}
	edges[to] = struct{}{}
}

func (g graph) walk(node string, visited map[string]struct{}) error {
	if _, ok := visited[node]; ok {
		return nil
	}
	visited[node] = struct{}{}
	edges, ok := g[node]
	if !ok {
		return fmt.Errorf("could not find node '%s' in graph", node)
	}
	for to := range edges {
		if err := g.walk(to, visited); err != nil {
			return err
		}
	}
	return nil
}
