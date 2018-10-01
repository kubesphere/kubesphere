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

package organizations

//TODO this function will be implemented by whm later
// get namespaces
func GetNamespaces(name string) ([]string, error) {
	return []string{"monitoring", "kube-system"}, nil
}

//TODO this function will be implemented by whm later
// get devops projects
func GetDevOpsProjects(name string) ([]string, error) {
	return []string{"devpos01"}, nil
}

//TODO this function will be implemented by whm later
// get enterprise members
func GetOrgMembers(name string) ([]string, error) {
	return []string{"tom", "jack", "mike"}, nil
}

//TODO this function will be implemented by whm later
// get enterprise roles
func GetOrgRoles(name string) ([]string, error) {
	return []string{"role01", "role02", "role03", "role04"}, nil
}

//TODO this function will be implemented by whm later
func GetAllOrgNums() (int64, error) {
	return 51, nil
}

//TODO this function will be implemented by whm later
func GetAllDevOpsProjectsNums() (int64, error) {
	return 52, nil
}

//TODO this function will be implemented by whm later
func GetOrgMembersNums() (int64, error) {
	return 53, nil
}

//TODO this function will be implemented by whm later
func GetOrgRolesNums() (int64, error) {
	return 54, nil
}
