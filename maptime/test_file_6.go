package main

import "time"

type chanTime chan time.Time

func test6() map[chanTime]bool {
	return map[chanTime]bool{}
}
