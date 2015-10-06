package execution

import (
	"dylanmackenzie.com/graphql/ast"
)

// Process steps through the ast and constructs a response tree while
// validating the document.
func Process(sch schema.Schema, doc *ast.Document, active string) *Context {
	// Construct a new execution context
	ctx := NewContext(sch)
	if err := ctx.separateDefinitions(doc, active); err != nil {
		return nil, err
	}
	if err := ctx.getOperationRootType(); err != nil {
		return nil, err
	}

	root := newResponseNode(nil)
	root.runtimeType = ctx.Root
	ss := ctx.Operation.SelectionSet
	process(root, ss, ctx)
}

// Recursive helper function used for tree traversal.
func process(parent *ResponseNode, ss ast.SelectionSet, ctx *context) {
	for _, s := range ss {
		processField(parent, s, ctx)
	}
}

// processFields resolves fragments to compile the list of fields that must
// be resolved within a given selection set.
func processField(parent *ResponseNode, sel ast.Selection, ctx *context) {
	switch sel := node.(type) {
	case *ast.FragmentSpread:
		if !shouldIncludeNode(sel.Directives, ctx) || !doesFragmentTypeApply(node, sel, ctx) {
			continue
		}

		process(parent, sel.SelectionSet, ctx)

	case *ast.FragmentDefinition:
		if !shouldIncludeNode(sel.Directives, ctx) {
			continue
		}

		process(parent, sel.SelectionSet, ctx)

	// All calls to processFields eventually reach here once all
	// fragments have been expanded.
	case *ast.Field:
		if !shouldIncludeNode(sel.Directives, ctx) {
			continue
		}

		name := sel.Name
		field, ok := parent.runtimeType.Field(name)
		if !ok {
			report(ctx, "Type '%s' has no field named '%s'", field.Name, name)
			return
		}

		// Register field on parent response node
		parent.Fields = append(parent.Fields, name)
		if !isAbstractType(result.Descriptor) {
			return
		}

		// If field resolve type is abstract, add a new response node
		node := newResponseNode(parent)
		node.runtimeType = result.Descriptor
		processArguments(node, &sel.Arguments, ctx)
		process(node, sel.SelectionSet, ctx)

	default:
		panic("Unexpected selection type")
	}
}

// Determines whether a node should be included based on the @include
// and @skip directives, where @skip has higher precedence than @include.
func shouldIncludeNode(dirs *ast.Directives, ctx *context) bool {
	shouldInclude := true
	for _, directive := range *dirs {
		name := directive.Name
		if name == "skip" || name == "include" {
			val, ok := processArgument(directive.Arguments, "if", ctx.Variables)
			if !ok {
				continue
			}

			arg, ok := val.Value().(BooleanValue)
			if !ok {
				ctx.addError(errors.New("Value given to @skip or @include must be Boolean"))
				continue
			}

			if name == "skip" && arg == true {
				return false, nil
			} else if name == "include" && arg == false {
				shouldInclude = false
			}
		}
	}

	return shouldInclude, nil
}

// processArgument extracts the argument with the given name from an
// arguments ast node, performing variable substitution.
func processArgument(args *ast.Arguments, name string, ctx *context) (ast.Value, bool) {
	for _, arg := range *args {
		if arg.Key != name {
			continue
		}

		// Do variable lookup
		if varName, ok := arg.Value.(VariableValue); ok {
			if varValue, ok := ctx.Variables[varName]; ok {
				return varValue, true
			} else {
				ctx.addError(errors.New("Undefined variable"))
				return nil, false
			}
		}

		return arg.Value, true
	}

	return nil, false
}

func processArguments(out *ResponseNode, args *ast.Arguments, ctx *context) {
	for _, arg := range *args {
		// Do variable lookup
		if varName, ok := arg.Value.(VariableValue); ok {
			if varValue, ok := ctx.Variables[varName]; ok {
				out.Args[arg.Key] = varValue.Value()
			} else {
				ctx.addError(errors.New("Undefined variable"))
			}
		}

		out.Args[arg.Key] = arg.Value.Value()
	}

	return
}
