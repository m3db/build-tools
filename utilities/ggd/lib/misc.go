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
	changedPkgsMap := make(map[string]struct{}, len(changedFiles))
	for _, f := range changedFiles {
		dir := filepath.Dir(f)
		changedPkgsMap[dir] = struct{}{}
	}
	changedPackages := make([]string, 0, len(changedPkgsMap))
	for p := range changedPkgsMap {
		changedPackages = append(changedPackages, p)
	}
	return changedPackages
}

// MatchesAny returns if input matches any of the patterns specified.
func MatchesAny(input string, patterns []string) bool {
	for _, ptrn := range patterns {
		if strings.Contains(input, ptrn) {
			return true
		}
	}
	return false
}
