package execution

// ResolveFunc is a callback which resolves the selection set of a
// GraphQL Object.
//
// Its first argument is the GraphQL object to be processed, which
// is itself the result of a ResolveFunc higher up the tree. From there,
// a ResolveFunc should send Response objects into the channel until
// it has processed all fields.
type ResolveFunc func(r *ResponseNode)

// Resolver is the interface for a struct which implements a ResolveFunc
type Resolver interface {
	ResolveGraphQL(r *ResponseNode)
}
