package schema

import (
	"bytes"
	"encoding/json"
	"log"
	"sync"

	"dylanmackenzie.com/graphql/ast"
)

type ResponseNode struct {
	//
	name string

	// The type expected as a result of this response node
	resultType ast.AbstractTypeDefinition
	isNullable bool

	// A map containing data about the object currently being resolved.
	// Leaf fields will be resolved automatically by the value in this
	// map corresponding to their name.
	resultMap

	Fields []string  // List of fields that must be resolved.
	Args   resultMap // The arguments for the current node.
	null   bool      // Whether or not the response is null.

	parent   *ResponseNode   // The ResponseNode that initiated this one.
	children []*ResponseNode // All Response nodes initiated by this one.

	// If a ResponseNode is cloned to create a list of responses, this
	// contains a pointer to all nodes that will comprise the list.
	siblings []*ResponseNode

	resolved bool            // Whether or not this ResponseNode has been resolved.
	wg       *sync.WaitGroup // The WaitGroup waiting for this ResponseNode to resolve.
}

// Constructor for a response node. Only for initializing the
// map and slice types which should never be nil.
func NewResponseNode(parent *ResponseNode, field *ast.TypeField) *ResponseNode {
	node := &ResponseNode{
		Fields:    make([]string, 0),
		Args:      make(map[string]interface{}),
		resultMap: make(map[string]interface{}),
		children:  make([]*ResponseNode, 0),
		parent:    parent,
		name:      "__root",
		wg:        new(sync.WaitGroup),
	}

	if field != nil {
		if field.Definition == nil {
			log.Panicf("Field '%s' in type '%s' has not been finalized", field.Name, parent.resultType.TypeName())
		}

		def, ok := field.Definition.(ast.AbstractTypeDefinition)
		if !ok {
			panic("NewResponseNode called with field which is not abstract")
		}

		node.resultType = def
		node.isNullable = field.Type.Nullable()
		node.name = field.Name
	}

	if parent != nil {
		parent.children = append(parent.children, node)
	}

	return node
}

func (r *ResponseNode) panicIfResolved() {
	if r.resolved {
		log.Panicf("Response for field '%s' has already been resolved", r.resultType.TypeName())
	}
}

// Null sets the given response node to null.
func (r *ResponseNode) Null(b bool) {
	if b == true && !r.isNullable {
		panic("Null called on non-nullable type")
	}

	r.null = b
}

// Resolve marks a request as complete
func (r *ResponseNode) resolve() {
	r.panicIfResolved()
	r.resolved = true
}

// ResponseNode implements json.Marshaler
func (r *ResponseNode) MarshalJSON() ([]byte, error) {
	buf := new(bytes.Buffer)
	err := r.marshalJSON(buf)
	return buf.Bytes(), err
}

func (r *ResponseNode) marshalJSON(buf *bytes.Buffer) error {
	// Handle null ResponseNode
	if r.null == true {
		if !r.isNullable {
			panic("Response node for non-nullable type has been set to null")
		}

		log.Println("null")
		_, err := buf.Write([]byte("null"))
		if err != nil {
			return err
		}

		return nil

	}

	if err := buf.WriteByte(byte('{')); err != nil {
		return err
	}

	for i, fieldName := range r.Fields {
		field, ok := r.resultType.Field(fieldName)
		if !ok {
			panic("MarshalJSON called on node with invalid field")
		}

		if i != 0 {
			if err := buf.WriteByte(byte(',')); err != nil {
				return err
			}
		}

		if err := buf.WriteByte(byte('"')); err != nil {
			return err
		}

		if _, err := buf.WriteString(fieldName); err != nil {
			return err
		}

		if _, err := buf.Write([]byte{'"', ':'}); err != nil {
			return err
		}

		// If the field is a scalar, look it up in the result map and
		// write it to the buffer.
		if !ast.IsAbstractType(field.Definition) {
			result, ok := r.resultMap.Get(fieldName)
			if !ok {
				panic("No field set")
			}

			json, err := json.Marshal(result)
			if err != nil {
				return err
			}

			if _, err := buf.Write(json); err != nil {
				return err
			}

			continue
		}

		// Otherwise, the field is abstract, so we look in the list of
		// children for the ResponseNode which resolves it.
		var node *ResponseNode
		for _, child := range r.children {
			if child.name == field.Name {
				node = child
			}
		}

		if _, ok := field.Type.(*ast.ListType); ok {
			if err := marshalList(node, buf); err != nil {
				return err
			}
		} else {
			if err := node.marshalJSON(buf); err != nil {
				return err
			}
		}
	}

	if err := buf.WriteByte(byte('}')); err != nil {
		return err
	}
	return nil
}

func marshalList(r *ResponseNode, buf *bytes.Buffer) error {
	if err := buf.WriteByte(byte('[')); err != nil {
		return err
	}
	if err := r.marshalJSON(buf); err != nil {
		return err
	}
	for _, sibling := range r.siblings {
		if err := buf.WriteByte(byte(',')); err != nil {
			return err
		}
		if err := sibling.marshalJSON(buf); err != nil {
			return err
		}
	}

	if err := buf.WriteByte(byte(']')); err != nil {
		return err
	}

	return nil
}
