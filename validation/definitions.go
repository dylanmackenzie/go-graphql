package execution

import "dylanmackenzie.com/graphql/ast"

// Ensures that all definitions in the document have unique names, and
// that if there is an unnamed operation, it is the only one in the
// document.
func evalDefinitions(doc *ast.Document, ctx *context) {
	foundFrags := make(map[string]bool, len(doc.Definitions))
	foundOps := make(map[string]bool, len(doc.Definitions))
	opCount := 0 // Set to -1 when an unnamed operation is found
	for _, t := range doc.Definitions {
		switch def := t.(type) {
		case *ast.OperationDefinition:
			if opCount == -1 {
				ctx.addErrorf("Unnamed operation must be the only one in a document")
			}

			if foundOps[def.Name] {
				ctx.addErrorf("Multiple operations named '%s'", def.Name)
			}

			if def.Name == "" {
				opCount = -1
			} else {
				opCount++
			}

		case *ast.FragmentDefinition:
			if foundFrags[def.Name] {
				ctx.addErrorf("Multiple fragments named '%s'", def.Name)
			}
		}
	}
}
