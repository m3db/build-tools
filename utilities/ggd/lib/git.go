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
	"fmt"
	"os/exec"
	"path/filepath"
	"strings"
)

// ChangedFiles returns the files changed in the provided commit range.
func ChangedFiles(commitRange string, basePkg string) ([]string, error) {
	cmd := exec.Command("git", "diff-tree", "--no-commit-id", "--name-only", "-r", commitRange)
	dat, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("Could not run git diff-tree: %v", err)
	}
	files := strings.Split(string(dat), "\n")
	var res []string
	for _, f := range files {
		f = strings.TrimSpace(f)
		if len(f) == 0 {
			continue
		}
		res = append(res, filepath.Join(basePkg, f))
	}
	return res, nil
}

// CWDIsDirty returns whether the current directory is dirty.
func CWDIsDirty() (bool, error) {
	cmd := exec.Command("git", "status", "--porcelain")
	dat, err := cmd.Output()
	if err != nil {
		return true, fmt.Errorf("Could not run git status: %v", err)
	}
	return "" != strings.TrimSpace(string(dat)), nil
}

// CurrentSHA1 returns the SHA1 of HEAD.
func CurrentSHA1() (string, error) {
	cmd := exec.Command("git", "rev-parse", "--verify", "HEAD")
	dat, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("Could not run git rev-parse: %v", err)
	}
	return strings.TrimSpace(string(dat)), nil
}

// Checkout the specified git sha1.
func Checkout(ref string) (string, error) {
	cmd := exec.Command("git", "checkout", ref)
	dat, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("Could not run git checkout: %v", err)
	}
	return strings.TrimSpace(string(dat)), nil
}
