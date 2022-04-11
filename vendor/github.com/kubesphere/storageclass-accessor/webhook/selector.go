package webhook

import (
	"github.com/kubesphere/storageclass-accessor/client/apis/accessor/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	workspacev1alpha1 "kubesphere.io/api/tenant/v1alpha1"
)

func matchLabel(info map[string]string, expressions []v1alpha1.MatchExpressions) bool {
	if len(expressions) == 0 {
		return true
	}

	for _, rule := range expressions {
		rulePass := true
		for _, item := range rule.MatchExpressions {
			switch item.Operator {
			case v1alpha1.In:
				rulePass = rulePass && inList(info[item.Key], item.Values)
			case v1alpha1.NotIn:
				rulePass = rulePass && !inList(info[item.Key], item.Values)
			}
		}
		if rulePass {
			return rulePass
		}
	}
	return false
}

func matchField(ns *corev1.Namespace, expressions []v1alpha1.FieldExpressions) bool {
	//If not set limit, default pass
	if len(expressions) == 0 {
		return true
	}

	for _, rule := range expressions {
		rulePass := true
		for _, item := range rule.FieldExpressions {
			var val string
			switch item.Field {
			case v1alpha1.Name:
				val = ns.Name
			case v1alpha1.Status:
				val = string(ns.Status.Phase)
			}
			switch item.Operator {
			case v1alpha1.In:
				rulePass = rulePass && inList(val, item.Values)
			case v1alpha1.NotIn:
				rulePass = rulePass && !inList(val, item.Values)
			}
		}
		if rulePass {
			return rulePass
		}
	}
	return false
}

func wsMatchField(ws *workspacev1alpha1.Workspace, expressions []v1alpha1.FieldExpressions) bool {
	if len(expressions) == 0 {
		return true
	}

	for _, rule := range expressions {
		pass := true
		for _, item := range rule.FieldExpressions {
			switch item.Field {
			case v1alpha1.Status:
				continue
			}
			switch item.Operator {
			case v1alpha1.In:
				pass = pass && inList(ws.Name, item.Values)
			case v1alpha1.NotIn:
				pass = pass && !inList(ws.Name, item.Values)
			}
		}
		if pass {
			return pass
		}
	}
	return false
}

func inList(val string, list []string) bool {
	for _, elements := range list {
		if val == elements {
			return true
		}
	}
	return false
}
