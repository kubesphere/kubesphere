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

package readerutils

import (
	"crypto/md5"
	"hash"
	"io"
)

type MD5Reader struct {
	md5  hash.Hash
	body io.Reader
}

func (reader *MD5Reader) Read(b []byte) (int, error) {
	n, err := reader.body.Read(b)
	if err != nil {
		return n, err
	}
	return reader.md5.Write(b[:n])
}

func (reader *MD5Reader) MD5() []byte {
	return reader.md5.Sum(nil)
}

func NewMD5Reader(reader io.Reader) *MD5Reader {
	return &MD5Reader{
		md5:  md5.New(),
		body: reader,
	}
}
