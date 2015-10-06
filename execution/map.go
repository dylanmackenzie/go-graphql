package execution

// A simple wrapper around a map that allows struct embedding and a
// convenient syntax for converting a value to a given type.
type untypedMap map[string]interface{}

// Sets key to value.
func (m untypedMap) Set(key string, value interface{}) {
	m[key] = value
}

// Gets `key` from map. Returns false as its second argument if no
// value was present.
func (m untypedMap) Get(key string) (interface{}, bool) {
	return m[key]
}

// Gets `key` from map and converts it to a string. Returns false as
// its second argument if no value was present or the value was not of
// the proper type.
func (m untypedMap) GetAsString(key string) (ret string, ok bool) {
	if v, found := m[key]; found {
		ret, ok = v.(string)
	}
	return
}

// Gets `key` from map and converts it to a slice of strings. Returns
// false as its second argument if no value was present or the value
// was not of the proper type.
func (m untypedMap) GetAsStrings(key string) (ret []string, ok bool) {
	if v, found := m[key]; found {
		ret, ok = v.([]string)
	}
	return
}

// Gets `key` from map and converts it to an int. Returns false as its
// second argument if no value was present or the value was not of the
// proper type.
func (m untypedMap) GetAsInt(key string) (ret int, ok bool) {
	if v, found := m[key]; found {
		ret, ok = v.(int)
	}
	return
}

// Gets `key` from map and converts it to a slice of ints. Returns false
// as its second argument if no value was present or the value was not
// of the proper type.
func (m untypedMap) GetAsInts(key string) (ret []int, ok bool) {
	if v, found := m[key]; found {
		ret, ok = v.([]int)
	}
	return
}

// Gets `key` from map and converts it to a float. Returns false as a
// its second argument if no value was present or the value was not of
// the proper type.
func (m untypedMap) GetAsFloat(key string) (ret float64, ok bool) {
	if v, found := m[key]; found {
		ret, ok = v.(float64)
	}
	return
}

// Gets `key` from map and converts it to a slice of floats. Returns
// false as its second argument if no value was present or the value
// was not of the proper type.
func (m untypedMap) GetAsFloats(key string) (ret []float64, ok bool) {
	if v, found := m[key]; found {
		ret, ok = v.([]float64)
	}
	return
}

// Gets `key` from map and converts it to a bool. Returns false as its
// second argument if no value was present or the value was not of the
// proper type.
func (m untypedMap) GetAsBool(key string) (ret bool, ok bool) {
	if v, found := m[key]; found {
		ret, ok = v.(bool)
	}
	return
}

// Gets `key` from map and converts it to a slice of bools. Returns
// false as its second argument if no value was present or the value
// was not of the proper type.
func (m untypedMap) GetAsBools(key string) (ret []bool, ok bool) {
	if v, found := m[key]; found {
		ret, ok = v.([]bool)
	}
	return
}
