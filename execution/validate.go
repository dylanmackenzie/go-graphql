package execution

import (
	"fmt"

	"dylanmackenzie.com/graphql/ast"
)

func report(ctx Context, s string, v ...interface{}) {
	ctx.addError(fmt.Sprintf(s, v...))
}

// Ensures that all definitions in the document have unique names, and
// that if there is an unnamed operation, it is the only one in the
// document.
func evalDefinitions(doc *ast.Document, ctx *Context) error {
	foundFrags := make(map[string]bool, len(doc.Definitions))
	foundOps := make(map[string]bool, len(doc.Definitions))
	opCount := 0 // Set to -1 when an unnamed operation is found
	for _, t := range doc.Definitions {
		switch def := t.(type) {
		case *ast.OperationDefinition:
			if opCount == -1 {
				report(ctx, "Unnamed operation must be the only one in a document")
			}

			if foundOps[def.Name] {
				report(ctx, "Multiple operations named '%s'", def.Name)
			}

			if def.Name == "" {
				opCount = -1
			} else {
				opCount++
			}

		case *ast.FragmentDefinition:
			if foundFrags[def.Name] {
				report(ctx, "Multiple operations named '%s'", def.Name)
			}
		}
	}
}
