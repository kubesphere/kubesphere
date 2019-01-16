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
package authentication

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/mholt/caddy/caddyhttp/httpserver"
	"k8s.io/api/rbac/v1"
	k8sErr "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apiserver/pkg/authorization/authorizer"
	"k8s.io/apiserver/pkg/endpoints/request"
	lister "k8s.io/client-go/listers/rbac/v1"
	"k8s.io/kubernetes/pkg/util/slice"

	"kubesphere.io/kubesphere/pkg/informers"
	sliceutils "kubesphere.io/kubesphere/pkg/utils"
)

type Authentication struct {
	Rule Rule
	Next httpserver.Handler
}

type Rule struct {
	Path         string
	ExceptedPath []string
}

var (
	clusterRoleLister        lister.ClusterRoleLister
	clusterRoleBindingLister lister.ClusterRoleBindingLister
	roleLister               lister.RoleLister
	roleBindingLister        lister.RoleBindingLister
)

func init() {
	clusterRoleLister = informers.SharedInformerFactory().Rbac().V1().ClusterRoles().Lister()
	roleLister = informers.SharedInformerFactory().Rbac().V1().Roles().Lister()
}

func getFixedAttributes(ctx context.Context) (authorizer.Attributes, error) {
	attribs := authorizer.AttributesRecord{}

	user, ok := request.UserFrom(ctx)
	if ok {
		attribs.User = user
	}

	requestInfo, found := request.RequestInfoFrom(ctx)
	if !found {
		return nil, errors.New("no RequestInfo found in the context")
	}

	// Start with common attributes that apply to resource and non-resource requests
	attribs.ResourceRequest = requestInfo.IsResourceRequest
	attribs.Path = requestInfo.Path
	attribs.Verb = requestInfo.Verb

	attribs.APIGroup = requestInfo.APIGroup
	attribs.APIVersion = requestInfo.APIVersion
	attribs.Resource = requestInfo.Resource
	attribs.Subresource = requestInfo.Subresource
	attribs.Namespace = requestInfo.Namespace
	attribs.Name = requestInfo.Name

	// Hard fix
	if requestInfo.Name == "namespaces" {
		switch requestInfo.Resource {
		case "volumes":
			attribs.Namespace = requestInfo.Subresource
			attribs.Resource = "persistentvolumeclaims"
			attribs.Name = ""
			attribs.Subresource = ""
			attribs.Verb = "list"
			break
		case "status":
			fallthrough
		case "monitoring":
			fallthrough
		case "quota":
			attribs.Namespace = requestInfo.Subresource
			attribs.Subresource = ""
			attribs.Resource = requestInfo.Resource
			attribs.Name = ""
			attribs.Verb = "list"
			break
		}
	}

	if requestInfo.Subresource == "members" && requestInfo.Resource == "workspaces" && requestInfo.APIGroup == "account.kubesphere.io" {
		attribs.APIGroup = "kubesphere.io"
	}

	return &attribs, nil
}

func (c Authentication) ServeHTTP(w http.ResponseWriter, r *http.Request) (int, error) {

	if httpserver.Path(r.URL.Path).Matches(c.Rule.Path) {

		attrs, err := getFixedAttributes(r.Context())

		// without auth info
		if err != nil {
			return c.Next.ServeHTTP(w, r)
		}

		for _, path := range c.Rule.ExceptedPath {
			if httpserver.Path(r.URL.Path).Matches(path) {
				return c.Next.ServeHTTP(w, r)
			}
		}

		permitted, err := permissionValidate(attrs)

		if err != nil {
			return http.StatusInternalServerError, err
		}

		if !permitted {
			err = k8sErr.NewForbidden(schema.GroupResource{Group: attrs.GetAPIGroup(), Resource: attrs.GetResource()}, attrs.GetName(), fmt.Errorf("permission undefined"))
			return handleForbidden(w, err), nil
		}
	}

	return c.Next.ServeHTTP(w, r)

}

func handleForbidden(w http.ResponseWriter, err error) int {
	message := fmt.Sprintf("Forbidden,%s", err.Error())
	w.Header().Add("WWW-Authenticate", message)
	return http.StatusForbidden
}

func permissionValidate(attrs authorizer.Attributes) (bool, error) {

	if openAPIValidate(attrs) {
		return true, nil
	}

	permitted, err := clusterRoleValidate(attrs)

	if err != nil {
		return false, err
	}

	if permitted {
		return true, nil
	}

	if attrs.GetNamespace() != "" {
		permitted, err = roleValidate(attrs)

		if err != nil {
			return false, err
		}

		if permitted {
			return true, nil
		}
	}

	return false, nil
}

func roleValidate(attrs authorizer.Attributes) (bool, error) {

	roleBindings, err := roleBindingLister.RoleBindings(attrs.GetNamespace()).List(labels.Everything())

	if err != nil {
		return false, err
	}

	fullSource := attrs.GetResource()

	if attrs.GetSubresource() != "" {
		fullSource = fullSource + "/" + attrs.GetSubresource()
	}

	for _, roleBinding := range roleBindings {

		for _, subj := range roleBinding.Subjects {

			if (subj.Kind == v1.UserKind && subj.Name == attrs.GetUser().GetName()) ||
				(subj.Kind == v1.GroupKind && slice.ContainsString(attrs.GetUser().GetGroups(), subj.Name, nil)) {
				role, err := roleLister.Roles(attrs.GetNamespace()).Get(roleBinding.RoleRef.Name)

				if err != nil {
					return false, err
				}

				for _, rule := range role.Rules {
					if ruleMatchesRequest(rule, attrs.GetAPIGroup(), "", attrs.GetResource(), attrs.GetSubresource(), attrs.GetName(), attrs.GetVerb()) {
						return true, nil
					}
				}
			}
		}
	}

	return false, nil
}

func openAPIValidate(attrs authorizer.Attributes) bool {

	combinedResource := attrs.GetResource()

	if attrs.GetSubresource() != "" {
		combinedResource = combinedResource + "/" + attrs.GetSubresource()
	}

	if strings.HasPrefix(attrs.GetPath(), "/job") {
		return true
	}

	if attrs.GetPath() == "/apis/kubesphere.io/v1alpha1/registries/validation" {
		return true
	}

	//if attrs.GetPath() == "/apis/account.kubesphere.io/v1alpha1/users/current" && attrs.GetVerb() == "get" {
	//	return true
	//}

	if attrs.GetResource() == "users" && attrs.GetVerb() == "get" && attrs.GetName() == "current" {
		return true
	}
	if attrs.GetResource() == "users" && attrs.GetName() == attrs.GetUser().GetName() {
		return true
	}

	if attrs.GetPath() == "/apis/kubesphere.io/v1alpha1/workspaces" && attrs.GetVerb() == "list" {
		return true
	}

	if combinedResource == "rulesmapping" && attrs.GetVerb() == "get" {
		return true
	}

	if combinedResource == "workspaces/rules" && attrs.GetVerb() == "get" {
		return true
	}

	if combinedResource == "workspaces/roles" && attrs.GetVerb() == "get" {
		return true
	}

	if combinedResource == "workspaces/namespaces" && attrs.GetVerb() == "get" {
		return true
	}

	if combinedResource == "workspaces/devops" && attrs.GetVerb() == "get" {
		return true
	}

	return false
}

func clusterRoleValidate(attrs authorizer.Attributes) (bool, error) {

	clusterRoleBindings, err := clusterRoleBindingLister.List(labels.Everything())

	if err != nil {
		return false, err
	}

	for _, clusterRoleBinding := range clusterRoleBindings {

		for _, subject := range clusterRoleBinding.Subjects {

			if (subject.Kind == v1.UserKind && subject.Name == attrs.GetUser().GetName()) ||
				(subject.Kind == v1.GroupKind && sliceutils.HasString(attrs.GetUser().GetGroups(), subject.Name)) {

				clusterRole, err := clusterRoleLister.Get(clusterRoleBinding.RoleRef.Name)

				if err != nil {
					return false, err
				}

				for _, rule := range clusterRole.Rules {
					if attrs.IsResourceRequest() {
						if ruleMatchesRequest(rule, attrs.GetAPIGroup(), "", attrs.GetResource(), attrs.GetSubresource(), attrs.GetName(), attrs.GetVerb()) {
							return true, nil
						}
					} else {
						if ruleMatchesRequest(rule, "", attrs.GetPath(), "", "", "", attrs.GetVerb()) {
							return true, nil
						}
					}

				}

			}
		}
	}

	return false, nil
}

func ruleMatchesResources(rule v1.PolicyRule, apiGroup string, resource string, subresource string, resourceName string) bool {

	if resource == "" {
		return false
	}

	if !sliceutils.HasString(rule.APIGroups, apiGroup) && !sliceutils.HasString(rule.APIGroups, v1.ResourceAll) {
		return false
	}

	if len(rule.ResourceNames) > 0 && !sliceutils.HasString(rule.ResourceNames, resourceName) {
		return false
	}

	combinedResource := resource

	if subresource != "" {
		combinedResource = combinedResource + "/" + subresource
	}

	for _, res := range rule.Resources {

		// match "*"
		if res == v1.ResourceAll || res == combinedResource {
			return true
		}

		// match "*/subresource"
		if len(subresource) > 0 && strings.HasPrefix(res, "*/") && subresource == strings.TrimLeft(res, "*/") {
			return true
		}
		// match "resource/*"
		if strings.HasSuffix(res, "/*") && resource == strings.TrimRight(res, "/*") {
			return true
		}
	}

	return false
}

func ruleMatchesRequest(rule v1.PolicyRule, apiGroup string, nonResourceURL string, resource string, subresource string, resourceName string, verb string) bool {

	if !sliceutils.HasString(rule.Verbs, verb) && !sliceutils.HasString(rule.Verbs, v1.VerbAll) {
		return false
	}

	if nonResourceURL == "" {
		return ruleMatchesResources(rule, apiGroup, resource, subresource, resourceName)
	} else {
		return ruleMatchesNonResource(rule, nonResourceURL)
	}
}

func ruleMatchesNonResource(rule v1.PolicyRule, nonResourceURL string) bool {

	if nonResourceURL == "" {
		return false
	}

	for _, spec := range rule.NonResourceURLs {
		if pathMatches(nonResourceURL, spec) {
			return true
		}
	}

	return false
}

func pathMatches(path, spec string) bool {
	if spec == "*" {
		return true
	}
	if spec == path {
		return true
	}
	if strings.HasSuffix(spec, "*") && strings.HasPrefix(path, strings.TrimRight(spec, "*")) {
		return true
	}
	return false
}
