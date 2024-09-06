/*
 * Please refer to the LICENSE file in the root directory of the project.
 * https://github.com/kubesphere/kubesphere/blob/master/LICENSE
 */

package rbac

import (
	"testing"

	rbacv1 "k8s.io/api/rbac/v1"
	"sigs.k8s.io/yaml"
)

func TestSquashRules(t *testing.T) {

	var raw = `
- apiGroups:
  - application.kubesphere.io
  resources:
  - apps
  verbs:
  - '*'
- apiGroups:
  - application.kubesphere.io
  resources:
  - apps/versions
  verbs:
  - '*'
- apiGroups:
  - application.kubesphere.io
  resources:
  - applications
  verbs:
  - '*'
- apiGroups:
  - application.kubesphere.io
  resources:
  - attachments
  verbs:
  - '*'
- apiGroups:
  - application.kubesphere.io
  resources:
  - repos
  verbs:
  - '*'
- apiGroups:
  - application.kubesphere.io
  resources:
  - repos/events
  verbs:
  - '*'
- apiGroups:
  - '*'
  resources:
  - workspaces
  verbs:
  - get
- apiGroups:
  - '*'
  resources:
  - workspaces
  verbs:
  - list
- apiGroups:
  - '*'
  resources:
  - workspaces
  verbs:
  - watch
- apiGroups:
  - '*'
  resources:
  - workspaces
  verbs:
  - get
- apiGroups:
  - '*'
  resources:
  - workspaces
  verbs:
  - list
- apiGroups:
  - '*'
  resources:
  - workspaces
  verbs:
  - watch
- apiGroups:
  - '*'
  resources:
  - workspacemembers
  verbs:
  - get
- apiGroups:
  - '*'
  resources:
  - workspacemembers
  verbs:
  - list
- apiGroups:
  - '*'
  resources:
  - workspacemembers
  verbs:
  - watch
- apiGroups:
  - '*'
  resources:
  - quotas
  verbs:
  - get
- apiGroups:
  - '*'
  resources:
  - quotas
  verbs:
  - list
- apiGroups:
  - '*'
  resources:
  - quotas
  verbs:
  - watch
- apiGroups:
  - '*'
  resources:
  - abnormalworkloads
  verbs:
  - get
- apiGroups:
  - '*'
  resources:
  - abnormalworkloads
  verbs:
  - list
- apiGroups:
  - '*'
  resources:
  - abnormalworkloads
  verbs:
  - watch
- apiGroups:
  - '*'
  resources:
  - pods
  verbs:
  - get
- apiGroups:
  - '*'
  resources:
  - pods
  verbs:
  - list
- apiGroups:
  - '*'
  resources:
  - pods
  verbs:
  - watch
- apiGroups:
  - '*'
  resources:
  - namespaces
  verbs:
  - create
- apiGroups:
  - '*'
  resources:
  - namespaces
  verbs:
  - watch
- apiGroups:
  - '*'
  resources:
  - federatednamespaces
  verbs:
  - create
- apiGroups:
  - '*'
  resources:
  - federatednamespaces
  verbs:
  - watch`

	var expectedRaw = `- apiGroups:
    - application.kubesphere.io
  resources:
    - apps
    - apps/versions
    - applications
    - attachments
    - repos
    - repos/events
  verbs:
    - '*'
- apiGroups:
    - '*'
  resources:
    - quotas
    - workspacemembers
    - workspaces
    - abnormalworkloads
    - pods
  verbs:
    - get
    - list
    - watch
- apiGroups:
    - '*'
  resources:
    - namespaces
    - federatednamespaces
  verbs:
    - create
    - watch`

	rules := []rbacv1.PolicyRule{}
	err := yaml.Unmarshal([]byte(raw), &rules)
	if err != nil {
		t.Fatal(err)
	}

	squashRules := SquashRules(len(rules), rules)

	expectedRules := []rbacv1.PolicyRule{}
	err = yaml.Unmarshal([]byte(expectedRaw), &expectedRules)
	if err != nil {
		t.Fatal(err)
	}

	lefCovers, _ := Covers(expectedRules, squashRules)
	rightCover, _ := Covers(squashRules, expectedRules)

	if !lefCovers || !rightCover {
		t.Errorf("failed")
	}

}
