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
	"testing"

	"github.com/stretchr/testify/require"
)

func TestSha1Extraction(t *testing.T) {
	for _, tc := range []struct {
		name         string
		sha1Fn       currentSha1Fn
		args         []string
		expectedBase string
		expectedHead string
		expectedErr  bool
	}{
		{"case0", testSha1Fn(), nil, "master", "TESTSHA1", false},
		{"case1", failingTestSha1Fn(), nil, "", "", true},
		{"case2", testSha1Fn(), []string{"master"}, "master", "TESTSHA1", false},
		{"case3", failingTestSha1Fn(), []string{"master"}, "", "", true},
		{"case4", testSha1Fn(), []string{"master..HEAD"}, "master", "TESTSHA1", false},
		{"case5", failingTestSha1Fn(), []string{"master..HEAD"}, "", "", true},
		{"case6", testSha1Fn(), []string{"xyz..HEAD"}, "xyz", "TESTSHA1", false},
		{"case7", testSha1Fn(), []string{"xyz..DDD"}, "xyz", "DDD", false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			obsBase, obsHead, obsErr := rangeExtractor{tc.sha1Fn}.extractRanges(tc.args)
			require.Equal(t, tc.expectedHead, obsHead)
			require.Equal(t, tc.expectedBase, obsBase)
			if !tc.expectedErr {
				require.NoError(t, obsErr)
			} else {
				require.Error(t, obsErr)
			}
		})
	}
}

func testSha1Fn() currentSha1Fn {
	return func() (string, error) {
		return "TESTSHA1", nil
	}
}

func failingTestSha1Fn() currentSha1Fn {
	return func() (string, error) {
		return "", fmt.Errorf("random error")
	}
}
