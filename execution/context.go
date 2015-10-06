package execution

import (
	"errors"
	"fmt"

	"dylanmackenzie.com/graphql/ast"
	"dylanmackenzie.com/graphql/schema"
)

type context struct {
	// Schema

	Schema *schema.Schema            // A reference to the graphql type system.
	Root   *schema.ObjectDescription // The root object in the schema responsible for the active operation.

	// Document

	Operation      *ast.OperationDefinition           // The active operation.
	VariableValues map[string]ast.Value               // The variables supplied to the graphql request.
	Fragments      map[string]*ast.FragmentDefinition // The fragments present in the graphql document.

	// Response

	// The root response node
	Response *ResponseNode

	// Error Handling

	Errors []error // The list of errors encountered while processing the request.

	// A boolean indicating whether the execution should panic
	// immediately after the first error it encounters or continue
	// parsing the entire document. Should be set to true only for
	// development purposes.
	lazyPanic bool
}

func NewContext(sch *schema.Schema) *context {
	return &context{
		Schema:         sch,
		VariableValues: make(map[string]ast.Value),
		Fragments:      make(map[string]*ast.FragmentDefinition),
		Errors:         make([]Errors, 0),
	}
}

func (ctx *context) addError(err error) {
	ctx.Errors = append(ctx.Errors, err)
	if !ctx.lazyPanic {
		panic(ctx.Errors)
	}
}

func report(ctx Context, s string, v ...interface{}) {
	ctx.addError(fmt.Errorf(s, v...))
}

// getOperationRootType finds the appropriate root in the schema for the
// active GraphQL operation and stores it in ctx.Root
func (ctx *context) getOperationRootType() {
	switch ctx.Operation.OpType {
	case QUERY:
		ctx.Root = ctx.Schema.QueryRoot
	case MUTATION:
		ctx.Root = ctx.Schema.MutationRoot
	default:
		// This should be caught in the parsing stage
		report(ctx, "Operation Type must be either query or mutation")
	}

	if ctx.Root == nil {
		report(ctx, "Schema does not provide a root object for the selected operation")
	}
}

// processDefininition collects all top level fragments and finds the
// active operation of the given document, storing it in ctx, while
// ensuring the uniqueness of all definitions. Covers sections 5.1 and
// 5.4.1 of Validation.
func (ctx *context) processDefinitions(doc *ast.Document, active string) {
	foundOps := make(map[string]bool, len(doc.Definitions))
	opCount := 0
	for _, t := range doc.Definitions {
		switch def := t.(type) {
		case *ast.OperationDefinition:
			if foundOps[def.Name] {
				report(ctx, "Multiple operations named '%s'", def.Name)
				break
			}

			opCount++
			foundOps[def.Name] = def

			// Check unnamed fragment count
			if def.Name == "" && opCount != 1 {
				report(ctx, "Unnamed operation must be the only one in a document")
			}

			if def.Name == active {
				ctx.Operation = def
			}

		case *ast.FragmentDefinition:
			if _, ok := ctx.Fragments[def.Name]; ok {
				report(ctx, "Multiple fragments named '%s'", def.Name)
			} else {
				ctx.Fragments[def.Name] = def
			}
		}
	}

	if ctx.Operation == nil {
		if active == "" {
			report(ctx, "Expecting unnamed definition, but none found")
		} else {
			report(ctx, "Expecting definition named '%s', but none found", active)
		}
	}
}

// ParseVariablesFromJSON parses a set of GraphQL variables from a JSON string
func (ctx *context) ParseVariablesFromJSON(json string) error {
	return nil
}
