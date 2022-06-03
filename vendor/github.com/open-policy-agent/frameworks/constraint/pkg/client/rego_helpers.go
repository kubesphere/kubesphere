package client

import (
	"fmt"

	"github.com/open-policy-agent/opa/ast"
	"github.com/pkg/errors"
)

var (
	// Currently rules should only access data.inventory
	validDataFields = map[string]bool{
		"inventory": true,
	}
)

// parseModule parses the module and also fails empty modules.
func parseModule(path, rego string) (*ast.Module, error) {
	module, err := ast.ParseModule(path, rego)
	if err != nil {
		return nil, err
	}
	if module == nil {
		return nil, errors.New("Empty module")
	}
	return module, nil
}

// rewriteModulePackage rewrites the module's package path to path.
func rewriteModulePackage(path string, module *ast.Module) error {
	pathParts, err := ast.ParseRef(path)
	if err != nil {
		return err
	}
	packageRef := ast.Ref([]*ast.Term{ast.VarTerm("data")})
	newPath := packageRef.Extend(pathParts)
	module.Package.Path = newPath
	return nil
}

// requireRulesModule makes sure the listed rules are specified
func requireRulesModule(module *ast.Module, requiredRules map[string]struct{}) error {
	ruleSets := make(map[string]struct{}, len(module.Rules))
	for _, rule := range module.Rules {
		ruleSets[string(rule.Head.Name)] = struct{}{}
	}

	var errs Errors
	for name := range requiredRules {
		_, ok := ruleSets[name]
		if !ok {
			errs = append(errs, fmt.Errorf("Missing required rule: %s", name))
			continue
		}
	}
	if len(errs) != 0 {
		return errs
	}
	return nil
}
