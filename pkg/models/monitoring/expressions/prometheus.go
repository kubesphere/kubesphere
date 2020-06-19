package expressions

import (
	"fmt"
	"github.com/prometheus/common/model"
	"github.com/prometheus/prometheus/promql"
	"github.com/prometheus/prometheus/storage/metric"
)

func init() {
	register("prometheus", labelReplace)
}

func labelReplace(input, ns string) (string, error) {
	root, err := promql.ParseExpr(input)
	if err != nil {
		return "", err
	}

	setRecursive(root, ns)
	if err != nil {
		return "", err
	}

	return root.String(), nil
}

// Inspired by https://github.com/prometheus-community/prom-label-proxy
func setRecursive(node promql.Node, namespace string) (err error) {
	switch n := node.(type) {
	case *promql.EvalStmt:
		if err := setRecursive(n.Expr, namespace); err != nil {
			return err
		}
	case promql.Expressions:
		for _, e := range n {
			if err := setRecursive(e, namespace); err != nil {
				return err
			}
		}
	case *promql.AggregateExpr:
		if err := setRecursive(n.Expr, namespace); err != nil {
			return err
		}
	case *promql.BinaryExpr:
		if err := setRecursive(n.LHS, namespace); err != nil {
			return err
		}
		if err := setRecursive(n.RHS, namespace); err != nil {
			return err
		}
	case *promql.Call:
		if err := setRecursive(n.Args, namespace); err != nil {
			return err
		}
	case *promql.ParenExpr:
		if err := setRecursive(n.Expr, namespace); err != nil {
			return err
		}
	case *promql.UnaryExpr:
		if err := setRecursive(n.Expr, namespace); err != nil {
			return err
		}
	case *promql.NumberLiteral, *promql.StringLiteral:
		// nothing to do
	case *promql.MatrixSelector:
		n.LabelMatchers = enforceLabelMatchers(n.LabelMatchers, namespace)
	case *promql.VectorSelector:
		n.LabelMatchers = enforceLabelMatchers(n.LabelMatchers, namespace)
	default:
		return fmt.Errorf("promql.Walk: unhandled node type %T", node)
	}
	return err
}

func enforceLabelMatchers(matchers metric.LabelMatchers, namespace string) metric.LabelMatchers {
	var found bool
	for i, m := range matchers {
		if m.Name == "namespace" {
			matchers[i] = &metric.LabelMatcher{
				Name:  "namespace",
				Type:  metric.Equal,
				Value: model.LabelValue(namespace),
			}
			found = true
			break
		}
	}

	if !found {
		matchers = append(matchers, &metric.LabelMatcher{
			Name:  "namespace",
			Type:  metric.Equal,
			Value: model.LabelValue(namespace),
		})
	}
	return matchers
}
