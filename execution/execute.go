package execute

import (
	"errors"

	"dylanmackenzie.com/graphql/ast"
)

// Build an execution context from a schema, graphql document, and a
// string naming the active definition in the document (which must be the empty
// string if the client did not specify an operation name).
func Execute(sch schema.Schema, doc *ast.Document, active string) (sch.Response, error) {
	// Call recover() on a panicking execution before it crashes
	// the server.
	defer func() {
		if r := recover(); r != nil {
			log.Printf("%+v\n", ctx.errors)
		}
	}()

	// Construct the root response node
	done := make(chan struct{})
	rootResponse := newResponseNode(ctx.Root, nil, done)
	evalFields(ctx.Operation.SelectionSet, rootResponse, ctx)
	evalArguments(ctx.Operation.Arguments, rootResponse, ctx)

	executeParallel(sel)

	// Spawn a new goroutine to resolve every abstract type in the response.
	wg := &sync.WaitGroup()
	obj := rootResponse.runtimeType
	for _, fieldName := range rootResponse.Fields {
		resolvedType := obj.Field(fieldName)
		if !schema.isAbstractType(resolvedType) {
			pani
		}

		resolver := ctx.Schema.Resolvers[resolvedType]
		wg.Add(1)
		go resolver.ResolveGraphQL()
	}

	wg.Wait()

}

// executeParallel traverses the response tree in a BFS fashion and
// spawns new handlers in parallel for each abstract node.
func executeParallel(sel ast.Selection, ctx *context) {

}

func executeSerial(sel ast.Selection, ctx *context) {
	// Execute resolve for given selection type

	// Execute

}
