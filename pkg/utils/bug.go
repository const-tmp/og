package utils

import (
	"fmt"
	"runtime"
)

func BugPanic(message string) {
	_, file, line, _ := runtime.Caller(3)
	panic(fmt.Sprintf(
		"[ THIS IS A BUG ] %s:%d\n%s\n"+
			"please submit an issue: https://github.com/nullc4t/og/issues/new",
		file, line, message,
	))
}
