package schema

import (
	"bytes"

	"dylanmackenzie.com/graphql/ast"
)

//go:generate stringer -type=Kind
type Kind int

// Kind is used to perform type conversions between scalars
const (
	Other Kind = iota

	// Scalars
	Int
	Float
	String
	Boolean
)

// A Descriptor is an abstract representation of a type.
//
// Each instance which implements Descriptor describes a separate Type
type Descriptor interface {
	Name() string   // The name of the type. Must be unique within a GraphQL schema.
	Nullable() bool // Whether or not the type is nullable
}

type AbstractDescriptor interface {
	Descriptor
	Field(string) (*Field, bool)
}

// Leaf types
type ScalarDescriptor struct {
	name string
	Kind Kind
}

type EnumDescriptor struct {
	name   string
	Values map[string]int
}

// Abstract types

type Field struct {
	Name      string
	Result    Descriptor
	Arguments map[string]Descriptor
}

type ObjectDescriptor struct {
	name       string
	Fields     []Field
	Implements []*InterfaceDescriptor
}

// Convenience method for finding a field by name
func (obj *ObjectDescriptor) Field(name string) (*Field, bool) {
	for i, field := range obj.Fields {
		if field.Name == name {
			return &obj.Fields[i], true
		}
	}

	return nil, false
}

type InterfaceDescriptor struct {
	name   string
	Fields []Field
}

// Convenience method for finding a field by name
func (obj *InterfaceDescriptor) Field(name string) (*Field, bool) {
	for i, field := range obj.Fields {
		if field.Name == name {
			return &obj.Fields[i], true
		}
	}

	return nil, false
}

type UnionDescriptor struct {
	name    string
	Members []*ObjectDescriptor
}

// Composite types

type NonNullDescriptor struct {
	OfType Descriptor
}

type ListDescriptor struct {
	OfType Descriptor
}

type InputObjectDescriptor struct {
	Fields map[string]Descriptor
}

// Descriptor Methods
//
// Abstract types must have pointer receivers because they could be
// self-referential.

func (t *ScalarDescriptor) Name() string      { return t.name }
func (t *ScalarDescriptor) Nullable() bool    { return true }
func (t *EnumDescriptor) Name() string        { return t.name }
func (t *EnumDescriptor) Nullable() bool      { return true }
func (t *ObjectDescriptor) Name() string      { return t.name }
func (t *ObjectDescriptor) Nullable() bool    { return true }
func (t *InterfaceDescriptor) Name() string   { return t.name }
func (t *InterfaceDescriptor) Nullable() bool { return true }
func (t *UnionDescriptor) Name() string       { return t.name }
func (t *UnionDescriptor) Nullable() bool     { return true }
func (t *NonNullDescriptor) Name() string     { return t.OfType.Name() + "!" }
func (t *NonNullDescriptor) Nullable() bool   { return false }
func (t *ListDescriptor) Name() string        { return "[" + t.OfType.Name() + "]" }
func (t *ListDescriptor) Nullable() bool      { return true }
func (t *InputObjectDescriptor) Name() string {
	buf := &bytes.Buffer{}
	buf.WriteString("{")
	first := true
	for k, v := range t.Fields {
		if !first {
			buf.WriteString(",")
		}

		buf.WriteString(k)
		buf.WriteString(":")
		buf.WriteString(v.Name())
		first = false
	}
	buf.WriteString("}")
	return buf.String()
}
func (t *InputObjectDescriptor) Nullable() bool { return true }

// Predefined Scalar Types
var (
	IntType     = &ScalarDescriptor{name: "Int", Kind: Int}
	FloatType   = &ScalarDescriptor{name: "Fload", Kind: Float}
	StringType  = &ScalarDescriptor{name: "String", Kind: String}
	BooleanType = &ScalarDescriptor{name: "Boolean", Kind: Boolean}
	IDType      = &ScalarDescriptor{name: "ID", Kind: String}
)

// Cache of composite type descriptors so that types can be compared by
// reference equality.
// TODO: Ensure that naming collisions are not possible.
var cache = map[string]Descriptor{}

// Composite Constructors
func NonNullOf(desc Descriptor) Descriptor {
	t := &NonNullDescriptor{OfType: desc}
	name := t.Name()
	if cached, ok := cache[name]; ok {
		return cached
	}

	cache[name] = t
	return t
}

func ListOf(desc Descriptor) Descriptor {
	t := &ListDescriptor{OfType: desc}
	name := t.Name()
	if cached, ok := cache[name]; ok {
		return cached
	}

	cache[name] = t
	return t
}

func InputObjectOf(m map[string]Descriptor) Descriptor {
	t := &InputObjectDescriptor{Fields: m}
	name := t.Name()
	if cached, ok := cache[name]; ok {
		return cached
	}

	cache[name] = t
	return t
}

// Identity Functions

func IsScalarType(desc Descriptor) bool {
	switch desc.(type) {
	case *ScalarDescriptor, *EnumDescriptor:
		return true
	default:
		return false
	}
}

func IsAbstractType(desc Descriptor) bool {
	switch t := desc.(type) {
	case *ObjectDescriptor, *InterfaceDescriptor, *UnionDescriptor:
		return true
	case *ListDescriptor:
		return IsAbstractType(t.OfType)
	case *NonNullDescriptor:
		return IsAbstractType(t.OfType)
	default:
		return false
	}
}

// Type Coercion

func IsCoercible(v ast.Value, desc Descriptor) bool {
	switch desc.(type) {
	case *ScalarDescriptor:

	}

	return true
}
