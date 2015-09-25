package parse

// Document is a GraphQL document consisting of a series of definitions
type Document struct {
	Definitions Definitions
}

type Definitions []Definition

// Definition is
type Definition interface {
	Node
	definition()
}

type OperationType uint8

const (
	QUERY OperationType = iota
	MUTATION

	FRAGMENT
)

type OperationDefinition struct {
	Name         string
	OpType       OperationType
	Variables    Variables
	Directives   Directives
	SelectionSet SelectionSet
}

type FragmentDefinition struct {
	Name         string
	Type         string
	Directives   Directives
	SelectionSet SelectionSet
}

func (*FragmentDefinition) definition()  {}
func (*OperationDefinition) definition() {}

type Variables []Variable
type Variable struct {
	Name     string
	Type     string
	Nullable bool
	Default  Value
}

// Selection

type SelectionSet []Selection
type Selection interface {
	Node
	selection()
}

func (*Field) selection()              {}
func (*FragmentSpread) selection()     {}
func (*FragmentDefinition) selection() {}

type Field struct {
	Name         string
	Alias        string
	Directives   Directives
	Arguments    Arguments
	SelectionSet SelectionSet
}

type FragmentSpread struct {
	Name       string
	Directives Directives
}

type Arguments []Argument
type Argument struct {
	Key   string
	Value Value
}

type Directives []Directive
type Directive struct {
	Name      string
	Arguments Arguments
}

type Value interface {
	Node
	Value() interface{}
}

type VariableValue string
type IntValue int
type FloatValue float64
type StringValue string
type EnumValue string
type BooleanValue bool
type ListValue []Value
type ObjectValue map[string]Value

func (v VariableValue) Value() interface{} { return v }
func (v IntValue) Value() interface{}      { return v }
func (v FloatValue) Value() interface{}    { return v }
func (v StringValue) Value() interface{}   { return v }
func (v EnumValue) Value() interface{}     { return v }
func (v BooleanValue) Value() interface{}  { return v }
func (v ListValue) Value() interface{}     { return v }
func (v ObjectValue) Value() interface{}   { return v }

// Node interface
type Node interface {
}
