// Package parser contains a lexer, parser, and AST for GraphQL
// documents.
package parse

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

// The type of a given operation (either "query" or "mutation").
type OperationType uint8

const (
	QUERY OperationType = iota
	MUTATION
)

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
}

func (*FragmentDefinition) definition()  {}
func (*OperationDefinition) definition() {}

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

func (*Field) selection()              {}
func (*FragmentSpread) selection()     {}
func (*FragmentDefinition) selection() {}

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
func (v IntValue) Value() interface{}      { return v }
func (v FloatValue) Value() interface{}    { return v }
func (v StringValue) Value() interface{}   { return v }
func (v EnumValue) Value() interface{}     { return v }
func (v BooleanValue) Value() interface{}  { return v }
func (v ListValue) Value() interface{}     { return v }
func (v ObjectValue) Value() interface{}   { return v }

// An interface implemented by all nodes in the AST to allow serializing
// them
type Node interface {
	// io.WriterTo
	// fmt.Stringer
}
