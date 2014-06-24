package utils

import (
	"runtime"

	"github.com/idealeak/goserver/core/logger"
)

var AvoidRepeateDumper = make(map[string][]uintptr)

func DumpStackIfPanic(f string) {
	if err := recover(); err != nil {
		logger.Logger.Error(f, " panic,error=", err)
		var buf [4096]byte
		len := runtime.Stack(buf[:], false)
		logger.Logger.Error("stack--->", string(buf[:len]))
	}
}
