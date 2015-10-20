// Package ast contains a lexer, parser, and AST for GraphQL
// documents.
package ast

import (
	"bytes"
	"reflect"
)

// The type of a given operation (either "query" or "mutation").
type OperationType uint8

const (
	QUERY OperationType = iota
	MUTATION
)

// An interface implemented by all nodes in the AST to allow serializing
// them
type Node interface {
	//io.WriterTo
}

// Document is a GraphQL document consisting of a series of definitions.
// It is the root node in the AST.
type Document struct {
	Definitions Definitions
}

// A slice of Definition.
type Definitions []Definition

// A single definition in a GraphQL document.
type Definition interface {
	Node
	definition()
}

// A Definition representing some type of operation on the dataset.
type OperationDefinition struct {
	Name         string
	OpType       OperationType
	Variables    Variables
	Directives   Directives
	SelectionSet SelectionSet
}

// A Definition representing the structure of some data which can be
// composed to make other Definitions.
type FragmentDefinition struct {
	Name         string
	Type         string
	Directives   Directives
	SelectionSet SelectionSet

	inline bool
}

// A slice of Variable.
type Variables []Variable

// A Variable is the declaration of a GraphQL variable.
type Variable struct {
	Name     string
	Type     string
	Nullable bool
	Default  Value
}

// A SelectionSet is a slice of Selection.
type SelectionSet []Selection

// A Selection is a group of fields which will be used in an operation.
type Selection interface {
	Node
	selection()
}

// A Field is a discrete piece of information about an object in the
// dataset.
type Field struct {
	Name         string
	Alias        string
	Directives   Directives
	Arguments    Arguments
	SelectionSet SelectionSet
}

// A FragmentSpread is the instantiation of a FragmentDefinition
// within some other definition.
type FragmentSpread struct {
	Name       string
	Directives Directives
}

// A slice of Argument.
type Arguments []Argument

// An Argument is a key-value pair used to parameterize operations on
// the dataset.
type Argument struct {
	Key   string
	Value Value
}

// A slice of Directive.
type Directives []Directive

// A Directive is a way to describe alternate runtime execution and type
// validation behavior in a GraphQL document.
type Directive struct {
	Name      string
	Arguments Arguments
}

// Values

// A Value is a container that can hold any data type suitable for an
// argument.
type Value interface {
	Node
	Value() interface{}
}

type VariableValue string         // A GraphQL variable
type IntValue int                 // An int
type FloatValue float64           // A float
type StringValue string           // A string
type EnumValue string             // An enum declared somewhere else in the document
type BooleanValue bool            // A boolean
type ListValue []Value            // A list of one of the above values
type ObjectValue map[string]Value // A map of name-value pairs

func (v VariableValue) Value() interface{} { return v }
func (v IntValue) Value() interface{}      { return int(v) }
func (v FloatValue) Value() interface{}    { return float64(v) }
func (v StringValue) Value() interface{}   { return string(v) }
func (v EnumValue) Value() interface{}     { return v }
func (v BooleanValue) Value() interface{}  { return bool(v) }
func (v ListValue) Value() interface{}     { return v }
func (v ObjectValue) Value() interface{}   { return v }

// Types

type TypeDefinition interface {
	Node
	TypeName() string
	typeDefinition()
}

type AbstractTypeDefinition interface {
	TypeDefinition
	Field(string) (*TypeField, bool)
}

type ScalarDefinition struct {
	Name string
	Kind reflect.Kind
}

type EnumDefinition struct {
	Name   string
	Values map[string]int
}

type ObjectDefinition struct {
	Name       string
	Fields     TypeFields
	Implements []string
}

type InterfaceDefinition struct {
	Name   string
	Fields TypeFields
}

type UnionDefinition struct {
	Name    string
	Members []TypeDescriptor
}

type TypeFields []TypeField
type TypeField struct {
	Name      string
	Type      TypeDescriptor
	Arguments ArgumentDeclarations

	// TODO: This field will be filled out when the query is actually
	// resolved, not when it is parsed. This avoids looking up the type
	// in the schema, but requires an extra pass before any work is
	// done.
	Definition TypeDefinition
}

type ArgumentDeclarations []ArgumentDeclaration
type ArgumentDeclaration struct {
	Key  string
	Type TypeDescriptor
}

func (obj *InterfaceDefinition) Field(name string) (*TypeField, bool) {
	return findTypeField(obj.Fields, name)
}
func (obj *ObjectDefinition) Field(name string) (*TypeField, bool) {
	return findTypeField(obj.Fields, name)
}
func findTypeField(fields []TypeField, name string) (*TypeField, bool) {
	for i, field := range fields {
		if field.Name == name {
			return &fields[i], true
		}
	}

	return nil, false
}

// A TypeDescriptor is a reference to a TypeDefinition which could be a
// nullable version of that type, a list of that type, an input object
// containing several types, or a combination of any of the above.
type TypeDescriptor interface {
	Name() string
	Nullable() bool
}

type BaseType struct {
	name     string
	nullable bool
}

type ListType struct {
	OfType   TypeDescriptor
	nullable bool
}

type InputObjectType struct {
	Fields   map[string]TypeDescriptor
	nullable bool
}

// Interface implementations

func (*FragmentDefinition) definition()  {}
func (*OperationDefinition) definition() {}
func (*ScalarDefinition) definition()    {}
func (*EnumDefinition) definition()      {}
func (*ObjectDefinition) definition()    {}
func (*InterfaceDefinition) definition() {}
func (*UnionDefinition) definition()     {}

func (d *ScalarDefinition) typeDefinition()    {}
func (d *EnumDefinition) typeDefinition()      {}
func (d *ObjectDefinition) typeDefinition()    {}
func (d *InterfaceDefinition) typeDefinition() {}
func (d *UnionDefinition) typeDefinition()     {}

func (d *ScalarDefinition) TypeName() string    { return d.Name }
func (d *EnumDefinition) TypeName() string      { return d.Name }
func (d *ObjectDefinition) TypeName() string    { return d.Name }
func (d *InterfaceDefinition) TypeName() string { return d.Name }
func (d *UnionDefinition) TypeName() string     { return d.Name }

func (*Field) selection()              {}
func (*FragmentSpread) selection()     {}
func (*FragmentDefinition) selection() {}

// TODO: InputObject should not have a BaseType method. Maybe separate
// interface?
func (t *BaseType) Name() string   { return t.name }
func (t *BaseType) Nullable() bool { return t.nullable }
func (t *ListType) Name() string   { return "[" + t.OfType.Name() + "]" }
func (t *ListType) Nullable() bool { return t.nullable }
func (t *InputObjectType) Name() string {
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
func (t *InputObjectType) Nullable() bool { return t.nullable }
