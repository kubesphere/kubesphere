package log

import (
	"fmt"

	"github.com/golang/glog"
)

const (
	debug glog.Level = glog.Level(4)
	trace glog.Level = glog.Level(5)
)

func Info(args ...interface{}) {
	glog.InfoDepth(1, args...)
}

func Infof(format string, args ...interface{}) {
	glog.InfoDepth(1, fmt.Sprintf(format, args...))
}

func Warning(args ...interface{}) {
	glog.WarningDepth(1, args...)
}

func Warningf(format string, args ...interface{}) {
	glog.WarningDepth(1, fmt.Sprintf(format, args...))
}

func Error(args ...interface{}) {
	glog.ErrorDepth(1, args...)
}

func Errorf(format string, args ...interface{}) {
	glog.ErrorDepth(1, fmt.Sprintf(format, args...))
}

// Debug will log a message at verbose level 4 and will ensure the caller's stack frame is used
func Debug(args ...interface{}) {
	if glog.V(debug) {
		glog.InfoDepth(1, "DEBUG: "+fmt.Sprint(args...)) // 1 == depth in the stack of the caller
	}
}

// Debugf will log a message at verbose level 4 and will ensure the caller's stack frame is used
func Debugf(format string, args ...interface{}) {
	if glog.V(debug) {
		glog.InfoDepth(1, fmt.Sprintf("DEBUG: "+format, args...)) // 1 == depth in the stack of the caller
	}
}

func IsDebug() bool {
	return bool(glog.V(debug))
}

// Trace will log a message at verbose level 5 and will ensure the caller's stack frame is used
func Trace(args ...interface{}) {
	if glog.V(trace) {
		glog.InfoDepth(1, "TRACE: "+fmt.Sprint(args...)) // 1 == depth in the stack of the caller
	}
}

// Tracef will log a message at verbose level 5 and will ensure the caller's stack frame is used
func Tracef(format string, args ...interface{}) {
	if glog.V(trace) {
		glog.InfoDepth(1, fmt.Sprintf("TRACE: "+format, args...)) // 1 == depth in the stack of the caller
	}
}

func IsTrace() bool {
	return bool(glog.V(trace))
}
