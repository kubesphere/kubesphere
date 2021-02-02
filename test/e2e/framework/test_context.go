/*
Copyright 2020 KubeSphere Authors

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

package framework

import (
	"flag"
	"os"
)

type TestContextType struct {
	Host         string
	InMemoryTest bool
	Username     string
	Password     string
}

func registerFlags(t *TestContextType) {
	flag.BoolVar(&t.InMemoryTest, "in-memory-test", false,
		"Whether KubeSphere controllers and APIServer be started in memory.")
	flag.StringVar(&t.Host, "ks-apiserver", os.Getenv("KS_APISERVER"),
		"KubeSphere API Server IP/DNS")
	flag.StringVar(&t.Username, "username", os.Getenv("KS_USERNAME"),
		"Username to login to KubeSphere API Server")
	flag.StringVar(&t.Password, "password", os.Getenv("KS_PASSWORD"),
		"Password to login to KubeSphere API Server")
}

var TestContext *TestContextType = &TestContextType{}

func ParseFlags() {
	registerFlags(TestContext)
	flag.Parse()
}
