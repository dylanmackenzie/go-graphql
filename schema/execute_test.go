package schema

import (
	"testing"

	"dylanmackenzie.com/graphql/ast"
)

func TestProcessArgument(t *testing.T) {
	ctx := NewContext(nil)
	ctx.Variables["ifVar"] = ast.BooleanValue(true)

	args := &ast.Arguments{
		{"if", ast.VariableValue("ifVar")},
	}

	expect := ast.BooleanValue(true)
	val, ok := processArgument(args, "if", ctx)
	if !ok || val != expect {
		t.Errorf("Expected '%s', got '%s'\n", expect, val)
	}
}
