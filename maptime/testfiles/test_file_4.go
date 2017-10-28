package testfiles

type sneakyStruct struct {
	inner map[timeAlias]bool
}

func test4() sneakyStruct {
	return sneakyStruct{}
}
