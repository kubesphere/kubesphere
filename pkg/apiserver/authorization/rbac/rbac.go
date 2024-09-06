// NOTE: This file is copied from k8s.io/kubernetes/plugin/pkg/auth/authorizer/rbac.

package rbac

import (
	"bytes"
	"context"
	"fmt"

	"github.com/open-policy-agent/opa/rego"
	rbacv1 "k8s.io/api/rbac/v1"
	utilerrors "k8s.io/apimachinery/pkg/util/errors"
	"k8s.io/apiserver/pkg/authentication/serviceaccount"
	"k8s.io/apiserver/pkg/authentication/user"
	"k8s.io/klog/v2"

	corev1alpha1 "kubesphere.io/api/core/v1alpha1"
	iamv1beta1 "kubesphere.io/api/iam/v1beta1"

	"kubesphere.io/kubesphere/pkg/apiserver/authorization/authorizer"
	"kubesphere.io/kubesphere/pkg/apiserver/request"
	"kubesphere.io/kubesphere/pkg/models/iam/am"
	ksserviceaccount "kubesphere.io/kubesphere/pkg/utils/serviceaccount"
	"kubesphere.io/kubesphere/pkg/utils/sliceutil"
)

const (
	defaultRegoQuery    = "data.authz.allow"
	defaultRegoFileName = "authz.rego"
)

type Authorizer struct {
	am am.AccessManagementInterface
}

// authorizingVisitor short-circuits once allowed, and collects any resolution errors encountered
type authorizingVisitor struct {
	requestAttributes authorizer.Attributes

	allowed bool
	reason  string
	errors  []error
}

func (v *authorizingVisitor) visit(source fmt.Stringer, regoPolicy string, rule *rbacv1.PolicyRule, err error) bool {
	if regoPolicy != "" && regoPolicyAllows(v.requestAttributes, regoPolicy) {
		v.allowed = true
		v.reason = fmt.Sprintf("RBAC: allowed by %s", source.String())
		return false
	}
	if rule != nil && ruleAllows(v.requestAttributes, rule) {
		v.allowed = true
		v.reason = fmt.Sprintf("RBAC: allowed by %s", source.String())
		return false
	}
	if err != nil {
		v.errors = append(v.errors, err)
	}
	return true
}

type ruleAccumulator struct {
	rules  []rbacv1.PolicyRule
	errors []error
}

func (r *ruleAccumulator) visit(_ fmt.Stringer, _ string, rule *rbacv1.PolicyRule, err error) bool {
	if rule != nil {
		r.rules = append(r.rules, *rule)
	}
	if err != nil {
		r.errors = append(r.errors, err)
	}
	return true
}

func (r *Authorizer) Authorize(requestAttributes authorizer.Attributes) (authorizer.Decision, string, error) {
	ruleCheckingVisitor := &authorizingVisitor{requestAttributes: requestAttributes}

	r.visitRulesFor(requestAttributes, ruleCheckingVisitor.visit)

	if ruleCheckingVisitor.allowed {
		return authorizer.DecisionAllow, ruleCheckingVisitor.reason, nil
	}

	// Build a detailed log of the denial.
	// Make the whole block conditional, so we don't do a lot of string-building we won't use.
	if klog.V(4).Enabled() {
		var operation string
		if requestAttributes.IsResourceRequest() {
			b := &bytes.Buffer{}
			b.WriteString(`"`)
			b.WriteString(requestAttributes.GetVerb())
			b.WriteString(`" resource "`)
			b.WriteString(requestAttributes.GetResource())
			if len(requestAttributes.GetAPIGroup()) > 0 {
				b.WriteString(`.`)
				b.WriteString(requestAttributes.GetAPIGroup())
			}
			if len(requestAttributes.GetSubresource()) > 0 {
				b.WriteString(`/`)
				b.WriteString(requestAttributes.GetSubresource())
			}
			b.WriteString(`"`)
			if len(requestAttributes.GetName()) > 0 {
				b.WriteString(` named "`)
				b.WriteString(requestAttributes.GetName())
				b.WriteString(`"`)
			}
			operation = b.String()
		} else {
			operation = fmt.Sprintf("%q nonResourceURL %q", requestAttributes.GetVerb(), requestAttributes.GetPath())
		}

		var scope string
		if requestAttributes.GetResourceScope() == request.NamespaceScope {
			scope = fmt.Sprintf("in namespace %q", requestAttributes.GetNamespace())
		} else if requestAttributes.GetResourceScope() == request.WorkspaceScope {
			scope = fmt.Sprintf("in workspace %q", requestAttributes.GetWorkspace())
		} else if requestAttributes.GetResourceScope() == request.ClusterScope {
			scope = "cluster scope"
		} else {
			scope = "global-wide"
		}

		klog.V(4).Infof("RBAC: no rules authorize user %q with groups %q to %s %s", requestAttributes.GetUser().GetName(), requestAttributes.GetUser().GetGroups(), operation, scope)
	}

	reason := ""
	if len(ruleCheckingVisitor.errors) > 0 {
		reason = fmt.Sprintf("RBAC: %v", utilerrors.NewAggregate(ruleCheckingVisitor.errors))
	}
	return authorizer.DecisionNoOpinion, reason, nil
}

func NewRBACAuthorizer(am am.AccessManagementInterface) *Authorizer {
	return &Authorizer{am: am}
}

func ruleAllows(requestAttributes authorizer.Attributes, rule *rbacv1.PolicyRule) bool {
	if requestAttributes.IsResourceRequest() {
		combinedResource := requestAttributes.GetResource()
		if len(requestAttributes.GetSubresource()) > 0 {
			combinedResource = requestAttributes.GetResource() + "/" + requestAttributes.GetSubresource()
		}

		return VerbMatches(rule, requestAttributes.GetVerb()) &&
			APIGroupMatches(rule, requestAttributes.GetAPIGroup()) &&
			ResourceMatches(rule, combinedResource, requestAttributes.GetSubresource()) &&
			ResourceNameMatches(rule, requestAttributes.GetName())
	}

	return VerbMatches(rule, requestAttributes.GetVerb()) &&
		NonResourceURLMatches(rule, requestAttributes.GetPath())
}

func regoPolicyAllows(requestAttributes authorizer.Attributes, regoPolicy string) bool {
	// Call the rego.New function to create an object that can be prepared or evaluated
	//  After constructing a new rego.Rego object you can call PrepareForEval() to obtain an executable query
	query, err := rego.New(rego.Query(defaultRegoQuery), rego.Module(defaultRegoFileName, regoPolicy)).PrepareForEval(context.Background())

	if err != nil {
		klog.Warningf("syntax error:%s, content: %s", err, regoPolicy)
		return false
	}

	// The policy decision is contained in the results returned by the Eval() call. You can inspect the decision and handle it accordingly.
	results, err := query.Eval(context.Background(), rego.EvalInput(requestAttributes))

	if err != nil {
		klog.Warningf("syntax error:%s, content: %s", err, regoPolicy)
		return false
	}

	if len(results) > 0 && results[0].Expressions[0].Value == true {
		return true
	}

	return false
}

func (r *Authorizer) rulesFor(requestAttributes authorizer.Attributes) ([]rbacv1.PolicyRule, error) {
	visitor := &ruleAccumulator{}
	r.visitRulesFor(requestAttributes, visitor.visit)
	return visitor.rules, utilerrors.NewAggregate(visitor.errors)
}

func (r *Authorizer) visitRulesFor(requestAttributes authorizer.Attributes, visitor func(source fmt.Stringer, regoPolicy string, rule *rbacv1.PolicyRule, err error) bool) {
	if globalRoleBindings, err := r.am.ListGlobalRoleBindings("", ""); err != nil {
		visitor(nil, "", nil, err)
		return
	} else {
		sourceDescriber := &globalRoleBindingDescriber{}
		for _, globalRoleBinding := range globalRoleBindings {
			subjectIndex, applies := appliesTo(requestAttributes.GetUser(), globalRoleBinding.Subjects, "")
			if !applies {
				continue
			}
			regoPolicy, rules, err := r.am.GetRoleReferenceRules(globalRoleBinding.RoleRef, "")
			if err != nil {
				visitor(nil, "", nil, err)
				continue
			}
			sourceDescriber.binding = &globalRoleBinding
			sourceDescriber.subject = &globalRoleBinding.Subjects[subjectIndex]
			if !visitor(sourceDescriber, regoPolicy, nil, nil) {
				return
			}
			for i := range rules {
				if !visitor(sourceDescriber, "", &rules[i], nil) {
					return
				}
			}
		}

		if requestAttributes.GetResourceScope() == request.GlobalScope {
			return
		}
	}

	var targetWorkspace string
	if requestAttributes.GetResourceScope() == request.NamespaceScope {
		if workspace, err := r.am.GetNamespaceControlledWorkspace(requestAttributes.GetNamespace()); err != nil {
			visitor(nil, "", nil, err)
			return
		} else {
			targetWorkspace = workspace
		}
	}

	if requestAttributes.GetResourceScope() == request.WorkspaceScope {
		targetWorkspace = requestAttributes.GetWorkspace()
	}

	// workspace managed resources
	if targetWorkspace != "" {
		if workspaceRoleBindings, err := r.am.ListWorkspaceRoleBindings("", "", nil, targetWorkspace); err != nil {
			visitor(nil, "", nil, err)
			return
		} else {
			sourceDescriber := &workspaceRoleBindingDescriber{}
			for _, workspaceRoleBinding := range workspaceRoleBindings {
				subjectIndex, applies := appliesTo(requestAttributes.GetUser(), workspaceRoleBinding.Subjects, "")
				if !applies {
					continue
				}
				regoPolicy, rules, err := r.am.GetRoleReferenceRules(workspaceRoleBinding.RoleRef, "")
				if err != nil {
					visitor(nil, "", nil, err)
					return
				}
				sourceDescriber.binding = &workspaceRoleBinding
				sourceDescriber.subject = &workspaceRoleBinding.Subjects[subjectIndex]
				if !visitor(sourceDescriber, regoPolicy, nil, nil) {
					return
				}
				for i := range rules {
					if !visitor(sourceDescriber, "", &rules[i], nil) {
						return
					}
				}
			}
		}
	}

	var targetNamespace string
	if requestAttributes.GetResourceScope() == request.NamespaceScope {
		targetNamespace = requestAttributes.GetNamespace()
	}

	if targetNamespace != "" {
		if roleBindings, err := r.am.ListRoleBindings("", "", nil, targetNamespace); err != nil {
			visitor(nil, "", nil, err)
			return
		} else {
			sourceDescriber := &roleBindingDescriber{}
			for _, roleBinding := range roleBindings {
				subjectIndex, applies := appliesTo(requestAttributes.GetUser(), roleBinding.Subjects, targetNamespace)
				if !applies {
					continue
				}
				regoPolicy, rules, err := r.am.GetRoleReferenceRules(roleBinding.RoleRef, targetNamespace)
				if err != nil {
					visitor(nil, "", nil, err)
					return
				}
				sourceDescriber.binding = &roleBinding
				sourceDescriber.subject = &roleBinding.Subjects[subjectIndex]
				if !visitor(sourceDescriber, regoPolicy, nil, nil) {
					return
				}
				for i := range rules {
					if !visitor(sourceDescriber, "", &rules[i], nil) {
						return
					}
				}
			}
		}
	}

	if clusterRoleBindings, err := r.am.ListClusterRoleBindings("", ""); err != nil {
		visitor(nil, "", nil, err)
		return
	} else {
		sourceDescriber := &clusterRoleBindingDescriber{}
		for _, clusterRoleBinding := range clusterRoleBindings {
			subjectIndex, applies := appliesTo(requestAttributes.GetUser(), clusterRoleBinding.Subjects, "")
			if !applies {
				continue
			}
			regoPolicy, rules, err := r.am.GetRoleReferenceRules(clusterRoleBinding.RoleRef, "")
			if err != nil {
				visitor(nil, "", nil, err)
				return
			}
			sourceDescriber.binding = &clusterRoleBinding
			sourceDescriber.subject = &clusterRoleBinding.Subjects[subjectIndex]
			if !visitor(sourceDescriber, regoPolicy, nil, nil) {
				return
			}
			for i := range rules {
				if !visitor(sourceDescriber, "", &rules[i], nil) {
					return
				}
			}
		}
	}
}

// appliesTo returns whether any of the bindingSubjects applies to the specified subject,
// and if true, the index of the first subject that applies
func appliesTo(user user.Info, bindingSubjects []rbacv1.Subject, namespace string) (int, bool) {
	for i, bindingSubject := range bindingSubjects {
		if appliesToUser(user, bindingSubject, namespace) {
			return i, true
		}
	}
	return 0, false
}

func appliesToUser(user user.Info, subject rbacv1.Subject, namespace string) bool {
	switch subject.Kind {
	case rbacv1.UserKind:
		return user.GetName() == subject.Name

	case rbacv1.GroupKind:
		return sliceutil.HasString(user.GetGroups(), subject.Name)

	case rbacv1.ServiceAccountKind:
		// Default the namespace to namespace we're working in if it's available.
		// This allows role bindings that reference
		// SAs in the local namespace to avoid having to qualify them.
		saNamespace := namespace
		if len(subject.Namespace) > 0 {
			saNamespace = subject.Namespace
		}
		if len(saNamespace) == 0 {
			return false
		}
		switch subject.APIGroup {
		case rbacv1.GroupName:
			// use a more efficient comparison for RBAC checking
			return serviceaccount.MatchesUsername(saNamespace, subject.Name, user.GetName())
		case corev1alpha1.GroupName:
			return ksserviceaccount.MatchesUsername(saNamespace, subject.Name, user.GetName())
		default:
			return false
		}

	default:
		return false
	}
}

type globalRoleBindingDescriber struct {
	binding *iamv1beta1.GlobalRoleBinding
	subject *rbacv1.Subject
}

func (d *globalRoleBindingDescriber) String() string {
	return fmt.Sprintf("GlobalRoleBinding %q of %s %q to %s",
		d.binding.Name,
		d.binding.RoleRef.Kind,
		d.binding.RoleRef.Name,
		describeSubject(d.subject, ""),
	)
}

type clusterRoleBindingDescriber struct {
	binding *iamv1beta1.ClusterRoleBinding
	subject *rbacv1.Subject
}

func (d *clusterRoleBindingDescriber) String() string {
	return fmt.Sprintf("ClusterRoleBinding %q of %s %q to %s",
		d.binding.Name,
		d.binding.RoleRef.Kind,
		d.binding.RoleRef.Name,
		describeSubject(d.subject, ""),
	)
}

type workspaceRoleBindingDescriber struct {
	binding *iamv1beta1.WorkspaceRoleBinding
	subject *rbacv1.Subject
}

func (d *workspaceRoleBindingDescriber) String() string {
	return fmt.Sprintf("GlobalRoleBinding %q of %s %q to %s",
		d.binding.Name,
		d.binding.RoleRef.Kind,
		d.binding.RoleRef.Name,
		describeSubject(d.subject, ""),
	)
}

type roleBindingDescriber struct {
	binding *iamv1beta1.RoleBinding
	subject *rbacv1.Subject
}

func (d *roleBindingDescriber) String() string {
	return fmt.Sprintf("RoleBinding %q of %s %q to %s",
		d.binding.Name+"/"+d.binding.Namespace,
		d.binding.RoleRef.Kind,
		d.binding.RoleRef.Name,
		describeSubject(d.subject, d.binding.Namespace),
	)
}

func describeSubject(s *rbacv1.Subject, bindingNamespace string) string {
	switch s.Kind {
	case rbacv1.ServiceAccountKind:
		if len(s.Namespace) > 0 {
			return fmt.Sprintf("%s %q", s.Kind, s.Name+"/"+s.Namespace)
		}
		return fmt.Sprintf("%s %q", s.Kind, s.Name+"/"+bindingNamespace)
	default:
		return fmt.Sprintf("%s %q", s.Kind, s.Name)
	}
}
