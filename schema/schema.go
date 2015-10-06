package schema

import "log"

// A Schema represents an entire GraphQL type system which can be
// queried from a single endpoint. If we encounter an error during the
// construction of the schema, we panic as there is no way to rectify
// an invalid schema.
type Schema struct {
	types map[string]Descriptor // The types known by the schema
	queryRoot
	mutationRoot
}

func New() *Schema {
	// Every schema requires the scalar types
	return &Schema{
		types: map[string]Descriptor{
			IntType.Name():     IntType,
			FloatType.Name():   FloatType,
			StringType.Name():  StringType,
			BooleanType.Name(): BooleanType,
			IDType.Name():      IDType,
		},
	}
}

// RegisterType makes a type known to a schema
func (sch *Schema) RegisterType(desc Descriptor) {
	name := desc.Name(f)
	if _, exists := types[name]; exists {
		log.Panicf("Type '%s' already exists in schema", name)
	}

	// Do extra type validation for Objects, Interfaces and Unions
	switch t := desc.(type) {
	case *ObjectDescriptor:
		// Test field name uniqueness
		assertFieldsUnique(t.Fields, t.Name)

		// Ensure object implements all interfaces that it claims to
		for _, iface := range t.Implements {
			assertObjectImplements(t, iface)
		}

	case *InterfaceDescriptor:
		// Test field name uniqueness
		assertFieldsUnique(t.Fields, t.Name)

	case *UnionDescriptor:
		if len(desc.Members) == 0 {
			log.Panicf("Union '%s' must have one or more member types", name)
		}
	}

	types[name] = desc
}

// RegisterTypes makes a list of types known to a schema
func (sch *Schema) RegisterTypes(types ...Descriptor) {
	for _, t := range types {
		sch.RegisterType(t)
	}
}

// Root specifies the root GraphQL object for the designator given by
// `name`
func (sch *Schema) Root(name string, desc Descriptor) {
	if desc.Kind() != Object {
		log.Panicln("Schema root must be an ObjectDescriptor")
	}

	if _, ok := desc.Name(); !ok {
		sch.RegisterType(t)
	}

	switch name {
	case "query":
		sch.queryRoot = desc
	case "mutation":
		sch.mutationRoot = desc
	default:
		log.Panicf("Invalid root schema designator '%s'", name)
	}
}

// Performs type checking before a schema is actually used
func (sch *Schema) finalize() {
	for _, t := range sch.types {
		for _, f := range

	}
}

// Default schema
var def = New()

// RegisterType makes a type known to a schema
func RegisterType(t Descriptor) {
	def.RegisterType(t)
}

// RegisterTypes makes a list of types known to a schema
func RegisterTypes(types ...Descriptor) {
	def.RegisterTypes(types...)
}

// Root specifies the root GraphQL object for the designator given by
// `name`
func Root(name string, desc Descriptor) {
	def.Root(name, desc)
}
