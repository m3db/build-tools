package testfiles

import "time"

type structWithInnerTime struct {
	inner time.Time
}

func test5() map[structWithInnerTime]bool {
	return map[structWithInnerTime]bool{}
}
