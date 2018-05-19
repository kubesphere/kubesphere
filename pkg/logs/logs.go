/*
Copyright 2018 The KubeSphere Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package logs

import (
	"flag"
	"log"

	"github.com/golang/glog"
)

func init() {

	flag.Set("logtostderr", "true")
}

// GlogWriter serves as a bridge between the standard log package and the glog package.
type GlogWriter struct{}

// Write implements the io.Writer interface.
func (writer GlogWriter) Write(data []byte) (n int, err error) {

	glog.Info(string(data))
	return len(data), nil
}

// InitLogs initializes logs the way we want for kubeSphere.
func InitLogs() {

	log.SetOutput(GlogWriter{})
	log.SetFlags(0)
}

// FlushLogs flushes logs immediately.
func FlushLogs() {

	glog.Flush()
}
