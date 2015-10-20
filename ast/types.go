package ast

import "reflect"

func GetBaseType(desc TypeDescriptor) *BaseType {
	switch t := desc.(type) {
	case *BaseType:
		return t
	case *ListType:
		return GetBaseType(t.OfType)
	}

	return nil
}

func IsAbstractType(def TypeDefinition) bool {
	switch def.(type) {
	case *EnumDefinition, *ScalarDefinition:
		return false
	default:
		return true
	}
}

// Type Coercion

func IsOfType(v Value, def TypeDefinition) bool {
	if IsAbstractType(def) {
		return false
	}

	switch value := v.(type) {
	case IntValue:
		t, ok := def.(*ScalarDefinition)
		return ok && t.Kind == reflect.Int
	case FloatValue:
		t, ok := def.(*ScalarDefinition)
		return ok && t.Kind == reflect.Float64

	case StringValue:
		t, ok := def.(*ScalarDefinition)
		return ok && t.Kind == reflect.String

	case BooleanValue:
		t, ok := def.(*ScalarDefinition)
		return ok && t.Kind == reflect.Bool

	case EnumValue:
		t, ok := def.(*EnumDefinition)
		if !ok {
			return false
		}

		_, ok = t.Values[string(value)]
		return ok

	case ListValue:
		return IsOfType(value[0], def)

	case ObjectValue:
		return false
	}

	return true
}
