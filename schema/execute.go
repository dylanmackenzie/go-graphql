package schema

import "dylanmackenzie.com/graphql/ast"

// Build an execution context from a schema, graphql document, and a
// string naming the active definition in the document (which must be the empty
// string if the client did not specify an operation name).
func Execute(sch *Schema, doc *ast.Document, active string) (ctx *context, err error) {
	// Construct a new execution context
	// the server.
	ctx = NewContext(sch)

	// Call recover() on a panicking execution before it crashes
	defer func() {
		if r := recover(); r != nil {
			ctx.addErrorf("%v", r)
		}
		err = ctx.Errors.Err()
	}()

	// Complete execution context
	ctx.processDefinitions(doc, active)
	ctx.getOperationRootType()
	if ctx.Root == nil {
		return
	}

	// Construct the root response node
	ctx.Response = NewResponseNode(nil, nil)
	ctx.Response.resultType = ctx.Root

	// Begin query execution in the same goroutine
	expandFields(ctx.Operation.SelectionSet, ctx.Response, ctx)
	ctx.Response.wg.Wait()

	return
}

func execute(field *ast.Field, node *ResponseNode, ctx *context) {
	defer func() {
		node.parent.wg.Done()
	}()

	// Process arguments
	processArguments(&field.Arguments, node, ctx)

	// Call the child handler, then wait for all sub-fields to fully
	// resolve.
	resolver := ctx.Schema.resolver(node.resultType.TypeName())
	resolver.ResolveGraphQL(node)
	expandFields(field.SelectionSet, node, ctx)
	node.wg.Wait()
}

// expandFields resolves fragments to compile the list of fields that must
// be resolved within a given selection set.
func expandFields(ss ast.SelectionSet, parent *ResponseNode, ctx *context) {
	def := parent.resultType
	for _, s := range ss {
		switch sel := s.(type) {
		case *ast.FragmentSpread:
			if !shouldIncludeNode(&sel.Directives, ctx) {
				continue
			}

			// lookup fragment
			frag, ok := ctx.Fragments[sel.Name]
			if !ok {
				ctx.addErrorf("No fragment named '%s' found", sel.Name)
			}

			expandFields(frag.SelectionSet, parent, ctx)

		case *ast.FragmentDefinition:
			if !shouldIncludeNode(&sel.Directives, ctx) {
				continue
			}

			expandFields(sel.SelectionSet, parent, ctx)

		// Calls to expandFields eventually reach here once all
		// fragments have been resolved.
		case *ast.Field:
			if !shouldIncludeNode(&sel.Directives, ctx) {
				continue
			}

			name := sel.Name
			field, ok := def.Field(name)
			if !ok {
				ctx.addErrorf("Type has no field named '%s'", name)
				continue
			}

			// Register field on parent response node
			parent.Fields = append(parent.Fields, name)

			// Determine if field is a (valid) leaf
			if !ast.IsAbstractType(field.Definition) {
				if len(sel.SelectionSet) != 0 {
					ctx.addErrorf("Scalar type has sub-fields in query")
				}
				continue
			} else if len(sel.SelectionSet) == 0 {
				ctx.addErrorf("Abstract type has no sub-fields in query")
				continue
			}

			// If field is not a leaf, create a ResponseNode for the
			// given field and execute the field's handler.
			node := NewResponseNode(parent, field)
			parent.wg.Add(1)
			if ctx.serialExecution {
				execute(sel, node, ctx)
			} else {
				go execute(sel, node, ctx)
			}

		default:
			panic("Unexpected selection type")
		}
	}
}

// Determines whether a node should be included based on the @include
// and @skip directives, where @skip has higher precedence than @include.
func shouldIncludeNode(dirs *ast.Directives, ctx *context) bool {
	shouldInclude := true
	for _, directive := range *dirs {
		name := directive.Name
		if name == "skip" || name == "include" {
			val, ok := processArgument(&directive.Arguments, "if", ctx)
			if !ok {
				continue
			}

			arg, ok := val.Value().(ast.BooleanValue)
			if !ok {
				ctx.addErrorf("Value given to @skip or @include must be Boolean")
				continue
			}

			if name == "skip" && arg == true {
				return false
			} else if name == "include" && arg == false {
				shouldInclude = false
			}
		}
	}

	return shouldInclude
}

// processArgument extracts the argument with the given name from an
// arguments ast node, performing variable substitution.
func processArgument(args *ast.Arguments, name string, ctx *context) (ast.Value, bool) {
	for _, arg := range *args {
		if arg.Key != name {
			continue
		}

		// Do variable lookup
		if varName, ok := arg.Value.(ast.VariableValue); ok {
			if varValue, ok := ctx.Variables[string(varName)]; ok {
				return varValue, true
			} else {
				ctx.addErrorf("Undefined variable $'%s'", varName)
				return nil, false
			}
		}

		return arg.Value, true
	}

	return nil, false
}

func processArguments(args *ast.Arguments, out *ResponseNode, ctx *context) {
	for _, arg := range *args {
		// Do variable lookup
		if varName, ok := arg.Value.(ast.VariableValue); ok {
			if varValue, ok := ctx.Variables[string(varName)]; ok {
				out.Args[arg.Key] = varValue.Value()
			} else {
				ctx.addErrorf("Undefined variable $'%s'", varName)
			}
		}

		out.Args[arg.Key] = arg.Value.Value()
	}

	return
}
