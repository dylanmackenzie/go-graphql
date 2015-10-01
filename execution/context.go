package execution

import (
	"net/url"

	ast "dylanmackenzie.com/graphql/parser"
	"dylanmackenzie.com/graphql/schema"
)

type Context struct {
	Schema schema.Schema
	Root   schema.Object

	Operation      *ast.OperationDefinition
	VariableValues map[string]ast.Value
	Fragments      map[string]*ast.FragmentDefinition
}

// Build an execution context from a schema, graphql document, and a
// string naming the active definition in the document (which must be the empty
// string if the client did not specify an operation name).
func NewContext(sch schema.Schema, doc *ast.Document, active string) (*Context, error) {
	ctx := &Context{Schema: sch}

	if err := ctx.separateDefinitions(doc, active); err != nil {
		return nil, err
	}

	if err := ctx.getOperationRootType(); err != nil {
		return nil, err
	}

	return ctx, nil
}

// getOperationRootType finds the appropriate root in the schema for the
// active GraphQL operation.
func (ctx *Context) getOperationRootType() error {
	switch ctx.Operation.OpType {
	case QUERY:
		ctx.Root = ctx.Schema.QueryRoot
	case MUTATION:
		ctx.Root = ctx.Schema.MutationRoot
	default:
		return errors.New("Operation Type must be either query or mutation")
	}

	if ctx.Root == nil {
		return nil, errors.New("Schema does not provide a root object for the selected operation")
	}

	return nil
}

// separateDefinitions collects all top level fragments and finds the
// active operation of the given document, storing it in ctx
func (ctx *Context) separateDefinitions(doc *ast.Document, active string) error {
	cnt := 0
	for i, def := range doc.Definitions {
		if op, ok := def.(*ast.OperationFragment); ok {
			if _, exists := ctx.Fragments[op.Name]; exists {
				return errors.New("Multiple fragments with same name")
			}
			ctx.Fragments[op.Name] = op
		}

		op := def.(*ast.OperationDefinition)

		if op.Name == active && active != "" {
			ctx.Operation = op
			continue
		}

		cnt++
		if cnt == 1 && active == "" {
			ctx.Operation = op
		}
	}

	if active != "" { // If we supplied a name to the search, ensure we had a match
		if ctx.Operation != nil {
			return nil
		} else {
			return errors.New("No operation with given name found in document")
		}
	} else { // Otherwise ensure exactly one operation was present
		if ctx.Operation != nil && cnt == 1 {
			return nil
		} else {
			return errors.New("Document did not contain exactly one operation")

		}
	}
}

// ParseVariablesFromJSON parses a set of GraphQL variables from a JSON string
func (ctx *Context) ParseVariablesFromJSON(json string) error {
	return nil
}
