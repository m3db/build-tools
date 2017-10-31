// +build included

// This file verifies that files with matching build flags will be linted
package testdata

import "time"

func test7() map[time.Time]bool {
	return map[time.Time]bool{}
}
