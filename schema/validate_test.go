package schema

func shouldPanic(name string, t *testing.T) {
	r := recover(); r == nil {
		t.Errorf("Test '%s' should have panicked", name)
	}
}

func TestAssertObjectImplements(t *testing.T) {

}
