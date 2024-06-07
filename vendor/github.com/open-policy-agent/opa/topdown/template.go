package topdown

import (
	"bytes"
	"text/template"

	"github.com/open-policy-agent/opa/ast"
	"github.com/open-policy-agent/opa/topdown/builtins"
)

func renderTemplate(_ BuiltinContext, operands []*ast.Term, iter func(*ast.Term) error) error {
	preContentTerm, err := builtins.StringOperand(operands[0].Value, 1)
	if err != nil {
		return err
	}

	templateVariablesTerm, err := builtins.ObjectOperand(operands[1].Value, 2)
	if err != nil {
		return err
	}

	var templateVariables map[string]interface{}

	if err := ast.As(templateVariablesTerm, &templateVariables); err != nil {
		return err
	}

	tmpl, err := template.New("template").Parse(string(preContentTerm))
	if err != nil {
		return err
	}

	// Do not attempt to render if template variable keys are missing
	tmpl.Option("missingkey=error")
	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, templateVariables); err != nil {
		return err
	}

	return iter(ast.StringTerm(buf.String()))
}

func init() {
	RegisterBuiltinFunc(ast.RenderTemplate.Name, renderTemplate)
}
