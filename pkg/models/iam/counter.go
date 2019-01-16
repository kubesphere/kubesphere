/*

 Copyright 2019 The KubeSphere Authors.

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
package iam

import "sync"

type Counter struct {
	value int
	m     sync.Mutex
}

func NewCounter(value int) Counter {
	c := Counter{}
	c.m = sync.Mutex{}
	c.Set(value)
	return c
}

func (c *Counter) Set(value int) {
	c.m.Lock()
	c.value = value
	c.m.Unlock()
}

func (c *Counter) Add(value int) {
	c.m.Lock()
	c.value += value
	c.m.Unlock()
}

func (c *Counter) Sub(value int) {
	c.m.Lock()
	c.value -= value
	c.m.Unlock()
}

func (c *Counter) Get() int {
	return c.value
}
