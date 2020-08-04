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

package notification

type Options struct {
	Endpoint string
}

func NewNotificationOptions() *Options {
	return &Options{
		Endpoint: "",
	}
}

func (s *Options) ApplyTo(options *Options) {
	if options == nil {
		options = s
		return
	}

	if s.Endpoint != "" {
		options.Endpoint = s.Endpoint
	}
}
