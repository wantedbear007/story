package domain

import "time"

// Now is a variable so it can be replaced in tests to control time.
var Now = func() time.Time {
	return time.Now()
}
