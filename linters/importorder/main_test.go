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
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

func getFilename(path string) string {
	filename, err := filepath.Abs(path)
	if err != nil {
		panic(err)
	}
	return filename
}

func TestImportLinter(t *testing.T) {
	groupedIntErrors := handleImportPaths(
		[]string{"./testdata/normal_order/"},
		[]string{"integration"},
		[]string{"STDLIB", "github.com/m3db/m3coordinator", "github.com/m3db", "EXTERNAL"},
	)

	require.Equal(t, errOutOfOrder, groupedIntErrors[0].err)     // test_file_1.go
	require.Equal(t, errOutOfOrder, groupedIntErrors[1].err)     // test_file_10.go
	require.Equal(t, errMultipleImport, groupedIntErrors[2].err) // test_file_2.go
	require.Equal(t, errOutOfOrder, groupedIntErrors[3].err)     // test_file_3.go
	require.Equal(t, errOutOfOrder, groupedIntErrors[4].err)     // test_file_4.go
	require.Equal(t, errOutOfOrder, groupedIntErrors[5].err)     // test_file_6.go
	require.Equal(t, errOutOfOrder, groupedIntErrors[6].err)     // test_file_8.go

	groupedExtErrors := handleImportPaths(
		[]string{"./testdata/ext_order/"},
		[]string{"included"},
		[]string{"STDLIB", "EXTERNAL", "github.com/m3db/m3coordinator", "github.com/m3db"},
	)

	require.Equal(t, errOutOfOrder, groupedExtErrors[1].err)
	require.Equal(t, errOutOfOrder, groupedExtErrors[2].err)
	require.Equal(t, errOutOfOrder, groupedExtErrors[3].err)
	require.Equal(t, errOutOfOrder, groupedExtErrors[4].err)
	require.Equal(t, errOutOfOrder, groupedExtErrors[5].err)
	require.Equal(t, errOutOfOrder, groupedExtErrors[6].err)

	groupedNoExtErrors := handleImportPaths(
		[]string{"./testdata/no_ext_order/"},
		[]string{"included"},
		[]string{"STDLIB", "github.com/m3db/m3coordinator", "github.com/m3db"},
	)

	require.Equal(t, errOutOfOrder, groupedNoExtErrors[0].err)
	require.Equal(t, errOutOfOrder, groupedNoExtErrors[1].err)
}
