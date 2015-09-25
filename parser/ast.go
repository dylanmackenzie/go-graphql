package parse

// Document is a GraphQL document consisting of a series of definitions
type Document struct {
	Definitions Definitions
}

type Definitions []Definition

// Definition is
type Definition interface {
	Node
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

type Variables []Variable
type Variable struct {
	Name     string
	Type     string
	Nullable bool
	Default  string
}

// Selection

type SelectionSet []Selection
type Selection interface {
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
	Value() interface{}
}

type IntValue struct{ Val int }
type FloatValue struct{ Val float64 }
type StringValue struct{ Val string }
type EnumValue struct{ Val string }
type BooleanValue struct{ Val bool }
type ListValue struct{ Val []Value }
type ObjectValue struct{ Val map[string]Value }

func (v IntValue) Value() interface{}     { return v.Val }
func (v FloatValue) Value() interface{}   { return v.Val }
func (v StringValue) Value() interface{}  { return v.Val }
func (v EnumValue) Value() interface{}    { return v.Val }
func (v BooleanValue) Value() interface{} { return v.Val }
func (v ListValue) Value() interface{}    { return v.Val }
func (v ObjectValue) Value() interface{}  { return v.Val }

// Node interface
type Node interface {
	node()
	// String() string
}

func (*Document) node()            {}
func (*OperationDefinition) node() {}
func (*FragmentDefinition) node()  {}
func (*SelectionSet) node()        {}
func (*Arguments) node()           {}
func (*Field) node()               {}
func (*Directive) node()           {}
func (*IntValue) node()            {}
func (*FloatValue) node()          {}
func (*StringValue) node()         {}
func (*EnumValue) node()           {}
func (*BooleanValue) node()        {}
func (*ListValue) node()           {}
func (*ObjectValue) node()         {}
