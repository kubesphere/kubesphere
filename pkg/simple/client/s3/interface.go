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

package s3

import (
	"io"
)

type Interface interface {
	//read the content, caller should close the io.ReadCloser.
	Read(key string) ([]byte, error)

	// Upload uploads a object to storage and returns object location if succeeded
	Upload(key, fileName string, body io.Reader) error

	GetDownloadURL(key string, fileName string) (string, error)

	// Delete deletes an object by its key
	Delete(key string) error
}
