package testdata

// Make sure it doesn't panic if the key is an alias to an empty struct
type emptyAlias struct{}

func test9() map[emptyAlias]bool {
	return map[emptyAlias]bool{}
}
