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

package idutils

import (
	"errors"
	"net"
	"os"

	"github.com/golang/example/stringutil"
	"github.com/sony/sonyflake"
	hashids "github.com/speps/go-hashids"

	"kubesphere.io/kubesphere/pkg/utils/stringutils"
)

var sf *sonyflake.Sonyflake

func init() {
	var st sonyflake.Settings
	if len(os.Getenv("DEVOPSPHERE_IP")) != 0 {
		st.MachineID = machineID
	}
	sf = sonyflake.NewSonyflake(st)
	if sf == nil {
		panic("failed to initialize sonyflake")
	}
}

func GetIntId() uint64 {
	id, err := sf.NextID()
	if err != nil {
		panic(err)
	}
	return id
}

// format likes: B6BZVN3mOPvx
func GetUuid(prefix string) string {
	id := GetIntId()
	hd := hashids.NewData()
	h, err := hashids.NewWithData(hd)
	if err != nil {
		panic(err)
	}
	i, err := h.Encode([]int{int(id)})
	if err != nil {
		panic(err)
	}

	return prefix + stringutils.Reverse(i)
}

const Alphabet36 = "abcdefghijklmnopqrstuvwxyz1234567890"

// format likes: 300m50zn91nwz5
func GetUuid36(prefix string) string {
	id := GetIntId()
	hd := hashids.NewData()
	hd.Alphabet = Alphabet36
	h, err := hashids.NewWithData(hd)
	if err != nil {
		panic(err)
	}
	i, err := h.Encode([]int{int(id)})
	if err != nil {
		panic(err)
	}

	return prefix + stringutil.Reverse(i)
}

func machineID() (uint16, error) {
	ipStr := os.Getenv("DEVOPSPHERE_IP")
	if len(ipStr) == 0 {
		return 0, errors.New("'DEVOPSPHERE_IP' environment variable not set")
	}
	ip := net.ParseIP(ipStr)
	if len(ip) < 4 {
		return 0, errors.New("invalid IP")
	}
	return uint16(ip[2])<<8 + uint16(ip[3]), nil
}
