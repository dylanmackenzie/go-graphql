package schema

// A simple wrapper around a map that allows struct embedding and a
// convenient syntax for converting a value to a given type.
type resultMap map[string]interface{}

// Sets key to value.
func (m resultMap) Set(key string, value interface{}) {
	m[key] = value
}

// Gets `key` from map. Returns false as its second argument if no
// value was present.
func (m resultMap) Get(key string) (val interface{}, ok bool) {
	val, ok = m[key]
	return
}

// Gets `key` from map and converts it to a string. Returns false as
// its second argument if no value was present or the value was not of
// the proper type.
func (m resultMap) GetAsString(key string) (ret string, ok bool) {
	if v, found := m[key]; found {
		ret, ok = v.(string)
	}
	return
}

// Gets `key` from map and converts it to a slice of strings. Returns
// false as its second argument if no value was present or the value
// was not of the proper type.
func (m resultMap) GetAsStrings(key string) (ret []string, ok bool) {
	if v, found := m[key]; found {
		ret, ok = v.([]string)
	}
	return
}

// Gets `key` from map and converts it to an int. Returns false as its
// second argument if no value was present or the value was not of the
// proper type.
func (m resultMap) GetAsInt(key string) (ret int, ok bool) {
	if v, found := m[key]; found {
		ret, ok = v.(int)
	}
	return
}

// Gets `key` from map and converts it to a slice of ints. Returns false
// as its second argument if no value was present or the value was not
// of the proper type.
func (m resultMap) GetAsInts(key string) (ret []int, ok bool) {
	if v, found := m[key]; found {
		ret, ok = v.([]int)
	}
	return
}

// Gets `key` from map and converts it to a float. Returns false as a
// its second argument if no value was present or the value was not of
// the proper type.
func (m resultMap) GetAsFloat(key string) (ret float64, ok bool) {
	if v, found := m[key]; found {
		ret, ok = v.(float64)
	}
	return
}

// Gets `key` from map and converts it to a slice of floats. Returns
// false as its second argument if no value was present or the value
// was not of the proper type.
func (m resultMap) GetAsFloats(key string) (ret []float64, ok bool) {
	if v, found := m[key]; found {
		ret, ok = v.([]float64)
	}
	return
}

// Gets `key` from map and converts it to a bool. Returns false as its
// second argument if no value was present or the value was not of the
// proper type.
func (m resultMap) GetAsBool(key string) (ret bool, ok bool) {
	if v, found := m[key]; found {
		ret, ok = v.(bool)
	}
	return
}

// Gets `key` from map and converts it to a slice of bools. Returns
// false as its second argument if no value was present or the value
// was not of the proper type.
func (m resultMap) GetAsBools(key string) (ret []bool, ok bool) {
	if v, found := m[key]; found {
		ret, ok = v.([]bool)
	}
	return
}
