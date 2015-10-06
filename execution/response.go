package execution

type ResponseNode struct {
	// The runtime type of the object which is being resolved by this
	// response node.
	runtimeType *schema.ObjectDescriptor

	// A map containing data about the object currently being resolved.
	// Leaf fields will be resolved automatically by the value in this
	// map corresponding to their name.
	untypedMap

	Fields []string   // List of fields that must be resolved.
	Args   untypedMap // The arguments for the current node.

	parent   *ResponseNode   // The ResponseNode that initiated this one.
	children []*ResponseNode // All Response nodes initiated by this one.

	resolved bool            // Whether or not this ResponseNode has been resolved.
	wg       *sync.WaitGroup // The WaitGroup waiting for this ResponseNode to resolve.
}

// Internal constructor for a response node. Only for initializing the
// map and slice types which should never be nil. The WaitGroup and
// runtimeType must be populated by the caller.
func newResponseNode(parent *ResponseNode) *ResponseNode {
	node := &ResponseNode{
		Fields:     make([]string),
		Args:       make(map[string]interface{}),
		untypedMap: make(map[string]interface{}),
		Children:   make([]*ResponseNode),
		Parent:     parent,
	}

	if parent != nil {
		parent.Children = append(parent.Children, node)
	}

	return node
}

func (r *ResponseNode) panicIfResolved() {
	if r.resolved {
		log.Panicf("Response for field '%s' has already been resolved", r.runtimeType.Name())
	}
}

// Resolve marks a request as complete
func (r *ResponseNode) resolve() {
	r.panicIfResolved()
	r.resolved = true
	wg.Done()
}
