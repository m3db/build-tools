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
	"path/filepath"
	"strings"
)

// ChangedPackages returns all changed packages.
func ChangedPackages(changedFiles []string) []string {
	changedPkgs := make(map[string]struct{}, len(changedFiles))
	for _, f := range changedFiles {
		changedPkgs[filepath.Dir(f)] = struct{}{}
	}
	pkgs := make([]string, 0, len(changedPkgs))
	for p := range changedPkgs {
		pkgs = append(pkgs, p)
	}
	return pkgs
}

// filterChanges filters changed files based on the following heuristic,
// only files meeting any of the following conditions affect go compilation
// 	- any .go file
//  - any directory called `testdata/`
//
// also filters any `vendor/` directory paths
func filterChanges(input []string) []string {
	out := make([]string, 0, len(input))
	for _, i := range input {
		if strings.Contains(i, "/vendor/") {
			continue
		} else if strings.HasSuffix(i, ".go") || strings.Contains(i, "/testdata/") {
			out = append(out, i)
		}
	}
	return out
}
