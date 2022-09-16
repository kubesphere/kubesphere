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

package alerting

import (
	"context"

	"github.com/go-logr/logr"
	promresourcesv1 "github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring/v1"
	promlabels "github.com/prometheus/prometheus/pkg/labels"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/uuid"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"sigs.k8s.io/controller-runtime/pkg/source"

	alertingv2beta1 "kubesphere.io/api/alerting/v2beta1"
)

type ClusterRuleGroupReconciler struct {
	client.Client

	Log logr.Logger
}

func (r *ClusterRuleGroupReconciler) Reconcile(ctx context.Context, req reconcile.Request) (reconcile.Result, error) {

	var (
		log = r.Log

		ruleLevel = RuleLevelCluster

		clusterrulegroupList = alertingv2beta1.ClusterRuleGroupList{}

		promruleNamespace = PrometheusRuleNamespace
	)

	// get all enabled clusterrulegroups
	err := r.Client.List(ctx, &clusterrulegroupList, &client.ListOptions{
		LabelSelector: labels.SelectorFromSet(labels.Set{
			SourceGroupResourceLabelKeyEnable: SourceGroupResourceLabelValueEnableTrue,
		}),
	})
	if err != nil {
		return reconcile.Result{}, err
	}

	// add rule_id label that may have been missed
	var updated bool
	for i := range clusterrulegroupList.Items {
		g := clusterrulegroupList.Items[i]
		for j := range g.Spec.Rules {
			if g.Spec.Rules[j].Labels == nil {
				g.Spec.Rules[j].Labels = make(map[string]string)
			}
			if _, ok := g.Spec.Rules[j].Labels[alertingv2beta1.RuleLabelKeyRuleId]; !ok {
				g.Spec.Rules[j].Labels[alertingv2beta1.RuleLabelKeyRuleId] = string(uuid.NewUUID())
				err = r.Client.Update(ctx, &g)
				if err != nil {
					return reconcile.Result{}, err
				}
				updated = true
			}
		}
	}
	if updated {
		return reconcile.Result{}, nil
	}

	// labels added to rule.labels
	enforceRuleLabels := map[string]string{
		RuleLabelKeyRuleLevel: string(ruleLevel),
	}
	// matchers enforced to rule.expr
	enforceRuleMatchers := []*promlabels.Matcher{}
	// labels added to PrometheusRule.metadata.labels
	promruleLabelSet := labels.Set{
		PrometheusRuleResourceLabelKeyRuleLevel: string(ruleLevel),
	}

	enforceFuncs := createEnforceRuleFuncs(enforceRuleMatchers, enforceRuleLabels)

	// make PrometheusRule Groups
	rulegroups, err := makePrometheusRuleGroups(log, &clusterrulegroupList, enforceFuncs...)
	if err != nil {
		return reconcile.Result{}, err
	}
	if len(rulegroups) == 0 {
		err = r.Client.DeleteAllOf(ctx, &promresourcesv1.PrometheusRule{}, &client.DeleteAllOfOptions{
			ListOptions: client.ListOptions{
				Namespace:     promruleNamespace,
				LabelSelector: labels.SelectorFromSet(promruleLabelSet),
			},
		})
		return reconcile.Result{}, err
	}

	// make desired PrometheuRule resources
	desired, err := makePrometheusRuleResources(rulegroups, promruleNamespace, PrometheusRulePrefixClusterLevel, promruleLabelSet, nil)
	if err != nil {
		return reconcile.Result{}, err
	}

	// get current PrometheusRules
	var current promresourcesv1.PrometheusRuleList
	err = r.Client.List(ctx, &current, &client.ListOptions{
		Namespace:     promruleNamespace,
		LabelSelector: labels.SelectorFromSet(promruleLabelSet),
	})
	if err != nil {
		return reconcile.Result{}, err
	}

	// update relevant prometheusrule resources
	err = bulkUpdatePrometheusRuleResources(r.Client, ctx, current.Items, desired)
	if err != nil && (apierrors.IsConflict(err) || apierrors.IsAlreadyExists(err)) {
		return reconcile.Result{Requeue: true}, nil
	}
	return reconcile.Result{}, err
}

func (r *ClusterRuleGroupReconciler) SetupWithManager(mgr ctrl.Manager) error {
	if r.Log == nil {
		r.Log = mgr.GetLogger()
	}
	if r.Client == nil {
		r.Client = mgr.GetClient()
	}

	ctr, err := controller.New("clusterrulegroup", mgr,
		controller.Options{
			Reconciler: r,
		})

	if err != nil {
		return err
	}

	err = ctr.Watch(
		&source.Kind{Type: &alertingv2beta1.ClusterRuleGroup{}},
		handler.EnqueueRequestsFromMapFunc(func(o client.Object) []reconcile.Request {
			return []reconcile.Request{{
				NamespacedName: types.NamespacedName{
					Namespace: PrometheusRuleNamespace,
				},
			}}
		}))
	return err
}
