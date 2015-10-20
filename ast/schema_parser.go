package ast

import (
	"errors"
	"reflect"
)

func parseObjectDefinition(def *ObjectDefinition, lex *lexer) error {
	if !lex.Expect(tokenIdent) {
		return errors.New("Expected name in type declaration")
	}

	// Name
	_, def.Name = lex.last()

	// Implements
	if lex.Optional(tokenColon) {
		cnt := 0
		for lex.Optional(tokenIdent) {
			cnt++
			_, iface := lex.last()
			def.Implements = append(def.Implements, iface)
		}

		if cnt == 0 {
			return errors.New("Implements list must have more than one name")
		}
	}

	if !lex.Expect(tokenLeftCurly) {
		return errors.New("Type declaration must have a body")
	}

	// Fields
	cnt := 0
	for lex.Optional(tokenIdent) {
		cnt++
		field := TypeField{}
		if err := parseTypeField(&field, lex); err != nil {
			return err
		}
		def.Fields = append(def.Fields, field)
	}

	if cnt == 0 {
		return errors.New("Type declaration must have at least one Field")
	}

	if !lex.Expect(tokenRightCurly) {
		return errors.New("Invalid field declaration")
	}

	return nil
}

func parseInterfaceDefinition(def *InterfaceDefinition, lex *lexer) error {
	if lex.Expect(tokenIdent) {
		_, def.Name = lex.last()
	} else {
		return errors.New("Expected name in interface declaration")
	}

	if !lex.Expect(tokenLeftCurly) {
		return errors.New("Interface declaration must have a body")
	}

	cnt := 0
	for lex.Optional(tokenIdent) {
		cnt++
		field := TypeField{}
		if err := parseTypeField(&field, lex); err != nil {
			return err
		}
		def.Fields = append(def.Fields, field)
	}

	if cnt == 0 {
		return errors.New("Interface declaration must have at least one Field")
	}

	if !lex.Expect(tokenRightCurly) {
		return errors.New("Invalid field declaration")
	}

	return nil
}

func parseEnumDefinition(def *EnumDefinition, lex *lexer) error {
	if !lex.Expect(tokenIdent) {
		return errors.New("Expected name in enum declaration")
	}

	_, def.Name = lex.last()

	if !lex.Expect(tokenLeftCurly) {
		return errors.New("Enum declaration must have a body")
	}

	cnt := 0
	for lex.Optional(tokenIdent) {
		_, ident := lex.last()
		if _, found := def.Values[ident]; found {
			return errors.New("Repeated value in enum")
		}

		def.Values[ident] = cnt
		cnt++
	}

	if cnt == 0 {
		return errors.New("Enum declaration must have at least one value")
	}

	if !lex.Expect(tokenRightCurly) {
		return errors.New("Invalid enum declaration")
	}

	return nil
}

func parseUnionDefinition(def *UnionDefinition, lex *lexer) error {
	if !lex.Expect(tokenIdent) {
		return errors.New("Expected name in enum declaration")
	}

	_, def.Name = lex.last()
	if !lex.Expect(tokenEqual) {
		return errors.New("Union declaration must contain list of members")
	}

	cnt := 0
	for lex.Expect(tokenIdent) {
		_, ident := lex.last()
		def.Members = append(def.Members, &BaseType{name: ident, nullable: false})
		cnt++

		if !lex.Optional(tokenPipe) {
			break
		}
	}

	return nil
}

func parseScalarDefinition(def *ScalarDefinition, lex *lexer) error {
	if !lex.Expect(tokenIdent) {
		return errors.New("Expected name of scalar declaration")
	}

	_, def.Name = lex.last()

	if !lex.Expect(tokenIdent) {
		return errors.New("Expected base type of new scalar")
	}

	switch _, lit := lex.last(); lit {
	case "Int":
		def.Kind = reflect.Int
	case "Float":
		def.Kind = reflect.Float64
	case "String":
		def.Kind = reflect.String
	case "Boolean":
		def.Kind = reflect.Bool
	default:
		return errors.New("Unknown base type for scalar")
	}

	return nil
}

func parseTypeField(field *TypeField, lex *lexer) error {
	// Sanity check
	if !lex.Assert(tokenIdent) {
		panic("parseTypeField called without name")
	}

	// Name
	_, field.Name = lex.last()

	// Arguments
	if lex.Optional(tokenLeftParen) {
		if err := parseArgumentDeclaration(&field.Arguments, lex); err != nil {
			return err
		}
	}

	// Colon
	if !lex.Expect(tokenColon) {
		return errors.New("TypeField must have a type")
	}

	// Type
	t, err := parseType(lex)
	if err != nil {
		return err
	}

	field.Type = t
	return nil
}

func parseArgumentDeclaration(args *ArgumentDeclarations, lex *lexer) error {
	// Sanity check
	if !lex.Assert(tokenLeftParen) {
		panic("parseArguments called outside of parens")
	}

	for {
		switch tok, lit := lex.Advance(); tok {
		case tokenIllegal:
			return errors.New(lit)
		case tokenEOF:
			return errors.New("Unexpected end of file")
		case tokenIdent:
			arg := &ArgumentDeclaration{Key: lit}

			if !lex.Expect(tokenColon) {
				return errors.New("ArgumentType key without type")
			}

			t, err := parseType(lex)
			if err != nil {
				return errors.New("Invalid type in ArgumentType")
			}

			arg.Type = t
			*args = append(*args, *arg)
		case tokenRightParen:
			lex.Discard() // Advance lexer to next token
			return nil
		default:
			return errors.New("Unexpected token")
		}
	}
}

func parseType(lex *lexer) (TypeDescriptor, error) {
	switch tok, lit := lex.Advance(); tok {
	case tokenIdent:
		t := &BaseType{name: lit}
		t.nullable = lex.Optional(tokenExclam)
		return t, nil

	case tokenLeftBracket:
		t := &ListType{}
		ofType, err := parseType(lex)
		if err != nil {
			return t, err
		}

		t.OfType = ofType
		if !lex.Expect(tokenRightBracket) {
			return t, errors.New("Unclosed List type")
		}

		t.nullable = lex.Optional(tokenExclam)
		return t, nil

	case tokenLeftCurly:
		t := &InputObjectType{}
		for {
			if lex.Optional(tokenRightCurly) {
				return t, nil
			}

			if !lex.Expect(tokenIdent) {
				return t, errors.New("InputObject type must have a key")
			}

			_, key := lex.last()
			if !lex.Expect(tokenColon) {
				return t, errors.New("InputObject key must have a type")
			}

			item, err := parseType(lex)
			if err != nil {
				return t, err
			}
			t.Fields[key] = item
		}

		if !lex.Expect(tokenRightCurly) {
			return t, errors.New("Unclosed InputObject type")
		}

		t.nullable = lex.Optional(tokenExclam)
		return t, nil

	default:
		return nil, errors.New("Invalid type")
	}
}
