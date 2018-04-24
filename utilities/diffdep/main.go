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
	"flag"
	"fmt"
	"go/build"
	"log"
	"strings"

	"golang.org/x/tools/go/buildutil"
)

func main() {
	var (
		branch string
	)
	flag.StringVar(&branch, "branch", "master", "the base branch to compare against to find changed files")
	flag.Parse()

	if len(flag.Args()) != 1 {
		log.Fatal("TODO")
	}
	repo := flag.Args()[0]

	ctx := build.Default

	changedDirs, err := findChangedDirs(&ctx, repo, branch)
	if err != nil {
		log.Fatalf("unable to find changed directories: %v", err)
	}
	if len(changedDirs) == 0 {
		return
	}

	pkgs := buildutil.ExpandPatterns(&ctx, []string{repo + "/..."})
	graph, err := newGraph(&ctx, pkgs)
	if err != nil {
		log.Fatalf("unable to build dependency graph: %v", err)
	}

	seen := make(map[string]struct{})
	for dir := range changedDirs {
		if _, ok := pkgs[dir]; !ok {
			continue
		}
		graph.walk(dir, seen)
	}

	for pkg := range seen {
		if !strings.Contains(pkg, "/vendor/") {
			fmt.Println(pkg)
		}
	}
}
