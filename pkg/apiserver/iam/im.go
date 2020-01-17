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
	"k8s.io/klog"
	"kubesphere.io/kubesphere/pkg/api"
	"kubesphere.io/kubesphere/pkg/models/resources/v1alpha2"
	"kubesphere.io/kubesphere/pkg/server/params"
	"net/http"
	"strings"

	"github.com/emicklei/go-restful"
	rbacv1 "k8s.io/api/rbac/v1"
	"kubesphere.io/kubesphere/pkg/models/iam"
	"kubesphere.io/kubesphere/pkg/server/errors"
)

func isUserManager(username string) (bool, error) {
	rules, err := iam.GetUserClusterRules(username)
	if err != nil {
		return false, err
	}
	if iam.RulesMatchesRequired(rules, rbacv1.PolicyRule{Verbs: []string{"update"}, Resources: []string{"users"}, APIGroups: []string{"iam.kubesphere.io"}}) {
		return true, nil
	}
	return false, nil
}

func UserLoginLogs(req *restful.Request, resp *restful.Response) {
	username := req.PathParameter("user")
	logs, err := iam.LoginLog(username)

	if err != nil {
		klog.Error(err)
		resp.WriteHeaderAndEntity(http.StatusInternalServerError, errors.Wrap(err))
		return
	}

	result := make([]map[string]string, 0)

	for _, v := range logs {
		item := strings.Split(v, ",")
		time := item[0]
		var ip string
		if len(item) > 1 {
			ip = item[1]
		}
		result = append(result, map[string]string{"login_time": time, "login_ip": ip})
	}

	resp.WriteAsJson(result)
}
