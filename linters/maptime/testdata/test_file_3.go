package testdata

import "time"

type timeMapAlias map[time.Time]bool

func test3() timeMapAlias {
	return timeMapAlias{}
}
