// Copyright 2015 Vadim Kravcenko
//
// Licensed under the Apache License, Version 2.0 (the "License"): you may
// not use this file except in compliance with the License. You may obtain
// a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS, WITHOUT
// WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the
// License for the specific language governing permissions and limitations
// under the License.

package gojenkins

import (
	"crypto/md5"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path"
)

// Represents an Artifact
type Artifact struct {
	Jenkins  *Jenkins
	Build    *Build
	FileName string
	Path     string
}

// Get raw byte data of Artifact
func (a Artifact) GetData() ([]byte, error) {
	var data string
	response, err := a.Jenkins.Requester.Get(a.Path, &data, nil)

	if err != nil {
		return nil, err
	}

	code := response.StatusCode
	if code != 200 {
		Error.Printf("Jenkins responded with StatusCode: %d", code)
		return nil, errors.New("Could not get File Contents")
	}
	return []byte(data), nil
}

// Save artifact to a specific path, using your own filename.
func (a Artifact) Save(path string) (bool, error) {
	data, err := a.GetData()

	if err != nil {
		return false, errors.New("No Data received, not saving file.")
	}

	if _, err = os.Stat(path); err == nil {
		Warning.Println("Local Copy already exists, Overwriting...")
	}

	err = ioutil.WriteFile(path, data, 0644)
	a.validateDownload(path)

	if err != nil {
		return false, err
	}
	return true, nil
}

// Save Artifact to directory using Artifact filename.
func (a Artifact) SaveToDir(dir string) (bool, error) {
	if _, err := os.Stat(dir); err != nil {
		Error.Printf("can't save artifact: directory %s does not exist", dir)
		return false, fmt.Errorf("can't save artifact: directory %s does not exist", dir)
	}
	saved, err := a.Save(path.Join(dir, a.FileName))
	if err != nil {
		return saved, nil
	}
	return saved, nil
}

// Compare Remote and local MD5
func (a Artifact) validateDownload(path string) (bool, error) {
	localHash := a.getMD5local(path)

	fp := FingerPrint{Jenkins: a.Jenkins, Base: "/fingerprint/", Id: localHash, Raw: new(FingerPrintResponse)}

	valid, err := fp.ValidateForBuild(a.FileName, a.Build)

	if err != nil {
		return false, err
	}
	if !valid {
		return false, errors.New("FingerPrint of the downloaded artifact could not be verified")
	}
	return true, nil
}

// Get Local MD5
func (a Artifact) getMD5local(path string) string {
	h := md5.New()
	localFile, err := os.Open(path)
	if err != nil {
		return ""
	}
	buffer := make([]byte, 2^20)
	n, err := localFile.Read(buffer)
	defer localFile.Close()
	for err == nil {
		io.WriteString(h, string(buffer[0:n]))
		n, err = localFile.Read(buffer)
	}
	return fmt.Sprintf("%x", h.Sum(nil))
}
