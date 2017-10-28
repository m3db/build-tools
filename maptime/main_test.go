package main

import (
	"fmt"
	"go/token"
	"path"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestTimeMapLint(t *testing.T) {
	expectedLintErrors := map[string][]int{
		"test_file_1.go": []int{18, 19},
		"test_file_2.go": []int{7, 8},
		"test_file_3.go": []int{5},
		"test_file_4.go": []int{4},
		"test_file_5.go": []int{9, 10},
		"test_file_6.go": []int{7, 8},
	}

	observedLintErrors := map[string][]int{}
	testCallback := func(position token.Position, keyStr, valStr string) {
		filePath := position.Filename
		filePathBase := path.Base(filePath)
		_, ok := expectedLintErrors[filePathBase]
		require.True(t, ok, fmt.Sprintf("Failed for file: %s", filePathBase))
		observedLintErrorsForFile, _ := observedLintErrors[filePathBase]
		observedLintErrors[filePathBase] = append(observedLintErrorsForFile, position.Line)
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
