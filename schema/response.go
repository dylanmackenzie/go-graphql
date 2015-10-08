package schema

import (
	"errors"
	"log"
	"sync"

	"dylanmackenzie.com/graphql/result"
)

type ResponseNode struct {
	// The runtime type of the object which is being resolved by this
	// response node.
	descriptor AbstractDescriptor

	// A map containing data about the object currently being resolved.
	// Leaf fields will be resolved automatically by the value in this
	// map corresponding to their name.
	result.Map

	Fields []string   // List of fields that must be resolved.
	Args   result.Map // The arguments for the current node.

	parent   *ResponseNode   // The ResponseNode that initiated this one.
	children []*ResponseNode // All Response nodes initiated by this one.

	resolved bool            // Whether or not this ResponseNode has been resolved.
	wg       *sync.WaitGroup // The WaitGroup waiting for this ResponseNode to resolve.
}

// Internal constructor for a response node. Only for initializing the
// map and slice types which should never be nil. The WaitGroup and
// descriptor must be populated by the caller.
func newResponseNode(parent *ResponseNode, name string) (*ResponseNode, error) {
	node := &ResponseNode{
		Fields:   make([]string, 0),
		Args:     make(map[string]interface{}),
		Map:      make(map[string]interface{}),
		children: make([]*ResponseNode, 0),
		parent:   parent,
		wg:       new(sync.WaitGroup),
	}

	if parent != nil {
		parent.children = append(parent.children, node)
		field, found := parent.descriptor.Field(name)
		if !found {
			return nil, errors.New("No field found by given name")
		}
		node.descriptor = field.Result.(AbstractDescriptor)
	}

	return node, nil
}

func (r *ResponseNode) panicIfResolved() {
	if r.resolved {
		log.Panicf("Response for field '%s' has already been resolved", r.descriptor.Name())
	}
}

// Resolve marks a request as complete
func (r *ResponseNode) resolve() {
	r.panicIfResolved()
	r.resolved = true
}
