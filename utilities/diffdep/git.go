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
	"log"
	"path/filepath"

	"gopkg.in/src-d/go-git.v4"
	"gopkg.in/src-d/go-git.v4/plumbing"
	"gopkg.in/src-d/go-git.v4/plumbing/object"
)

func findChangedDirs(ctx *build.Context, rootDir, baseBranch string) (map[string]struct{}, error) {
	path := filepath.Join(ctx.GOPATH, "src", rootDir)
	repo, err := git.PlainOpen(path)
	if err != nil {
		return nil, fmt.Errorf("unable to open git repository in '%v': %v", path, err)
	}

	headTree, err := getTree(repo, plumbing.HEAD.Short())
	if err != nil {
		return nil, err
	}

	baseTree, err := getTree(repo, baseBranch)
	if err != nil {
		return nil, err
	}

	diff, err := baseTree.Diff(headTree)
	if err != nil {
		return nil, fmt.Errorf("could not calculate diff between HEAD and branch '%v': %v", baseBranch, err)
	}

	changed := make(map[string]struct{})
	for _, change := range diff {
		maybeAddDir(changed, rootDir, change.From.Name)
		maybeAddDir(changed, rootDir, change.To.Name)
	}

	return changed, nil
}

func getTree(repo *git.Repository, branch string) (*object.Tree, error) {
	hash, err := repo.ResolveRevision(plumbing.Revision(branch))
	if err != nil {
		log.Fatalf("could not get SHA for branch '%v': %v", branch, err)
	}

	obj, err := repo.Object(plumbing.AnyObject, *hash)
	if err != nil {
		return nil, fmt.Errorf("could not get git object for branch '%v': %v", branch, err)
	}

	commit, ok := obj.(*object.Commit)
	if !ok {
		return nil, fmt.Errorf("unexpected typo of object, found %T, want *object.Commit", obj)
	}

	tree, err := commit.Tree()
	if err != nil {
		return nil, fmt.Errorf("could not get tree for branch '%v': %v", branch, err)
	}

	return tree, nil
}

func maybeAddDir(dirs map[string]struct{}, rootDir, file string) {
	// Get the directory of the file by prepending the root directory, removing the
	// filename, and stripping the trailing slash.
	dir, _ := filepath.Split(filepath.Join(rootDir, file))
	dirs[filepath.Clean(dir)] = struct{}{}
}
