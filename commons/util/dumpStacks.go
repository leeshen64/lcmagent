package util

import (
	"runtime"
)

func DumpStacks() string {
	buf := make([]byte, 16384)
	buf = buf[:runtime.Stack(buf, true)]
	return string(buf)
}
