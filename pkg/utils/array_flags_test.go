package utils

import "testing"

func TestArrayFlags(t *testing.T) {
	var flag ArrayFlags
	Check(flag.Set("first"))
	Check(flag.Set("second"))
	if got := flag.String(); got != "first second" {
		t.Errorf("String() = %s; want %s", got, "first second")
	}
}
