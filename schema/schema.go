package schema

import (
	"log"
	"reflect"
	"strings"

	"dylanmackenzie.com/graphql/ast"
)

// A Schema represents an entire GraphQL type system which can be
// queried from a single endpoint. If we encounter an error during the
// construction of the schema, we panic as there is no way to rectify
// an invalid schema.
type Schema struct {
	types        map[string]ast.TypeDefinition // The types known by the schema
	resolvers    map[string]Resolver           // The resolvers
	QueryRoot    *ast.ObjectDefinition
	MutationRoot *ast.ObjectDefinition

	mutable bool // Flag set to false after the schema has been finalized
}

func New() *Schema {
	// Every schema requires the scalar types
	return &Schema{
		resolvers: make(map[string]Resolver),
		types: map[string]ast.TypeDefinition{
			"Int":     &ast.ScalarDefinition{Name: "Int", Kind: reflect.Int},
			"Float":   &ast.ScalarDefinition{Name: "Float", Kind: reflect.Float64},
			"String":  &ast.ScalarDefinition{Name: "String", Kind: reflect.String},
			"Boolean": &ast.ScalarDefinition{Name: "Boolean", Kind: reflect.Bool},
			"ID":      &ast.ScalarDefinition{Name: "ID", Kind: reflect.String},
		},
		mutable: true,
	}
}

func (sch *Schema) resolver(name string) Resolver {
	res, ok := sch.resolvers[name]
	if !ok {
		log.Panicf("No resolver for type '%s' found", name)
	}

	return res
}

func (sch *Schema) AddResolver(name string, res Resolver) {
	def, ok := sch.types[name]
	if !ok {
		log.Panicf("No type named '%s' found", name)
	}

	_, ok = def.(ast.AbstractTypeDefinition)
	if !ok {
		log.Panicf("Attempting to add resolver to non-abstract type '%s'", name)
	}

	sch.resolvers[name] = res
}

func (sch *Schema) AddResolveFunc(name string, res ResolveFunc) {
	sch.AddResolver(name, Resolver(res))
}

// definition takes a result type and finds its definition in
// the schema. It panics if the referenced type is not found.
func (sch *Schema) definition(t *ast.BaseType) ast.TypeDefinition {
	def, ok := sch.types[t.Name()]
	if !ok {
		log.Panicf("Type '%s' not found in schema", t.Name)
	}
	return def
}

// verify ensures that for a given type, every type it references exists
// in the schema.
func (sch *Schema) verify(desc ast.TypeDescriptor) {
	switch t := desc.(type) {
	case *ast.BaseType:
		name := t.Name()
		if _, ok := sch.types[name]; !ok {
			log.Panicf("Type '%s' not found in schema", name)
		}

	case *ast.ListType:
		sch.verify(t.OfType)

	case *ast.InputObjectType:
		for _, sub := range t.Fields {
			sch.verify(sub)
		}
	default:
		panic("verify called on invalid type")
	}

}

// finalize ensures that every type referenced in the schema actually
// exists in the schema. It is called once all types have been added to
// the schema but before the schema is used. Once the type checking is
// complete, further mutations are prevented from occurring on the
// schema.
func (sch *Schema) Finalize() {
	sch.mutable = false

	if sch.QueryRoot == nil {
		panic("Schema must provide a root object for queries. Call schema.Root(\"query\", name).")
	}

	for _, def := range sch.types {
		switch t := def.(type) {
		case *ast.ObjectDefinition:
			for i, field := range t.Fields {
				sch.verify(field.Type)

				// Cache pointer to definition in TypeField
				t.Fields[i].Definition = sch.definition(ast.GetBaseType(field.Type))

				for _, arg := range field.Arguments {
					sch.verify(arg.Type)
				}
			}

			// Ensure object implements all interfaces which it claims
			for _, name := range t.Implements {
				iface, ok := sch.types[name]
				if !ok {
					log.Panicf("Interface '%s' not found in type system", name)
				}

				// Panics if iface is not an interface definition
				assertObjectImplements(t, iface.(*ast.InterfaceDefinition))
			}

		case *ast.InterfaceDefinition:
			for _, field := range t.Fields {
				sch.verify(field.Type)
				for _, arg := range field.Arguments {
					sch.verify(arg.Type)
				}
			}

		case *ast.UnionDefinition:
			for _, member := range t.Members {
				sch.verify(member)
			}

		case *ast.ScalarDefinition:
			switch t.Kind {
			case reflect.Int, reflect.Bool, reflect.Float64, reflect.String:
				continue
			default:
				panic("ScalarDefinition has invalid underlying type")
			}

		case *ast.EnumDefinition:
			// All enums are of type Int
			continue

		default:
			panic("finalize called on invalid type")
		}
	}
}

func (sch *Schema) AddDocument(doc *ast.Document) {
	if !sch.mutable {
		panic("Attempted to mutate schema after it has been finalized")
	}

	for _, def := range doc.Definitions {
		t, ok := def.(ast.TypeDefinition)
		if !ok {
			log.Panic("Document for schema must consist of only type definitions\n")
		}

		if strings.HasPrefix(t.TypeName(), "__") {
			log.Panic("Type names cannot start with '__'\n")
		}

		sch.addType(t)
	}
}

// AddType makes a type known to a schema
func (sch *Schema) addType(def ast.TypeDefinition) {
	if !sch.mutable {
		panic("Attempted to mutate schema after it has been finalized")
	}

	var name string

	// Do extra type validation for Objects, Interfaces and Unions
	switch t := def.(type) {
	case *ast.ScalarDefinition:
		name = t.Name
	case *ast.EnumDefinition:
		name = t.Name
	case *ast.ObjectDefinition:
		name = t.Name

		assertFieldsUnique(t.Fields, t.Name)

	case *ast.InterfaceDefinition:
		name = t.Name
		assertFieldsUnique(t.Fields, t.Name)

	case *ast.UnionDefinition:
		name = t.Name
		if len(t.Members) == 0 {
			log.Panicf("Union '%s' must have one or more member types", name)
		}
	}

	if _, exists := sch.types[name]; exists {
		log.Panicf("Type '%s' already exists in schema", name)
	}

	sch.types[name] = def
}

// Root specifies the root GraphQL object for the designator given by
// `name`
func (sch *Schema) Root(rootName, typeName string) {
	if !sch.mutable {
		panic("Attempted to mutate schema after it has been finalized")
	}

	t, ok := sch.types[typeName]
	if !ok {
		log.Panicf("No type named '%s' found", typeName)
	}

	switch rootName {
	case "query":
		if sch.QueryRoot, ok = t.(*ast.ObjectDefinition); !ok {
			panic("Root type must be an object")
		}
	case "mutation":
		if sch.MutationRoot, ok = t.(*ast.ObjectDefinition); !ok {
			panic("Root type must be an object")
		}
	default:
		log.Panicf("Invalid root schema designator '%s'", rootName)
	}
}

// Default schema
var def = New()

// AddType makes a type known to a schema
func addType(t ast.TypeDefinition) {
	def.addType(t)
}

// Root specifies the root GraphQL object for the designator given by
// `name`
func Root(rootName, typeName string) {
	def.Root(rootName, typeName)
}

func AddResolver(name string, res Resolver) {
	def.AddResolver(name, res)
}

func AddResolveFunc(name string, res ResolveFunc) {
	def.AddResolveFunc(name, res)
}
