package rbac

import (
	"context"
	"errors"
	"hash/fnv"
	"io"
	"sort"
	"testing"

	"github.com/google/go-cmp/cmp"
	corev1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apiserver/pkg/authentication/user"
	iamv1beta1 "kubesphere.io/api/iam/v1beta1"
	tenantv1beta1 "kubesphere.io/api/tenant/v1beta1"
	"sigs.k8s.io/controller-runtime/pkg/cache/informertest"
	runtimefakeclient "sigs.k8s.io/controller-runtime/pkg/client/fake"

	"kubesphere.io/kubesphere/pkg/apiserver/authorization/authorizer"
	"kubesphere.io/kubesphere/pkg/apiserver/request"
	"kubesphere.io/kubesphere/pkg/models/iam/am"
	"kubesphere.io/kubesphere/pkg/models/resources/v1beta1"
	"kubesphere.io/kubesphere/pkg/scheme"
)

// StaticRoles is a rule resolver that resolves from lists of role objects.
type StaticRoles struct {
	roles                 []*iamv1beta1.Role
	roleBindings          []*iamv1beta1.RoleBinding
	clusterRoles          []*iamv1beta1.ClusterRole
	clusterRoleBindings   []*iamv1beta1.ClusterRoleBinding
	workspaceRoles        []*iamv1beta1.WorkspaceRole
	workspaceRoleBindings []*iamv1beta1.WorkspaceRoleBinding
	globalRoles           []*iamv1beta1.GlobalRole
	globalRoleBindings    []*iamv1beta1.GlobalRoleBinding
	namespaces            []*corev1.Namespace
}

func (r *StaticRoles) GetRole(namespace, name string) (*iamv1beta1.Role, error) {
	if len(namespace) == 0 {
		return nil, errors.New("must provide namespace when getting role")
	}
	for _, role := range r.roles {
		if role.Namespace == namespace && role.Name == name {
			return role, nil
		}
	}
	return nil, errors.New("role not found")
}

func (r *StaticRoles) GetClusterRole(name string) (*iamv1beta1.ClusterRole, error) {
	for _, clusterRole := range r.clusterRoles {
		if clusterRole.Name == name {
			return clusterRole, nil
		}
	}
	return nil, errors.New("cluster role not found")
}

func (r *StaticRoles) ListRoleBindings(namespace string) ([]*iamv1beta1.RoleBinding, error) {
	if len(namespace) == 0 {
		return nil, errors.New("must provide namespace when listing role bindings")
	}

	var roleBindingList []*iamv1beta1.RoleBinding
	for _, roleBinding := range r.roleBindings {
		if roleBinding.Namespace != namespace {
			continue
		}
		roleBindingList = append(roleBindingList, roleBinding)
	}
	return roleBindingList, nil
}

func (r *StaticRoles) ListClusterRoleBindings() ([]*iamv1beta1.ClusterRoleBinding, error) {
	return r.clusterRoleBindings, nil
}

// compute a hash of a policy rule, so we can sort in a deterministic order
func hashOf(p rbacv1.PolicyRule) string {
	hash := fnv.New32()
	writeStrings := func(slis ...[]string) {
		for _, sli := range slis {
			for _, s := range sli {
				_, _ = io.WriteString(hash, s)
			}
		}
	}
	writeStrings(p.Verbs, p.APIGroups, p.Resources, p.ResourceNames, p.NonResourceURLs)
	return string(hash.Sum(nil))
}

// byHash sorts a set of policy rules by a hash of its fields
type byHash []rbacv1.PolicyRule

func (b byHash) Len() int           { return len(b) }
func (b byHash) Less(i, j int) bool { return hashOf(b[i]) < hashOf(b[j]) }
func (b byHash) Swap(i, j int)      { b[i], b[j] = b[j], b[i] }

func TestRBACAuthorizer(t *testing.T) {
	ruleReadPods := rbacv1.PolicyRule{
		Verbs:     []string{"GET", "WATCH"},
		APIGroups: []string{"v1"},
		Resources: []string{"pods"},
	}
	ruleReadServices := rbacv1.PolicyRule{
		Verbs:     []string{"GET", "WATCH"},
		APIGroups: []string{"v1"},
		Resources: []string{"services"},
	}
	ruleWriteNodes := rbacv1.PolicyRule{
		Verbs:     []string{"PUT", "CREATE", "UPDATE"},
		APIGroups: []string{"v1"},
		Resources: []string{"nodes"},
	}
	ruleAdmin := rbacv1.PolicyRule{
		Verbs:     []string{"*"},
		APIGroups: []string{"*"},
		Resources: []string{"*"},
	}

	staticRoles := StaticRoles{
		roles: []*iamv1beta1.Role{
			{
				ObjectMeta: metav1.ObjectMeta{Namespace: "namespace1", Name: "readthings"},
				Rules:      []rbacv1.PolicyRule{ruleReadPods, ruleReadServices},
			},
		},
		clusterRoles: []*iamv1beta1.ClusterRole{
			{
				ObjectMeta: metav1.ObjectMeta{Name: "cluster-admin"},
				Rules:      []rbacv1.PolicyRule{ruleAdmin},
			},
			{
				ObjectMeta: metav1.ObjectMeta{Name: "write-nodes"},
				Rules:      []rbacv1.PolicyRule{ruleWriteNodes},
			},
		},
		workspaceRoles: []*iamv1beta1.WorkspaceRole{
			{
				ObjectMeta: metav1.ObjectMeta{
					Name:   "system-workspace-workspace-manager",
					Labels: map[string]string{tenantv1beta1.WorkspaceLabel: "system-workspace"},
				},
				Rules: []rbacv1.PolicyRule{ruleAdmin},
			},
		},
		globalRoles: []*iamv1beta1.GlobalRole{
			{
				ObjectMeta: metav1.ObjectMeta{
					Name: "global-admin",
				},
				Rules: []rbacv1.PolicyRule{ruleAdmin},
			},
		},

		roleBindings: []*iamv1beta1.RoleBinding{
			{
				ObjectMeta: metav1.ObjectMeta{
					Namespace: "namespace1",
					Name:      "readthings",
				},
				Subjects: []rbacv1.Subject{
					{Kind: rbacv1.UserKind, Name: "foobar"},
					{Kind: rbacv1.GroupKind, Name: "group1"},
				},
				RoleRef: rbacv1.RoleRef{APIGroup: rbacv1.GroupName, Kind: "Role", Name: "readthings"},
			},
		},
		workspaceRoleBindings: []*iamv1beta1.WorkspaceRoleBinding{
			{
				ObjectMeta: metav1.ObjectMeta{
					Name:   "system-workspace-workspace-manager-tester",
					Labels: map[string]string{tenantv1beta1.WorkspaceLabel: "system-workspace"},
				},
				RoleRef: rbacv1.RoleRef{
					APIGroup: iamv1beta1.SchemeGroupVersion.Group,
					Kind:     iamv1beta1.ResourceKindWorkspaceRole,
					Name:     "system-workspace-workspace-manager",
				},
				Subjects: []rbacv1.Subject{
					{
						Kind:     iamv1beta1.ResourceKindUser,
						APIGroup: iamv1beta1.SchemeGroupVersion.Group,
						Name:     "tester",
					},
				},
			},
		},
		globalRoleBindings: []*iamv1beta1.GlobalRoleBinding{
			{
				ObjectMeta: metav1.ObjectMeta{
					Name: "admin",
				},
				RoleRef: rbacv1.RoleRef{
					APIGroup: iamv1beta1.SchemeGroupVersion.Group,
					Kind:     iamv1beta1.ResourceKindGlobalRole,
					Name:     "global-admin",
				},
				Subjects: []rbacv1.Subject{
					{
						Kind:     iamv1beta1.ResourceKindUser,
						APIGroup: iamv1beta1.SchemeGroupVersion.Group,
						Name:     "admin",
					},
				},
			},
		},
	}

	tests := []struct {
		StaticRoles

		// For a given context, what are the rules that apply?
		user           user.Info
		namespace      string
		workspace      string
		effectiveRules []rbacv1.PolicyRule
	}{
		{
			StaticRoles:    staticRoles,
			user:           &user.DefaultInfo{Name: "admin"},
			workspace:      "system-workspace",
			effectiveRules: []rbacv1.PolicyRule{ruleAdmin},
		},
		{
			StaticRoles:    staticRoles,
			user:           &user.DefaultInfo{Name: "admin"},
			namespace:      "namespace1",
			effectiveRules: []rbacv1.PolicyRule{ruleAdmin},
		},
		{
			StaticRoles:    staticRoles,
			user:           &user.DefaultInfo{Name: "tester"},
			workspace:      "system-workspace",
			effectiveRules: []rbacv1.PolicyRule{ruleAdmin},
		},
		{
			StaticRoles:    staticRoles,
			user:           &user.DefaultInfo{Name: "tester"},
			workspace:      "not-exists-workspace",
			effectiveRules: nil,
		},
		{
			StaticRoles:    staticRoles,
			user:           &user.DefaultInfo{Name: "foobar"},
			namespace:      "namespace1",
			effectiveRules: []rbacv1.PolicyRule{ruleReadPods, ruleReadServices},
		},
		{
			StaticRoles:    staticRoles,
			user:           &user.DefaultInfo{Name: "foobar"},
			namespace:      "namespace2",
			effectiveRules: nil,
		},

		{
			StaticRoles: staticRoles,
			// Same as above but without a namespace. Only global rules should apply.
			user:           &user.DefaultInfo{Name: "foobar"},
			effectiveRules: nil,
		},
		{
			StaticRoles:    staticRoles,
			user:           &user.DefaultInfo{},
			effectiveRules: nil,
		},
	}

	for i, tc := range tests {
		ruleResolver, err := newMockRBACAuthorizer(&tc.StaticRoles)
		if err != nil {
			t.Fatal(err)
		}

		scope := request.ClusterScope

		if tc.workspace != "" {
			scope = request.WorkspaceScope
		}

		if tc.namespace != "" {
			scope = request.NamespaceScope
		}

		rules, err := ruleResolver.rulesFor(authorizer.AttributesRecord{
			User:            tc.user,
			Namespace:       tc.namespace,
			Workspace:       tc.workspace,
			ResourceScope:   scope,
			ResourceRequest: true,
		})

		if err != nil {
			t.Errorf("case %d: GetEffectivePolicyRules(context)=%v", i, err)
			continue
		}

		// Sort for deep equals
		sort.Sort(byHash(rules))
		sort.Sort(byHash(tc.effectiveRules))

		if diff := cmp.Diff(rules, tc.effectiveRules); diff != "" {
			t.Errorf("case %d: %s", i, diff)
		}
	}
}

func TestRBACAuthorizerMakeDecision(t *testing.T) {
	t.Skipf("TODO: refactor this test case")
	staticRoles := StaticRoles{
		roles: []*iamv1beta1.Role{
			{
				ObjectMeta: metav1.ObjectMeta{
					Namespace: "kubesphere-system",
					Name:      "kubesphere-system-admin",
				},
				Rules: []rbacv1.PolicyRule{
					{
						Verbs:     []string{"*"},
						APIGroups: []string{"*"},
						Resources: []string{"*"},
					},
				},
			},
			{
				ObjectMeta: metav1.ObjectMeta{
					Namespace: "kubesphere-system",
					Name:      "kubesphere-system-viewer",
				},
				Rules: []rbacv1.PolicyRule{
					{
						Verbs:     []string{"get", "list", "watch"},
						APIGroups: []string{"*"},
						Resources: []string{"*"},
					},
				},
			},
		},
		clusterRoles: []*iamv1beta1.ClusterRole{
			{
				ObjectMeta: metav1.ObjectMeta{
					Name: "cluster-viewer",
				},
				Rules: []rbacv1.PolicyRule{
					{
						Verbs:     []string{"get", "list", "watch"},
						APIGroups: []string{"*"},
						Resources: []string{"*"},
					},
				},
			},
			{
				ObjectMeta: metav1.ObjectMeta{
					Name: "cluster-admin",
				},
				Rules: []rbacv1.PolicyRule{
					{
						Verbs:     []string{"*"},
						APIGroups: []string{"*"},
						Resources: []string{"*"},
					},
				},
			},
		},
		workspaceRoles: []*iamv1beta1.WorkspaceRole{
			{
				ObjectMeta: metav1.ObjectMeta{
					Name:   "system-workspace-admin",
					Labels: map[string]string{tenantv1beta1.WorkspaceLabel: "system-workspace"},
				},
				Rules: []rbacv1.PolicyRule{
					{
						Verbs:     []string{"*"},
						APIGroups: []string{"*"},
						Resources: []string{"*"},
					},
				},
			},
			{
				ObjectMeta: metav1.ObjectMeta{
					Name:   "system-workspace-viewer",
					Labels: map[string]string{tenantv1beta1.WorkspaceLabel: "system-workspace"},
				},
				Rules: []rbacv1.PolicyRule{
					{
						Verbs:     []string{"get", "list", "watch"},
						APIGroups: []string{"*"},
						Resources: []string{"*"},
					},
				},
			},
		},
		globalRoles: []*iamv1beta1.GlobalRole{
			{
				ObjectMeta: metav1.ObjectMeta{
					Name: "global-admin",
				},
				Rules: []rbacv1.PolicyRule{
					{
						Verbs:           []string{"*"},
						APIGroups:       []string{"*"},
						Resources:       []string{"*"},
						NonResourceURLs: []string{"*"},
					},
				},
			},
			{
				ObjectMeta: metav1.ObjectMeta{
					Name: "global-viewer",
				},
				Rules: []rbacv1.PolicyRule{
					{
						Verbs:     []string{"get", "list", "watch"},
						APIGroups: []string{"*"},
						Resources: []string{"*"},
					},
				},
			},
		},

		roleBindings: []*iamv1beta1.RoleBinding{
			{
				ObjectMeta: metav1.ObjectMeta{
					Namespace: "kubesphere-system",
					Name:      "kubesphere-system-admin",
				},
				Subjects: []rbacv1.Subject{
					{Kind: rbacv1.UserKind, Name: "kubesphere-system-admin"},
				},
				RoleRef: rbacv1.RoleRef{APIGroup: rbacv1.GroupName, Kind: "Role", Name: "kubesphere-system-admin"},
			},
			{
				ObjectMeta: metav1.ObjectMeta{
					Namespace: "kubesphere-system",
					Name:      "kubesphere-system-viewer",
				},
				Subjects: []rbacv1.Subject{
					{
						Kind: rbacv1.UserKind,
						Name: "kubesphere-system-viewer",
					},
				},
				RoleRef: rbacv1.RoleRef{APIGroup: rbacv1.GroupName, Kind: "Role", Name: "kubesphere-system-viewer"},
			},
		},
		workspaceRoleBindings: []*iamv1beta1.WorkspaceRoleBinding{
			{
				ObjectMeta: metav1.ObjectMeta{
					Name:   "system-workspace-admin",
					Labels: map[string]string{tenantv1beta1.WorkspaceLabel: "system-workspace"},
				},
				RoleRef: rbacv1.RoleRef{
					APIGroup: iamv1beta1.SchemeGroupVersion.Group,
					Kind:     iamv1beta1.ResourceKindWorkspaceRole,
					Name:     "system-workspace-admin",
				},
				Subjects: []rbacv1.Subject{
					{
						Kind: iamv1beta1.ResourceKindUser,
						Name: "system-workspace-admin",
					},
				},
			},
			{
				ObjectMeta: metav1.ObjectMeta{
					Name:   "system-workspace-viewer",
					Labels: map[string]string{tenantv1beta1.WorkspaceLabel: "system-workspace"},
				},
				RoleRef: rbacv1.RoleRef{
					APIGroup: iamv1beta1.SchemeGroupVersion.Group,
					Kind:     iamv1beta1.ResourceKindWorkspaceRole,
					Name:     "system-workspace-viewer",
				},
				Subjects: []rbacv1.Subject{
					{
						Kind: iamv1beta1.ResourceKindUser,
						Name: "system-workspace-viewer",
					},
				},
			},
		},
		clusterRoleBindings: []*iamv1beta1.ClusterRoleBinding{
			{
				ObjectMeta: metav1.ObjectMeta{
					Name: "cluster-admin",
				},
				Subjects: []rbacv1.Subject{
					{Kind: rbacv1.UserKind, Name: "cluster-admin"},
				},
				RoleRef: rbacv1.RoleRef{APIGroup: rbacv1.GroupName, Kind: "Role", Name: "cluster-admin"},
			},
			{
				ObjectMeta: metav1.ObjectMeta{
					Name: "cluster-viewer",
				},
				Subjects: []rbacv1.Subject{
					{Kind: rbacv1.UserKind, Name: "cluster-viewer"},
				},
				RoleRef: rbacv1.RoleRef{APIGroup: rbacv1.GroupName, Kind: "ClusterRole", Name: "cluster-viewer"},
			},
		},
		globalRoleBindings: []*iamv1beta1.GlobalRoleBinding{
			{
				ObjectMeta: metav1.ObjectMeta{
					Name: "admin",
				},
				RoleRef: rbacv1.RoleRef{
					APIGroup: iamv1beta1.SchemeGroupVersion.Group,
					Kind:     iamv1beta1.ResourceKindGlobalRole,
					Name:     "global-admin",
				},
				Subjects: []rbacv1.Subject{
					{
						Kind:     iamv1beta1.ResourceKindUser,
						APIGroup: iamv1beta1.SchemeGroupVersion.Group,
						Name:     "admin",
					},
				},
			},
			{
				ObjectMeta: metav1.ObjectMeta{
					Name: "viewer",
				},
				RoleRef: rbacv1.RoleRef{
					APIGroup: iamv1beta1.SchemeGroupVersion.Group,
					Kind:     iamv1beta1.ResourceKindGlobalRole,
					Name:     "global-viewer",
				},
				Subjects: []rbacv1.Subject{
					{
						Kind:     iamv1beta1.ResourceKindUser,
						APIGroup: iamv1beta1.SchemeGroupVersion.Group,
						Name:     "viewer",
					},
				},
			},
		},

		namespaces: []*corev1.Namespace{
			{
				ObjectMeta: metav1.ObjectMeta{
					Name:   "kubesphere-system",
					Labels: map[string]string{tenantv1beta1.WorkspaceLabel: "system-workspace"},
				},
			},
			{
				ObjectMeta: metav1.ObjectMeta{
					Name:   "kube-system",
					Labels: map[string]string{tenantv1beta1.WorkspaceLabel: "system-workspace"},
				},
			},
		},
	}

	tests := []struct {
		StaticRoles
		Request          authorizer.AttributesRecord
		ExpectedDecision authorizer.Decision
	}{
		{
			StaticRoles: staticRoles,
			Request: authorizer.AttributesRecord{
				User: &user.DefaultInfo{
					Name: "admin",
				},
				Verb:            "create",
				APIGroup:        "",
				APIVersion:      "v1",
				Resource:        "namespaces",
				ResourceRequest: true,
				ResourceScope:   request.ClusterScope,
			},
			ExpectedDecision: authorizer.DecisionAllow,
		},
		{
			StaticRoles: staticRoles,
			Request: authorizer.AttributesRecord{
				User: &user.DefaultInfo{
					Name: "viewer",
				},
				Verb:            "create",
				APIGroup:        "",
				APIVersion:      "v1",
				Resource:        "namespaces",
				ResourceRequest: true,
				ResourceScope:   request.ClusterScope,
			},
			ExpectedDecision: authorizer.DecisionNoOpinion,
		},
		{
			StaticRoles: staticRoles,
			Request: authorizer.AttributesRecord{
				User: &user.DefaultInfo{
					Name: "viewer",
				},
				Verb:            "list",
				APIGroup:        "",
				APIVersion:      "v1",
				Resource:        "namespaces",
				ResourceRequest: true,
				ResourceScope:   request.ClusterScope,
			},
			ExpectedDecision: authorizer.DecisionAllow,
		},
		{
			StaticRoles: staticRoles,
			Request: authorizer.AttributesRecord{
				User: &user.DefaultInfo{
					Name: "admin",
				},
				Verb:            "list",
				Workspace:       "system-workspace",
				APIGroup:        "tenant.kubesphere.io",
				APIVersion:      "v1alpha2",
				Resource:        "namespaces",
				ResourceRequest: true,
				ResourceScope:   request.WorkspaceScope,
			},
			ExpectedDecision: authorizer.DecisionAllow,
		},
		{
			StaticRoles: staticRoles,
			Request: authorizer.AttributesRecord{
				User: &user.DefaultInfo{
					Name: "system-workspace-admin",
				},
				Verb:            "list",
				Workspace:       "system-workspace",
				APIGroup:        "tenant.kubesphere.io",
				APIVersion:      "v1alpha2",
				Resource:        "namespaces",
				ResourceRequest: true,
				ResourceScope:   request.WorkspaceScope,
			},
			ExpectedDecision: authorizer.DecisionAllow,
		},
		{
			StaticRoles: staticRoles,
			Request: authorizer.AttributesRecord{
				User: &user.DefaultInfo{
					Name: "system-workspace-viewer",
				},
				Verb:            "list",
				Workspace:       "system-workspace",
				APIGroup:        "tenant.kubesphere.io",
				APIVersion:      "v1alpha2",
				Resource:        "namespaces",
				ResourceRequest: true,
				ResourceScope:   request.WorkspaceScope,
			},
			ExpectedDecision: authorizer.DecisionAllow,
		},
		{
			StaticRoles: staticRoles,
			Request: authorizer.AttributesRecord{
				User: &user.DefaultInfo{
					Name: "admin",
				},
				Verb:            "create",
				Workspace:       "system-workspace",
				APIGroup:        "tenant.kubesphere.io",
				APIVersion:      "v1alpha2",
				Resource:        "namespaces",
				ResourceRequest: true,
				ResourceScope:   iamv1beta1.ScopeWorkspace,
			},
			ExpectedDecision: authorizer.DecisionAllow,
		},
		{
			StaticRoles: staticRoles,
			Request: authorizer.AttributesRecord{
				User: &user.DefaultInfo{
					Name: "system-workspace-admin",
				},
				Verb:            "create",
				Workspace:       "system-workspace",
				APIGroup:        "tenant.kubesphere.io",
				APIVersion:      "v1alpha2",
				Resource:        "namespaces",
				ResourceRequest: true,
				ResourceScope:   request.WorkspaceScope,
			},
			ExpectedDecision: authorizer.DecisionAllow,
		},
		{
			StaticRoles: staticRoles,
			Request: authorizer.AttributesRecord{
				User: &user.DefaultInfo{
					Name: "system-workspace-viewer",
				},
				Verb:            "create",
				Workspace:       "system-workspace",
				APIGroup:        "tenant.kubesphere.io",
				APIVersion:      "v1alpha2",
				Resource:        "namespaces",
				ResourceRequest: true,
				ResourceScope:   request.WorkspaceScope,
			},
			ExpectedDecision: authorizer.DecisionNoOpinion,
		},
		{
			StaticRoles: staticRoles,
			Request: authorizer.AttributesRecord{
				User: &user.DefaultInfo{
					Name: "admin",
				},
				Verb:            "create",
				APIGroup:        "apps",
				APIVersion:      "v1",
				Resource:        "deployments",
				Namespace:       "kubesphere-system",
				ResourceRequest: true,
				ResourceScope:   request.NamespaceScope,
			},
			ExpectedDecision: authorizer.DecisionAllow,
		},
		{
			StaticRoles: staticRoles,
			Request: authorizer.AttributesRecord{
				User: &user.DefaultInfo{
					Name: "viewer",
				},
				Verb:            "create",
				APIGroup:        "apps",
				APIVersion:      "v1",
				Resource:        "deployments",
				Namespace:       "kubesphere-system",
				ResourceRequest: true,
				ResourceScope:   request.NamespaceScope,
			},
			ExpectedDecision: authorizer.DecisionNoOpinion,
		},
		{
			StaticRoles: staticRoles,
			Request: authorizer.AttributesRecord{
				User: &user.DefaultInfo{
					Name: "system-workspace-admin",
				},
				Verb:            "create",
				APIGroup:        "apps",
				APIVersion:      "v1",
				Resource:        "deployments",
				Namespace:       "kubesphere-system",
				ResourceRequest: true,
				ResourceScope:   request.NamespaceScope,
			},
			ExpectedDecision: authorizer.DecisionAllow,
		},
		{
			StaticRoles: staticRoles,
			Request: authorizer.AttributesRecord{
				User: &user.DefaultInfo{
					Name: "system-workspace-viewer",
				},
				Verb:            "create",
				APIGroup:        "apps",
				APIVersion:      "v1",
				Resource:        "deployments",
				Namespace:       "kubesphere-system",
				ResourceRequest: true,
				ResourceScope:   request.NamespaceScope,
			},
			ExpectedDecision: authorizer.DecisionNoOpinion,
		},
		{
			StaticRoles: staticRoles,
			Request: authorizer.AttributesRecord{
				User: &user.DefaultInfo{
					Name: "kubesphere-system-admin",
				},
				Verb:            "create",
				APIGroup:        "apps",
				APIVersion:      "v1",
				Resource:        "deployments",
				Namespace:       "kubesphere-system",
				ResourceRequest: true,
				ResourceScope:   request.NamespaceScope,
			},
			ExpectedDecision: authorizer.DecisionAllow,
		},
		{
			StaticRoles: staticRoles,
			Request: authorizer.AttributesRecord{
				User: &user.DefaultInfo{
					Name: "kubesphere-system-viewer",
				},
				Verb:            "create",
				APIGroup:        "apps",
				APIVersion:      "v1",
				Resource:        "deployments",
				Namespace:       "kubesphere-system",
				ResourceRequest: true,
				ResourceScope:   request.NamespaceScope,
			},
			ExpectedDecision: authorizer.DecisionNoOpinion,
		},
		{
			StaticRoles: staticRoles,
			Request: authorizer.AttributesRecord{
				User: &user.DefaultInfo{
					Name: "system-workspace-admin",
				},
				Verb:            "create",
				APIGroup:        "apps",
				APIVersion:      "v1",
				Resource:        "deployments",
				Namespace:       "kube-system",
				ResourceRequest: true,
				ResourceScope:   request.NamespaceScope,
			},
			ExpectedDecision: authorizer.DecisionAllow,
		},
		{
			StaticRoles: staticRoles,
			Request: authorizer.AttributesRecord{
				User: &user.DefaultInfo{
					Name: "kubesphere-system-admin",
				},
				Verb:            "create",
				APIGroup:        "apps",
				APIVersion:      "v1",
				Resource:        "deployments",
				Namespace:       "kube-system",
				ResourceRequest: true,
				ResourceScope:   request.NamespaceScope,
			},
			ExpectedDecision: authorizer.DecisionNoOpinion,
		},
	}

	for i, tc := range tests {
		ruleResolver, err := newMockRBACAuthorizer(&tc.StaticRoles)
		if err != nil {
			t.Fatal(err)
		}

		decision, message, err := ruleResolver.Authorize(&tc.Request)

		if err != nil {
			t.Errorf("case %d: %v: %s", i, err, message)
			continue
		}

		if decision != tc.ExpectedDecision {
			t.Errorf("case %d: %d != %d", i, decision, tc.ExpectedDecision)
		}
	}
}

func newMockRBACAuthorizer(staticRoles *StaticRoles) (*Authorizer, error) {
	client := runtimefakeclient.NewClientBuilder().
		WithScheme(scheme.Scheme).Build()

	for _, role := range staticRoles.roles {
		if err := client.Create(context.Background(), role.DeepCopy()); err != nil {
			return nil, err
		}
	}

	for _, roleBinding := range staticRoles.roleBindings {
		if err := client.Create(context.Background(), roleBinding.DeepCopy()); err != nil {
			return nil, err
		}
	}

	for _, clusterRole := range staticRoles.clusterRoles {
		if err := client.Create(context.Background(), clusterRole.DeepCopy()); err != nil {
			return nil, err
		}
	}

	for _, clusterRoleBinding := range staticRoles.clusterRoleBindings {
		if err := client.Create(context.Background(), clusterRoleBinding.DeepCopy()); err != nil {
			return nil, err
		}
	}

	for _, workspaceRole := range staticRoles.workspaceRoles {
		if err := client.Create(context.Background(), workspaceRole.DeepCopy()); err != nil {
			return nil, err
		}
	}

	for _, workspaceRoleBinding := range staticRoles.workspaceRoleBindings {
		if err := client.Create(context.Background(), workspaceRoleBinding.DeepCopy()); err != nil {
			return nil, err
		}
	}

	for _, globalRole := range staticRoles.globalRoles {
		if err := client.Create(context.Background(), globalRole.DeepCopy()); err != nil {
			return nil, err
		}
	}

	for _, globalRoleBinding := range staticRoles.globalRoleBindings {
		if err := client.Create(context.Background(), globalRoleBinding.DeepCopy()); err != nil {
			return nil, err
		}
	}

	fakeCache := &informertest.FakeInformers{Scheme: scheme.Scheme}

	resourceManager, err := v1beta1.New(context.Background(), client, fakeCache)
	if err != nil {
		return nil, err
	}

	return NewRBACAuthorizer(am.NewReadOnlyOperator(resourceManager)), nil
}

func TestAppliesTo(t *testing.T) {
	tests := []struct {
		subjects  []rbacv1.Subject
		user      user.Info
		namespace string
		appliesTo bool
		index     int
		testCase  string
	}{
		{
			subjects: []rbacv1.Subject{
				{Kind: rbacv1.UserKind, Name: "foobar"},
			},
			user:      &user.DefaultInfo{Name: "foobar"},
			appliesTo: true,
			index:     0,
			testCase:  "single subject that matches username",
		},
		{
			subjects: []rbacv1.Subject{
				{Kind: rbacv1.UserKind, Name: "barfoo"},
				{Kind: rbacv1.UserKind, Name: "foobar"},
			},
			user:      &user.DefaultInfo{Name: "foobar"},
			appliesTo: true,
			index:     1,
			testCase:  "multiple subjects, one that matches username",
		},
		{
			subjects: []rbacv1.Subject{
				{Kind: rbacv1.UserKind, Name: "barfoo"},
				{Kind: rbacv1.UserKind, Name: "foobar"},
			},
			user:      &user.DefaultInfo{Name: "zimzam"},
			appliesTo: false,
			testCase:  "multiple subjects, none that match username",
		},
		{
			subjects: []rbacv1.Subject{
				{Kind: rbacv1.UserKind, Name: "barfoo"},
				{Kind: rbacv1.GroupKind, Name: "foobar"},
			},
			user:      &user.DefaultInfo{Name: "zimzam", Groups: []string{"foobar"}},
			appliesTo: true,
			index:     1,
			testCase:  "multiple subjects, one that match group",
		},
		{
			subjects: []rbacv1.Subject{
				{Kind: rbacv1.UserKind, Name: "barfoo"},
				{Kind: rbacv1.GroupKind, Name: "foobar"},
			},
			user:      &user.DefaultInfo{Name: "zimzam", Groups: []string{"foobar"}},
			namespace: "namespace1",
			appliesTo: true,
			index:     1,
			testCase:  "multiple subjects, one that match group, should ignore namespace",
		},
		{
			subjects: []rbacv1.Subject{
				{Kind: rbacv1.UserKind, Name: "barfoo"},
				{Kind: rbacv1.GroupKind, Name: "foobar"},
				{Kind: rbacv1.ServiceAccountKind, APIGroup: rbacv1.GroupName, Namespace: "kube-system", Name: "default"},
			},
			user:      &user.DefaultInfo{Name: "system:serviceaccount:kube-system:default"},
			namespace: "default",
			appliesTo: true,
			index:     2,
			testCase:  "multiple subjects with a service account that matches",
		},
		{
			subjects: []rbacv1.Subject{
				{Kind: rbacv1.UserKind, Name: "*"},
			},
			user:      &user.DefaultInfo{Name: "foobar"},
			namespace: "default",
			appliesTo: false,
			testCase:  "* user subject name doesn't match all users",
		},
		{
			subjects: []rbacv1.Subject{
				{Kind: rbacv1.GroupKind, Name: user.AllAuthenticated},
				{Kind: rbacv1.GroupKind, Name: user.AllUnauthenticated},
			},
			user:      &user.DefaultInfo{Name: "foobar", Groups: []string{user.AllAuthenticated}},
			namespace: "default",
			appliesTo: true,
			index:     0,
			testCase:  "binding to all authenticated and unauthenticated subjects matches authenticated user",
		},
		{
			subjects: []rbacv1.Subject{
				{Kind: rbacv1.GroupKind, Name: user.AllAuthenticated},
				{Kind: rbacv1.GroupKind, Name: user.AllUnauthenticated},
			},
			user:      &user.DefaultInfo{Name: "system:anonymous", Groups: []string{user.AllUnauthenticated}},
			namespace: "default",
			appliesTo: true,
			index:     1,
			testCase:  "binding to all authenticated and unauthenticated subjects matches anonymous user",
		},
	}

	for _, tc := range tests {
		gotIndex, got := appliesTo(tc.user, tc.subjects, tc.namespace)
		if got != tc.appliesTo {
			t.Errorf("case %q want appliesTo=%t, got appliesTo=%t", tc.testCase, tc.appliesTo, got)
		}
		if gotIndex != tc.index {
			t.Errorf("case %q want index %d, got %d", tc.testCase, tc.index, gotIndex)
		}
	}
}
