package main

import (
	"fmt"
	"go/token"
	"path"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestTimeMapLint(t *testing.T) {
	type lintError struct {
		lineNumber int
		keyStr     string
		valStr     string
	}

	expectedLintErrors := map[string][]lintError{
		"test_file_1.go": []lintError{
			lintError{
				lineNumber: 18,
				keyStr:     "time.Time",
				valStr:     "bool",
			},
			lintError{
				lineNumber: 19,
				keyStr:     "time.Time",
				valStr:     "bool",
			},
		},
		"test_file_2.go": []lintError{
			lintError{
				lineNumber: 7,
				keyStr:     "maptime.timeAlias",
				valStr:     "bool",
			},
			lintError{
				lineNumber: 8,
				keyStr:     "maptime.timeAlias",
				valStr:     "bool",
			},
		},
		"test_file_3.go": []lintError{
			lintError{
				lineNumber: 5,
				keyStr:     "time.Time",
				valStr:     "bool",
			},
		},
		"test_file_4.go": []lintError{
			lintError{
				lineNumber: 4,
				keyStr:     "maptime.timeAlias",
				valStr:     "bool",
			},
		},
		"test_file_5.go": []lintError{
			lintError{
				lineNumber: 9,
				keyStr:     "maptime.structWithInnerTime",
				valStr:     "bool",
			},
			lintError{
				lineNumber: 10,
				keyStr:     "maptime.structWithInnerTime",
				valStr:     "bool",
			},
		},
		"test_file_6.go": []lintError{
			lintError{
				lineNumber: 7,
				keyStr:     "maptime.chanTime",
				valStr:     "bool",
			},
			lintError{
				lineNumber: 8,
				keyStr:     "maptime.chanTime",
				valStr:     "bool",
			},
		},
	}

	observedLintErrors := map[string][]lintError{}
	testCallback := func(position token.Position, keyStr, valStr string) {
		filePath := position.Filename
		filePathBase := path.Base(filePath)
		_, ok := expectedLintErrors[filePathBase]
		require.True(t, ok, fmt.Sprintf("Failed for file: %s", filePathBase))
		observedLintErrorsForFile, _ := observedLintErrors[filePathBase]
		observedLintErrors[filePathBase] = append(
			observedLintErrorsForFile,
			lintError{
				lineNumber: position.Line,
				keyStr:     keyStr,
				valStr:     valStr,
			},
		)
	}
	handleImportPaths([]string{"."}, testCallback)

	// Make sure all observed errors were expected
	for file, observedErrs := range observedLintErrors {
		expectedErrs, ok := expectedLintErrors[file]
		require.True(t, ok, fmt.Sprintf("Failed for file: %s", file))
		require.Equal(t, expectedErrs, observedErrs, fmt.Sprintf("Failed for file: %s", file))
	}

	// Make sure all expected errors were observed
	for file, expectedErrs := range expectedLintErrors {
		observedErrs, ok := observedLintErrors[file]
		require.True(t, ok, fmt.Sprintf("Failed for file: %s", file))
		require.Equal(t, observedErrs, expectedErrs, fmt.Sprintf("Failed for file: %s", file))
	}
}
