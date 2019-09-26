// Copyright 2018 The OpenPitrix Authors. All rights reserved.
// Use of this source code is governed by a Apache license
// that can be found in the LICENSE file.

package sender

import (
	"strings"

	"github.com/golang/protobuf/ptypes/wrappers"
)

// ${group_path}:${user_id}
type OwnerPath string

func (o OwnerPath) match(accessPath OwnerPath) bool {
	if string(accessPath) == "" {
		return true
	}
	return strings.HasPrefix(string(o), string(accessPath))
}

func (o OwnerPath) CheckOwnerPathPermission(ownerPaths ...string) bool {
	for _, ownerPath := range ownerPaths {
		if !OwnerPath(ownerPath).match(o) {
			return false
		}
	}
	return true
}

func (o OwnerPath) CheckPermission(s *Sender) bool {
	return o.match(s.GetAccessPath())
}

func (o OwnerPath) Owner() string {
	s := strings.Split(string(o), ":")
	if len(s) < 2 {
		return ""
	}
	return s[1]
}

func (o OwnerPath) ToProtoString() *wrappers.StringValue {
	return &wrappers.StringValue{Value: string(o)}
}
