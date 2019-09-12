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

import (
	"fmt"
	"kubesphere.io/kubesphere/pkg/simple/client"
	"regexp"
	"strings"
)

func convertDNToPath(dn string) string {

	paths := regexp.MustCompile("cn=[a-z0-9]([-a-z0-9]*[a-z0-9])?").FindAllString(dn, -1)

	if len(paths) > 1 {
		for i := 0; i < len(paths); i++ {
			paths[i] = strings.Replace(paths[i], "cn=", "", 1)
		}
		for i, j := 0, len(paths)-1; i < j; i, j = i+1, j-1 {
			paths[i], paths[j] = paths[j], paths[i]
		}
		return strings.Join(paths, ":")
	} else if len(paths) == 1 {
		return strings.Replace(paths[0], "cn=", "", -1)
	} else {
		return ""
	}
}

func splitPath(path string) (searchBase string, cn string) {
	ldapClient, err := client.ClientSets().Ldap()
	if err != nil {
		return "", ""
	}

	paths := strings.Split(path, ":")
	length := len(paths)
	if length > 2 {

		cn = paths[length-1]
		basePath := paths[:length-1]

		for i := 0; i < len(basePath); i++ {
			basePath[i] = fmt.Sprintf("cn=%s", basePath[i])
		}

		for i, j := 0, length-2; i < j; i, j = i+1, j-1 {
			basePath[i], basePath[j] = basePath[j], basePath[i]
		}

		searchBase = fmt.Sprintf("%s,%s", strings.Join(basePath, ","), ldapClient.GroupSearchBase())
	} else if length == 2 {
		searchBase = fmt.Sprintf("cn=%s,%s", paths[0], ldapClient.GroupSearchBase())
		cn = paths[1]
	} else {
		searchBase = ldapClient.GroupSearchBase()
		if paths[0] == "" {
			cn = "*"
		} else {
			cn = paths[0]
		}
	}
	return
}
