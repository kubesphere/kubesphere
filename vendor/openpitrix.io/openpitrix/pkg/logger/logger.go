// Copyright 2018 The OpenPitrix Authors. All rights reserved.
// Use of this source code is governed by a Apache license
// that can be found in the LICENSE file.

package logger

import (
	"context"
	"fmt"
	"io"
	"os"
	"runtime"
	"strings"
	"sync/atomic"
	"time"

	"openpitrix.io/openpitrix/pkg/util/ctxutil"
)

type Level uint32

const (
	CriticalLevel Level = iota
	ErrorLevel
	WarnLevel
	InfoLevel
	DebugLevel
)

func (level Level) String() string {
	switch level {
	case DebugLevel:
		return "debug"
	case InfoLevel:
		return "info"
	case WarnLevel:
		return "warning"
	case ErrorLevel:
		return "error"
	case CriticalLevel:
		return "critical"
	}

	return "unknown"
}

func StringToLevel(level string) Level {
	switch level {
	case "critical":
		return CriticalLevel
	case "error":
		return ErrorLevel
	case "warn", "warning":
		return WarnLevel
	case "debug":
		return DebugLevel
	case "info":
		return InfoLevel
	}
	return InfoLevel
}

var logger = NewLogger().WithDepth(4)

func Info(ctx context.Context, format string, v ...interface{}) {
	logger.Info(ctx, format, v...)
}

func Debug(ctx context.Context, format string, v ...interface{}) {
	logger.Debug(ctx, format, v...)
}

func Warn(ctx context.Context, format string, v ...interface{}) {
	logger.Warn(ctx, format, v...)
}

func Error(ctx context.Context, format string, v ...interface{}) {
	logger.Error(ctx, format, v...)
}

func Critical(ctx context.Context, format string, v ...interface{}) {
	logger.Critical(ctx, format, v...)
}

func SetOutput(output io.Writer) {
	logger.SetOutput(output)
}

var globalLogLevel = InfoLevel

func SetLevelByString(level string) {
	logger.SetLevelByString(level)
	globalLogLevel = StringToLevel(level)
}

func NewLogger() *Logger {
	return &Logger{
		Level:  globalLogLevel,
		output: os.Stdout,
		depth:  3,
	}
}

type Logger struct {
	Level         Level
	output        io.Writer
	hideCallstack bool
	depth         int
}

func (logger *Logger) level() Level {
	return Level(atomic.LoadUint32((*uint32)(&logger.Level)))
}

func (logger *Logger) SetLevel(level Level) {
	atomic.StoreUint32((*uint32)(&logger.Level), uint32(level))
}

func (logger *Logger) SetLevelByString(level string) {
	logger.SetLevel(StringToLevel(level))
}

var replacer = strings.NewReplacer("\r", "\\r", "\n", "\\n")

func (logger *Logger) formatOutput(ctx context.Context, level Level, output string) string {
	now := time.Now().Format("2006-01-02 15:04:05.99999")
	messageId := ctxutil.GetMessageId(ctx)
	requestId := ctxutil.GetRequestId(ctx)

	output = replacer.Replace(output)

	var suffix string
	if len(requestId) > 0 {
		messageId = append(messageId, requestId)
	}
	if len(messageId) > 0 {
		suffix = fmt.Sprintf("(%s)", strings.Join(messageId, "|"))
	}
	if logger.hideCallstack {
		return fmt.Sprintf("%-25s -%s- %s%s",
			now, strings.ToUpper(level.String()), output, suffix)
	} else {
		_, file, line, ok := runtime.Caller(logger.depth)
		if !ok {
			file = "???"
			line = 0
		}
		// short file name
		for i := len(file) - 1; i > 0; i-- {
			if file[i] == '/' {
				file = file[i+1:]
				break
			}
		}
		// 2018-03-27 02:08:44.93894 -INFO- Api service start http://openpitrix-api-gateway:9100 (main.go:44)
		return fmt.Sprintf("%-25s -%s- %s (%s:%d)%s",
			now, strings.ToUpper(level.String()), output, file, line, suffix)
	}
}

func (logger *Logger) logf(ctx context.Context, level Level, format string, args ...interface{}) {
	if logger.level() < level {
		return
	}
	fmt.Fprintln(logger.output, logger.formatOutput(ctx, level, fmt.Sprintf(format, args...)))
}

func (logger *Logger) Debug(ctx context.Context, format string, args ...interface{}) {
	logger.logf(ctx, DebugLevel, format, args...)
}

func (logger *Logger) Info(ctx context.Context, format string, args ...interface{}) {
	logger.logf(ctx, InfoLevel, format, args...)
}

func (logger *Logger) Warn(ctx context.Context, format string, args ...interface{}) {
	logger.logf(ctx, WarnLevel, format, args...)
}

func (logger *Logger) Error(ctx context.Context, format string, args ...interface{}) {
	logger.logf(ctx, ErrorLevel, format, args...)
}

func (logger *Logger) Critical(ctx context.Context, format string, args ...interface{}) {
	logger.logf(ctx, CriticalLevel, format, args...)
}

func (logger *Logger) SetOutput(output io.Writer) *Logger {
	logger.output = output
	return logger
}

func (logger *Logger) HideCallstack() *Logger {
	logger.hideCallstack = true
	return logger
}

func (logger *Logger) WithDepth(depth int) *Logger {
	logger.depth = depth
	return logger
}
