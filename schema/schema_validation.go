package schema

import (
	"log"

	"dylanmackenzie.com/graphql/ast"
)

// Schema validation happens once before the server is started and a
// GraphQL server cannot properly run with an invalid schema. Therefore
// if a schema fails to validate, we panic.

// Assert that every field name within fields is unique. Panics if the
// assertion fails.
func assertFieldsUnique(fields []ast.TypeField, desc string) {
	found := make(map[string]bool, len(fields))
	for _, field := range fields {
		if found[field.Name] {
			log.Panicf("Multiple fields named '%s' in '%s'", field.Name, desc)
		}
		found[field.Name] = true
	}
}

// Assert that an object implements an interface. Panics if the
// assertion fails.
func assertObjectImplements(obj *ast.ObjectDefinition, iface *ast.InterfaceDefinition) {
	for _, ifd := range iface.Fields {
		found := false
		for _, ofd := range obj.Fields {
			if ofd.Name == ifd.Name {
				found = true
				if ofd.Type.Name() != ifd.Type.Name() {
					log.Panicf(
						"Object field '%s' must be of type '%s', required by Interface '%s'",
						ofd.Name, ofd.Type.Name(), iface.Name)
				}
				assertArgumentDeclarationsCompatible(ofd, ifd)
				break
			}
		}

		if found == false {
			log.Panicf(
				"Object does not have field '%s', required by Interface '%s'",
				ifd.Name, iface.Name)
		}
	}
}

// Verifies that two fields have compatible arguments. Panics if they
// do not.
func assertArgumentsCompatible(f1, f2 ast.Field) {
	if len(f1.Arguments) != len(f2.Arguments) {
		log.Panicf(
			"Field '%s' has different argument arity than Field '%s'",
			f1.Name, f2.Name)
	}

	for i, arg1 := range f1.Arguments {
		arg2 := f2.Arguments[i]

		if arg1.Key != arg2.Key {
			log.Panicf(
				"Field '%s' has argument named '%s' of type '%s', but argument in Field '%s' is of type %s",
				f1.Name, i, arg1.Key, f2.Name, arg2.Key)
		}
	}
}

func assertArgumentDeclarationsCompatible(f1, f2 ast.TypeField) {
	if len(f1.Arguments) != len(f2.Arguments) {
		log.Panicf(
			"Field '%s' has different argument arity than Field '%s'",
			f1.Name, f2.Name)
	}

	for i, t1 := range f1.Arguments {
		t2 := f2.Arguments[i]

		if t1.Key != t2.Key {
			log.Panicf(
				"Field '%s' has argument named '%s' of type '%s', but argument in Field '%s' is of type %s",
				f1.Name, i, t1.Key, f2.Name, t2.Key)
		}
	}
}
