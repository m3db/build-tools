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
	"io/ioutil"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func renameFilePath(n int, suffix string) string {
	return fmt.Sprintf("testdata/renaming/example_%d.go.%s", n, suffix)
}

func newRenamingTestCase(n int) testCase {
	return testCase{
		input:  renameFilePath(n, "input"),
		output: renameFilePath(n, "output"),
	}
}

func TestRenamingExamples(t *testing.T) {
	testCases := []testCase{
		newRenamingTestCase(1),
		newRenamingTestCase(2),
	}
	for _, tc := range testCases {
		bytes, err := ioutil.ReadFile(tc.input)
		require.NoError(t, err, "", tc)

		expected, err := ioutil.ReadFile(tc.output)
		require.NoError(t, err, "", tc)

		obs, err := cleanupImports(bytes)
		require.NoError(t, err)

		e := strings.Trim(string(expected), " \t\n")
		o := strings.Trim(string(obs), " \n\t")
		require.True(t, e == o,
			fmt.Sprintf("expected: [%s]\n observed: [%s]\n", e, o), tc)
	}
}
