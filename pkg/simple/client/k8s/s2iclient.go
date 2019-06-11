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
package k8s

import (
	"log"
	"sync"

	s2i "github.com/kubesphere/s2ioperator/pkg/client/clientset/versioned"
)

var (
	s2iClient     *s2i.Clientset
	s2iClientOnce sync.Once
)

func S2iClient() *s2i.Clientset {

	s2iClientOnce.Do(func() {

		config, err := Config()

		if err != nil {
			log.Fatalln(err)
		}

		s2iClient = s2i.NewForConfigOrDie(config)
	})

	return s2iClient
}
