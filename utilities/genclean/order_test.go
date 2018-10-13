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
	"bytes"
	"fmt"
	"io/ioutil"
	"os"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

type testCase struct {
	input  string
	output string
}

func filePath(n int, suffix string) string {
	return fmt.Sprintf("testdata/order/example_%d.go.%s", n, suffix)
}

func newTestCase(n int) testCase {
	return testCase{
		input:  filePath(n, "input"),
		output: filePath(n, "output"),
	}
}

func TestOrderExample(t *testing.T) {
	numTest := 6
	testCases := []testCase{}
	for i := 1; i <= numTest; i++ {
		testCases = append(testCases, newTestCase(i))
	}
	for _, tc := range testCases {
		inputBytes, err := ioutil.ReadFile(tc.input)
		require.NoError(t, err, "", tc)

		expFile, err := os.Open(tc.output)
		require.NoError(t, err)
		expected, err := ioutil.ReadAll(expFile)
		require.NoError(t, err, "", tc)

		obs, err := reorderImports(inputBytes, strings.Fields(defaultGroupPrefixes))
		require.NoError(t, err)

		if !bytes.Equal(expected, obs) {
			e := string(expected)
			o := string(obs)
			require.Fail(t, fmt.Sprintf("[%v] did not match: \ne:[%v]\no:[%v]", tc, e, o))
		}
	}
}
