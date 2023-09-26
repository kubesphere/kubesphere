package webhook

import (
	corev1 "k8s.io/api/core/v1"
	"k8s.io/utils/strings/slices"
	workspacev1alpha1 "kubesphere.io/api/tenant/v1alpha1"

	"github.com/kubesphere/storageclass-accessor/client/apis/accessor/v1alpha1"
)

func matchLabel(info map[string]string, expressions []v1alpha1.MatchExpressions) bool {
	if len(expressions) == 0 {
		return true
	}

	for _, rule := range expressions {
		rulePass := true
		for _, item := range rule.MatchExpressions {
			if len(item.Values) == 0 {
				continue
			}
			var labelKeyExist bool
			_, ok := info[item.Key]
			if ok {
				labelKeyExist = true
			}
			switch item.Operator {
			case v1alpha1.In:
				rulePass = rulePass && labelKeyExist && (slices.Contains(item.Values, "*") || slices.Contains(item.Values, info[item.Key]))
			case v1alpha1.NotIn:
				if labelKeyExist {
					rulePass = rulePass && !slices.Contains(item.Values, "*") && !slices.Contains(item.Values, info[item.Key])
				}
			default:
				continue
			}
			if !rulePass {
				break
			}
		}
		if rulePass {
			return rulePass
		}
	}
	return false
}

func nsMatchField(ns *corev1.Namespace, expressions []v1alpha1.FieldExpressions) bool {
	//If not set limit, default pass
	if len(expressions) == 0 {
		return true
	}

	for _, rule := range expressions {
		rulePass := true
		for _, item := range rule.FieldExpressions {
			if len(item.Values) == 0 {
				continue
			}
			var val string
			switch item.Field {
			case v1alpha1.Name:
				val = ns.Name
			case v1alpha1.Status:
				val = string(ns.Status.Phase)
			default:
				continue
			}
			switch item.Operator {
			case v1alpha1.In:
				rulePass = rulePass && (slices.Contains(item.Values, "*") || slices.Contains(item.Values, val))
			case v1alpha1.NotIn:
				rulePass = rulePass && !slices.Contains(item.Values, "*") && !slices.Contains(item.Values, val)
			default:
				continue
			}
			if !rulePass {
				break
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
			if len(item.Values) == 0 {
				continue
			}
			var val string
			switch item.Field {
			case v1alpha1.Name:
				val = ws.Name
			case v1alpha1.Status:
				// TODO(stone): check status
				continue
			default:
				continue
			}
			switch item.Operator {
			case v1alpha1.In:
				pass = pass && (slices.Contains(item.Values, "*") || slices.Contains(item.Values, val))
			case v1alpha1.NotIn:
				pass = pass && !slices.Contains(item.Values, "*") && !slices.Contains(item.Values, val)
			default:
				continue
			}
			if !pass {
				break
			}
		}
		if pass {
			return pass
		}
	}
	return false
}
