package schema

import (
	"strings"
	"testing"

	"dylanmackenzie.com/graphql/ast"
)

var schema = `
enum DogCommand { SIT, DOWN, HEEL }

type Dog : Pet {
  name: String!
  nickname: String
  barkVolume: Int
  doesKnowCommand(dogCommand: DogCommand!) : Boolean!
  isHouseTrained(atOtherHomes: Boolean): Boolean!
}

interface Sentient {
  name: String!
}

interface Pet {
  name: String!
}

type Alien : Sentient {
  name: String!
  homePlanet: String
}

type Human : Sentient {
  name: String!
}

type Cat : Pet {
  name: String!
  nickname: String
  meowVolume: Int
}

union CatOrDog = Cat | Dog
union DogOrHuman = Dog | Human
union HumanOrAlien = Human | Alien
`

var result = map[string][]string{
	"DogCommand":   nil,
	"Dog":          {"name", "nickname", "barkVolume", "doesKnowCommand", "isHouseTrained"},
	"Sentient":     {"name"},
	"Pet":          {"name"},
	"Alien":        {"name", "homePlanet"},
	"Human":        {"name"},
	"Cat":          {"name", "nickname", "meowVolume"},
	"CatOrDog":     nil,
	"DogOrHuman":   nil,
	"HumanOrAlien": nil,
}

func TestAddDocument(t *testing.T) {
	defer func() {
		if r := recover(); r != nil {
			t.Error(r)
		}
	}()

	doc, err := ast.FromReader(strings.NewReader(schema))
	if err != nil {
		t.Error(err)
	}

	sch := New()
	sch.AddDocument(&doc)
	sch.Finalize()

	for name, fields := range result {
		ty, ok := sch.types[name]
		if !ok {
			t.Errorf("Expected to find type '%s'", name)
		}

		if fields == nil {
			continue
		}

		for _, fieldName := range fields {
			if _, ok := ty.(ast.AbstractTypeDefinition).Field(fieldName); !ok {
				t.Errorf("Expected type '%s' to have field '%s'", name, fieldName)
			}
		}
	}
}
