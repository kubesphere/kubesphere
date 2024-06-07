/*
 * Please refer to the LICENSE file in the root directory of the project.
 * https://github.com/kubesphere/kubesphere/blob/master/LICENSE
 */

package log

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"

	"gopkg.in/natefinch/lumberjack.v2"
	"k8s.io/klog/v2"

	"kubesphere.io/kubesphere/pkg/apiserver/auditing/internal"
)

const (
	WriteTimeout      = time.Second * 3
	DefaultMaxAge     = 7
	DefaultMaxBackups = 10
	DefaultMaxSize    = 100
)

type backend struct {
	path       string
	maxAge     int
	maxBackups int
	maxSize    int
	timeout    time.Duration

	writer io.Writer
}

func NewBackend(path string, maxAge, maxBackups, maxSize int) internal.Backend {
	b := backend{
		path:       path,
		maxAge:     maxAge,
		maxBackups: maxBackups,
		maxSize:    maxSize,
		timeout:    WriteTimeout,
	}

	if b.maxAge == 0 {
		b.maxAge = DefaultMaxAge
	}

	if b.maxBackups == 0 {
		b.maxBackups = DefaultMaxBackups
	}

	if b.maxSize == 0 {
		b.maxSize = DefaultMaxSize
	}

	if err := b.ensureLogFile(); err != nil {
		klog.Errorf("ensure audit log file error, %s", err)
		return nil
	}

	b.writer = &lumberjack.Logger{
		Filename:   b.path,
		MaxAge:     b.maxAge,
		MaxBackups: b.maxBackups,
		MaxSize:    b.maxSize,
		Compress:   false,
	}

	return &b
}

func (b *backend) ensureLogFile() error {
	if err := os.MkdirAll(filepath.Dir(b.path), 0700); err != nil {
		return err
	}
	mode := os.FileMode(0600)
	f, err := os.OpenFile(b.path, os.O_CREATE|os.O_APPEND|os.O_RDWR, mode)
	if err != nil {
		return err
	}
	return f.Close()
}

func (b *backend) ProcessEvents(events ...[]byte) {
	for _, event := range events {
		if _, err := fmt.Fprint(b.writer, string(event)+"\n"); err != nil {
			klog.Errorf("Log audit event error, %s. affecting audit event: %v\nImpacted event:\n", err, event)
			klog.Error(string(event))
		}
	}
}
