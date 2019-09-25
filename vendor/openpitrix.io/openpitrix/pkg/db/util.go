// Copyright 2018 The OpenPitrix Authors. All rights reserved.
// Use of this source code is governed by a Apache license
// that can be found in the LICENSE file.

package db

import (
	"github.com/fatih/structs"

	"openpitrix.io/openpitrix/pkg/util/stringutil"
)

func GetColumnsFromStruct(s interface{}) []string {
	names := structs.Names(s)
	for i, name := range names {
		names[i] = stringutil.CamelCaseToUnderscore(name)
	}
	return names
}

func GetColumnsFromStructWithPrefix(prefix string, s interface{}) []string {
	names := structs.Names(s)
	for i, name := range names {
		names[i] = WithPrefix(prefix, stringutil.CamelCaseToUnderscore(name))
	}
	return names
}

func WithPrefix(prefix, str string) string {
	return prefix + "." + str
}
