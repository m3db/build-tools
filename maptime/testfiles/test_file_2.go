package testfiles

import "time"

type timeAlias time.Time

func test2() map[timeAlias]bool {
	return map[timeAlias]bool{}
}
