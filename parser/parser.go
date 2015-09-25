package parse

import (
	"errors"
	"io"
	"strconv"
)

func Reader(r io.Reader) (Document, error) {
	lex := newLexer(r)
	doc := Document{}

	// If we have a shorthand document, we immediately parse the
	// first definition
	if lex.Optional(tokenLeftCurly) {
		def := &OperationDefinition{
			Name:   "",
			OpType: QUERY,
		}
		doc.Definitions = append(doc.Definitions, def)
		return doc, parseSelectionSet(&def.SelectionSet, lex)
	}

	// Otherwise we have a normal document
	for lex.Optional(tokenName) {
		switch _, lit := lex.last(); lit {
		case "query", "mutation":
			def := &OperationDefinition{}
			doc.Definitions = append(doc.Definitions, def)
			if err := parseOperationDefinition(def, lex); err != nil {
				return doc, err
			}
		case "fragment":
			def := &FragmentDefinition{}
			doc.Definitions = append(doc.Definitions, def)
			if err := parseFragmentDefinition(def, lex); err != nil {
				return doc, err
			}
		default:
			return doc, errors.New("Beginning of definition not one of query, mutation or fragment")
		}
	}

	// If we didn't find a name or EOF, throw an error
	if tok, _ := lex.Advance(); tok != tokenEOF {
		return doc, errors.New("Start of definition not a name")
	}

	return doc, nil
}

func parseOperationDefinition(def *OperationDefinition, lex *lexer) error {
	if !lex.Assert(tokenName) {
		panic("ParseOperationDefinition called without a name token")
	}

	// Operation Type
	switch _, opType := lex.last(); opType {
	case "query":
		def.OpType = QUERY
	case "mutation":
		def.OpType = MUTATION
	default:
		return errors.New("parseOperationDefinition called with invalid OperationType")
	}

	// Name
	if lex.Expect(tokenName) {
		_, def.Name = lex.last()
	} else {
		return errors.New("Expected Name in Operation Definition")
	}

	// Variable Definitions
	if lex.Optional(tokenLeftParen) {
		if err := parseVariableDefinitions(&def.Variables, lex); err != nil {
			return err
		}
	}

	// Directives
	if lex.Optional(tokenAt) {
		if err := parseDirectives(&def.Directives, lex); err != nil {
			return err
		}
	}

	// Selection Set
	if lex.Expect(tokenLeftCurly) {
		return parseSelectionSet(&def.SelectionSet, lex)
	} else {
		return errors.New("Operation Definition must have a selection set")
	}
}

func parseFragmentDefinition(def *FragmentDefinition, lex *lexer) error {
	if !lex.Assert(tokenName) {
		panic("parseFragmentDefinition called without name")
	}

	// Determine if we're parsing on inline fragment or not
	if _, lit := lex.last(); lit == "on" {
		// If we're parsing an inline fragment, we don't have a name so
		// skip directly to the type
		goto inline
	} else if lit != "fragment" {
		panic("parseFragmentDefinition must be called with 'fragment' or 'on'")
	}

	// Name
	if lex.Expect(tokenName) {
		_, def.Name = lex.last()
	} else {
		return errors.New("No name for fragment")
	}

	if tok, lit := lex.Advance(); tok != tokenName || lit != "on" {
		return errors.New("Fragment name must be followed by 'on'")
	}

inline:

	// Type
	if lex.Expect(tokenName) {
		_, def.Type = lex.last()
	} else {
		return errors.New("Fragment definition must be on a type")
	}

	// Directives
	if lex.Optional(tokenAt) {
		if err := parseDirectives(&def.Directives, lex); err != nil {
			return err
		}
	}

	// Selection Set
	if lex.Expect(tokenLeftCurly) {
		return parseSelectionSet(&def.SelectionSet, lex)
	} else {
		return errors.New("Fragment Definition must have a Selection Set")
	}
}

func parseSelectionSet(set *SelectionSet, lex *lexer) error {
	// Sanity check
	if !lex.Assert(tokenLeftCurly) {
		panic("parseSelectionSet called outside of block")
	}

	for {
		switch tok, lit := lex.Advance(); tok {
		case tokenIllegal:
			return errors.New(lit)
		case tokenEOF:
			return errors.New("Unclosed selection set")
		case tokenName:
			field := &Field{}
			*set = append(*set, field)
			if err := parseField(field, lex); err != nil {
				return err
			}
			tok, lit = lex.last()
		case tokenSpread:
			var frag Selection
			frag = &FragmentSpread{}

			// Determine if we have a fragment spread or an inline fragment
			err := parseFragmentSpread(frag.(*FragmentSpread), lex)
			if err != nil {
				if err.Error() != "Found an InlineFragment instead of a FragmentSpread" {
					return err
				}

				frag = &FragmentDefinition{}
				if err := parseFragmentDefinition(frag.(*FragmentDefinition), lex); err != nil {
					return err
				}
			}

			*set = append(*set, frag)
		case tokenRightCurly:
			lex.Discard() // Advance lexer to next token
			return nil
		default:
			return errors.New("Unexpected token")
		}
	}
}

func parseField(field *Field, lex *lexer) error {
	// Sanity check
	if !lex.Assert(tokenName) {
		panic("parseField called without name")
	}

	// Name or Alias
	_, field.Name = lex.last()
	if lex.Optional(tokenColon) {
		field.Alias = field.Name
		if lex.Expect(tokenName) {
			_, field.Name = lex.last()
		} else {
			return errors.New("Alias without a name")
		}
	}

	// Arguments
	if lex.Optional(tokenLeftParen) {
		if err := parseArguments(&field.Arguments, lex); err != nil {
			return err
		}
	}

	// Directives
	if lex.Optional(tokenAt) {
		if err := parseDirectives(&field.Directives, lex); err != nil {
			return err
		}
	}

	// Selection Set
	if lex.Optional(tokenLeftCurly) {
		if err := parseSelectionSet(&field.SelectionSet, lex); err != nil {
			return err
		}
	}

	return nil
}

func parseFragmentSpread(frag *FragmentSpread, lex *lexer) error {
	// Sanity check
	if !lex.Assert(tokenSpread) {
		panic("parseFragmentSpread called without spread operator")
	}

	// Name
	if !lex.Expect(tokenName) {
		return errors.New("Invalid Fragment Name in spread")
	}

	// 'on' is not a valid FragmentName
	if _, lit := lex.last(); lit != "on" {
		_, frag.Name = lex.last()
	} else {
		return errors.New("Found an InlineFragment instead of a FragmentSpread")
	}

	// Directives
	if lex.Optional(tokenAt) {
		return parseDirectives(&frag.Directives, lex)
	}

	return nil
}

func parseArguments(args *Arguments, lex *lexer) error {
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
		case tokenName:
			arg := &Argument{Key: lit}

			if !lex.Expect(tokenColon) {
				return errors.New("Argument key without value")
			}

			v, err := parseValue(lex)
			if err != nil {
				return err
			}

			arg.Value = v
			*args = append(*args, *arg)
		case tokenRightParen:
			lex.Discard() // Advance lexer to next token
			return nil
		default:
			return errors.New("Unexpected token")
		}
	}
}

func parseVariableDefinitions(vars *Variables, lex *lexer) error {
	// Sanity check
	if !lex.Assert(tokenLeftParen) {
		panic("parseVariableDefinitions called outside of parens")
	}

	for {
		switch tok, lit := lex.Advance(); tok {
		case tokenIllegal:
			return errors.New(lit)
		case tokenEOF:
			return errors.New("Unexpected end of file")
		case tokenVariableValue:
			v := &Variable{Nullable: true}
			_, v.Name = lex.last()

			// Type
			if !lex.Expect(tokenColon) {
				return errors.New("Variable without type")
			}

			// Type name
			switch tok, lit := lex.Advance(); tok {
			case tokenName:
				v.Type = lit
			case tokenLeftBracket:
				return errors.New("TODO: List not yet supported")
			default:
				return errors.New("Variable without type")
			}

			// Non null variable check
			if lex.Optional(tokenExclam) {
				v.Nullable = false
			}

			// Default Value(Optional)
			if lex.Optional(tokenEqual) {
				def, err := parseValue(lex)
				if err != nil {
					return err
				}
				v.Default = def
			}

			*vars = append(*vars, *v)
		case tokenRightParen:
			lex.Discard() // Advance lexer to next token
			return nil
		default:
			return errors.New("Unexpected token")
		}
	}
}

func parseDirectives(dirs *Directives, lex *lexer) error {
	// sanity check
	if !lex.Assert(tokenAt) {
		panic("parseDirectives called without at symbol")
	}

	for lex.Assert(tokenAt) {
		dir := &Directive{}

		// Name
		if lex.Expect(tokenName) {
			_, dir.Name = lex.last()
		} else {
			return errors.New("Expected Name in Operation Definition")
		}

		// Arguments
		if lex.Optional(tokenLeftParen) {
			if err := parseArguments(&dir.Arguments, lex); err != nil {
				return err
			}
		}

		*dirs = append(*dirs, *dir)
	}

	return nil
}

func parseValue(lex *lexer) (Value, error) {
	switch tok, lit := lex.Advance(); tok {
	case tokenIntValue:
		num, err := strconv.Atoi(lit)
		if err != nil {
			return nil, errors.New("Invalid integer literal")
		}
		return IntValue(num), nil
	case tokenFloatValue:
		num, err := strconv.ParseFloat(lit, 64)
		if err != nil {
			return nil, errors.New("Invalid integer literal")
		}
		return FloatValue(num), nil
	case tokenStringValue:
		return StringValue(lit), nil
	case tokenVariableValue:
		return VariableValue(lit), nil
	case tokenName:
		if lit == "true" || lit == "false" {
			return BooleanValue(lit == "true"), nil
		} else if lit == "null" {
			return nil, errors.New("Value cannot be null")
		} else {
			return EnumValue(lit), nil
		}
	case tokenLeftCurly:
		obj, err := parseObjectValue(lex)
		if err != nil {
			return nil, err
		}
		return obj, nil
	case tokenLeftBracket:
		list, err := parseListValue(lex)
		if err != nil {
			return nil, err
		}
		return list, nil
	default:
		return nil, errors.New("Invalid value")
	}
}

func parseListValue(lex *lexer) (ListValue, error) {
	val := ListValue{}

	for {
		if lex.Optional(tokenRightBracket) {
			return val, nil
		}

		item, err := parseValue(lex)
		if err != nil {
			return val, err
		}
		val = append(val, item)
	}
}

func parseObjectValue(lex *lexer) (ObjectValue, error) {
	var key string
	val := ObjectValue{}

	for {
		if lex.Optional(tokenRightCurly) {
			return val, nil
		}

		if !lex.Expect(tokenName) {
			return val, errors.New("ObjectValue must have a key")
		}

		_, key = lex.last()
		if !lex.Expect(tokenColon) {
			return val, errors.New("ObjectValue must have a value")
		}

		item, err := parseValue(lex)
		if err != nil {
			return val, err
		}
		val[key] = item
	}
}
