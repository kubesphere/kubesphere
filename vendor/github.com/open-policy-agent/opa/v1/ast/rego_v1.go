package ast

import (
	"fmt"

	"github.com/open-policy-agent/opa/v1/ast/internal/tokens"
)

func checkDuplicateImports(modules []*Module) (errors Errors) {
	for _, module := range modules {
		processedImports := map[Var]*Import{}

		for _, imp := range module.Imports {
			name := imp.Name()

			if processed, conflict := processedImports[name]; conflict {
				errors = append(errors, NewError(CompileErr, imp.Location, "import must not shadow %v", processed))
			} else {
				processedImports[name] = imp
			}
		}
	}
	return
}

func checkRootDocumentOverrides(node interface{}) Errors {
	errors := Errors{}

	WalkRules(node, func(rule *Rule) bool {
		var name string
		if len(rule.Head.Reference) > 0 {
			name = rule.Head.Reference[0].Value.(Var).String()
		} else {
			name = rule.Head.Name.String()
		}
		if RootDocumentRefs.Contains(RefTerm(VarTerm(name))) {
			errors = append(errors, NewError(CompileErr, rule.Location, "rules must not shadow %v (use a different rule name)", name))
		}

		for _, arg := range rule.Head.Args {
			if _, ok := arg.Value.(Ref); ok {
				if RootDocumentRefs.Contains(arg) {
					errors = append(errors, NewError(CompileErr, arg.Location, "args must not shadow %v (use a different variable name)", arg))
				}
			}
		}

		return true
	})

	WalkExprs(node, func(expr *Expr) bool {
		if expr.IsAssignment() {
			// assign() can be called directly, so we need to assert its given first operand exists before checking its name.
			if nameOp := expr.Operand(0); nameOp != nil {
				name := nameOp.String()
				if RootDocumentRefs.Contains(RefTerm(VarTerm(name))) {
					errors = append(errors, NewError(CompileErr, expr.Location, "variables must not shadow %v (use a different variable name)", name))
				}
			}
		}
		return false
	})

	return errors
}

func walkCalls(node interface{}, f func(interface{}) bool) {
	vis := &GenericVisitor{func(x interface{}) bool {
		switch x := x.(type) {
		case Call:
			return f(x)
		case *Expr:
			if x.IsCall() {
				return f(x)
			}
		case *Head:
			// GenericVisitor doesn't walk the rule head ref
			walkCalls(x.Reference, f)
		}
		return false
	}}
	vis.Walk(node)
}

func checkDeprecatedBuiltins(deprecatedBuiltinsMap map[string]struct{}, node interface{}) Errors {
	errs := make(Errors, 0)

	walkCalls(node, func(x interface{}) bool {
		var operator string
		var loc *Location

		switch x := x.(type) {
		case *Expr:
			operator = x.Operator().String()
			loc = x.Loc()
		case Call:
			terms := []*Term(x)
			if len(terms) > 0 {
				operator = terms[0].Value.String()
				loc = terms[0].Loc()
			}
		}

		if operator != "" {
			if _, ok := deprecatedBuiltinsMap[operator]; ok {
				errs = append(errs, NewError(TypeErr, loc, "deprecated built-in function calls in expression: %v", operator))
			}
		}

		return false
	})

	return errs
}

func checkDeprecatedBuiltinsForCurrentVersion(node interface{}) Errors {
	deprecatedBuiltins := make(map[string]struct{})
	capabilities := CapabilitiesForThisVersion()
	for _, bi := range capabilities.Builtins {
		if bi.IsDeprecated() {
			deprecatedBuiltins[bi.Name] = struct{}{}
		}
	}

	return checkDeprecatedBuiltins(deprecatedBuiltins, node)
}

type RegoCheckOptions struct {
	NoDuplicateImports      bool
	NoRootDocumentOverrides bool
	NoDeprecatedBuiltins    bool
	NoKeywordsAsRuleNames   bool
	RequireIfKeyword        bool
	RequireContainsKeyword  bool
	RequireRuleBodyOrValue  bool
}

func NewRegoCheckOptions() RegoCheckOptions {
	// all options are enabled by default
	return RegoCheckOptions{
		NoDuplicateImports:      true,
		NoRootDocumentOverrides: true,
		NoDeprecatedBuiltins:    true,
		NoKeywordsAsRuleNames:   true,
		RequireIfKeyword:        true,
		RequireContainsKeyword:  true,
		RequireRuleBodyOrValue:  true,
	}
}

// CheckRegoV1 checks the given module or rule for errors that are specific to Rego v1.
// Passing something other than an *ast.Rule or *ast.Module is considered a programming error, and will cause a panic.
func CheckRegoV1(x interface{}) Errors {
	return CheckRegoV1WithOptions(x, NewRegoCheckOptions())
}

func CheckRegoV1WithOptions(x interface{}, opts RegoCheckOptions) Errors {
	switch x := x.(type) {
	case *Module:
		return checkRegoV1Module(x, opts)
	case *Rule:
		return checkRegoV1Rule(x, opts)
	}
	panic(fmt.Sprintf("cannot check rego-v1 compatibility on type %T", x))
}

func checkRegoV1Module(module *Module, opts RegoCheckOptions) Errors {
	var errors Errors
	if opts.NoDuplicateImports {
		errors = append(errors, checkDuplicateImports([]*Module{module})...)
	}
	if opts.NoRootDocumentOverrides {
		errors = append(errors, checkRootDocumentOverrides(module)...)
	}
	if opts.NoDeprecatedBuiltins {
		errors = append(errors, checkDeprecatedBuiltinsForCurrentVersion(module)...)
	}

	for _, rule := range module.Rules {
		errors = append(errors, checkRegoV1Rule(rule, opts)...)
	}

	return errors
}

func checkRegoV1Rule(rule *Rule, opts RegoCheckOptions) Errors {
	t := "rule"
	if rule.isFunction() {
		t = "function"
	}

	var errs Errors

	if opts.NoKeywordsAsRuleNames && IsKeywordInRegoVersion(rule.Head.Name.String(), RegoV1) {
		errs = append(errs, NewError(ParseErr, rule.Location, "%s keyword cannot be used for rule name", rule.Head.Name.String()))
	}
	if opts.RequireRuleBodyOrValue && rule.generatedBody && rule.Head.generatedValue {
		errs = append(errs, NewError(ParseErr, rule.Location, "%s must have value assignment and/or body declaration", t))
	}
	if opts.RequireIfKeyword && rule.Body != nil && !rule.generatedBody && !ruleDeclarationHasKeyword(rule, tokens.If) && !rule.Default {
		errs = append(errs, NewError(ParseErr, rule.Location, "`if` keyword is required before %s body", t))
	}
	if opts.RequireContainsKeyword && rule.Head.RuleKind() == MultiValue && !ruleDeclarationHasKeyword(rule, tokens.Contains) {
		errs = append(errs, NewError(ParseErr, rule.Location, "`contains` keyword is required for partial set rules"))
	}

	return errs
}
