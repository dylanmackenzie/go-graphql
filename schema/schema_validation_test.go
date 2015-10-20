package schema

import "testing"

func shouldPanic(name string, t *testing.T) {
	if r := recover(); r == nil {
		t.Errorf("Test '%s' should have panicked", name)
	}
}

func TestAssertObjectImplements(t *testing.T) {

}
