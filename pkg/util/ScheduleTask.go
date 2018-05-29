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

package util

import (
	"time"
	"reflect"
	"github.com/golang/glog"
)

// implement a function to execute a taskFunction at specific time and periodicity
// ex  ScheduleTask(func,"18:12:08","15s",args)



func ScheduleTask(taskFunction interface{}, start, interval string, funcArgs ...interface{}) {
	taskFuncProperty := reflect.ValueOf(taskFunction)
	if taskFuncProperty.Kind() != reflect.Func {
		glog.Fatal("only function can be schedule.")
	}
	if len(funcArgs) != taskFuncProperty.Type().NumIn() {
		glog.Fatal("The number of args valid.")
	}
	// Get function args.
	in := make([]reflect.Value, len(funcArgs))
	for i, arg := range funcArgs {
		in[i] = reflect.ValueOf(arg)
	}

	// Get interval d.
	d, err := time.ParseDuration(interval)
	if err != nil {
		glog.Fatal(err)
	}
	location, err := time.LoadLocation("Asia/Shanghai")
	if err != nil {
		glog.Fatal(err)
	}
	t, err := time.ParseInLocation("15:04:05", start, location)
	if err != nil {
		glog.Fatal(err)
	}
	now := time.Now()

	// when to start.
	t = time.Date(now.Year(), now.Month(), now.Day(), t.Hour(), t.Minute(), t.Second(), 0, location)

	if now.After(t) {
		t = t.Add((now.Sub(t)/d + 1) * d)
	}

	time.Sleep(t.Sub(now))
	go taskFuncProperty.Call(in)
	ticker := time.NewTicker(d)
	go func() {
		for _ = range ticker.C {
			go taskFuncProperty.Call(in)
		}
	}()
}
