// Copyright (c) 2017 Uber Technologies, Inc.
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
	expectedInternalLintErrors := lintErrors{
		lintError{
			fileName:    getFilename("./testdata/normal_order/test_file_1.go"),
			importName:  "\"github.com/m3db/m3coordinator/models\"",
			line:        39,
			patternSeen: "github.com/m3db/m3coordinator",
			err:         errImportMatchedAlready,
		},
		lintError{
			fileName: getFilename("./testdata/normal_order/test_file_2.go"),
			err:      errMultipleImport,
		},
		lintError{
			fileName:    getFilename("./testdata/normal_order/test_file_3.go"),
			importName:  "\"github.com/m3db/m3db/digest\"",
			line:        36,
			patternSeen: "github.com/m3db",
			err:         errImportMatchedAlready,
		},
		lintError{
			fileName:    getFilename("./testdata/normal_order/test_file_4.go"),
			importName:  "\"time\"",
			line:        26,
			patternSeen: standardImportGroup,
			err:         errImportMatchedAlready,
		},
		lintError{
			fileName:   getFilename("./testdata/normal_order/test_file_4.go"),
			importName: "\"gopkg.in/alecthomas/kingpin.v2\"",
			line:       30,
			err:        errNoMatch,
		},
		lintError{
			fileName:    getFilename("./testdata/normal_order/test_file_6.go"),
			importName:  "\"go.uber.org/zap\"",
			line:        25,
			patternSeen: "github.com/m3db/m3coordinator",
			err:         errOutOfOrder,
		},
		lintError{
			fileName:    getFilename("./testdata/normal_order/test_file_6.go"),
			importName:  "\"gopkg.in/alecthomas/kingpin.v2\"",
			line:        26,
			patternSeen: "github.com/m3db/m3coordinator",
			err:         errOutOfOrder,
		},
		lintError{
			fileName:    getFilename("./testdata/normal_order/test_file_6.go"),
			importName:  "\"github.com/m3db/m3coordinator/services/m3coordinator/config\"",
			line:        28,
			patternSeen: "github.com/m3db/m3coordinator",
			err:         errImportMatchedAlready,
		},
		lintError{
			fileName:    getFilename("./testdata/normal_order/test_file_6.go"),
			importName:  "\"context\"",
			line:        35,
			patternSeen: standardImportGroup,
			err:         errOutOfOrder,
		},
		lintError{
			fileName:   getFilename("./testdata/normal_order/test_file_6.go"),
			importName: "\"context\"",
			line:       35,
			err:        errNoMatch,
		},
		lintError{
			fileName:    getFilename("./testdata/normal_order/test_file_8.go"),
			importName:  "\"go.uber.org/zap\"",
			line:        30,
			patternSeen: "github.com/m3db/m3coordinator",
			err:         errOutOfOrder,
		},
	}

	expectedExternalLintErrors := lintErrors{
		lintError{
			fileName:    getFilename("./testdata/ext_order/test_file_2.go"),
			importName:  "\"github.com/m3db/m3coordinator/services/m3coordinator/config\"",
			line:        32,
			patternSeen: "github.com/m3db/m3coordinator",
			err:         errImportMatchedAlready,
		},
		lintError{
			fileName:    getFilename("./testdata/ext_order/test_file_3.go"),
			importName:  "\"go.uber.org/zap\"",
			line:        25,
			patternSeen: "github.com/m3db/m3coordinator",
			err:         errOutOfOrder,
		},
		lintError{
			fileName:    getFilename("./testdata/ext_order/test_file_3.go"),
			importName:  "\"github.com/m3db/m3coordinator/services/m3coordinator/config\"",
			line:        27,
			patternSeen: "github.com/m3db/m3coordinator",
			err:         errImportMatchedAlready,
		},
		lintError{
			fileName:    getFilename("./testdata/ext_order/test_file_4.go"),
			importName:  "\"github.com/m3db/m3coordinator/services/m3coordinator/config\"",
			line:        31,
			patternSeen: "github.com/m3db/m3coordinator",
			err:         errImportMatchedAlready,
		},
		lintError{
			fileName:    getFilename("./testdata/ext_order/test_file_4.go"),
			importName:  "\"github.com/alecthomas/template\"",
			line:        35,
			patternSeen: externalImportGroup,
			err:         errImportMatchedAlready,
		},
		lintError{
			fileName:   getFilename("./testdata/ext_order/test_file_5.go"),
			importName: "\"fmt\"",
			line:       33,
			err:        errNoMatch,
		},
		lintError{
			fileName:   getFilename("./testdata/ext_order/test_file_6.go"),
			importName: "\"time\"",
			line:       35,
			err:        errNoMatch,
		},
		lintError{
			fileName:    getFilename("./testdata/ext_order/test_file_7.go"),
			importName:  "\"github.com/m3db/m3db/client\"",
			line:        28,
			patternSeen: "github.com/m3db/m3coordinator",
			err:         errOutOfOrder,
		},
		lintError{
			fileName:    getFilename("./testdata/ext_order/test_file_7.go"),
			importName:  "\"github.com/m3db/m3coordinator/services/m3coordinator/httpd\"",
			line:        30,
			patternSeen: "github.com/m3db/m3coordinator",
			err:         errImportMatchedAlready,
		},
	}

	groupedIntErrors := handleImportPaths(
		[]string{"./testdata/normal_order/"},
		[]string{"integration"},
		[]string{"STDLIB", "github.com/m3db/m3coordinator", "github.com/m3db", "EXTERNAL"},
	)

	require.Equal(t, expectedInternalLintErrors, groupedIntErrors)

	groupedExtErrors := handleImportPaths(
		[]string{"./testdata/ext_order/"},
		[]string{"included"},
		[]string{"STDLIB", "EXTERNAL", "github.com/m3db/m3coordinator", "github.com/m3db"},
	)

	require.Equal(t, expectedExternalLintErrors, groupedExtErrors)
}
