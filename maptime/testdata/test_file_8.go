// +build excluded

// This file verifies that files with unspecified build flags will not be linted
package testdata

import "time"

func test8() map[time.Time]bool {
	return map[time.Time]bool{}
}
