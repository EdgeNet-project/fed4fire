package utils

import "testing"

func TestArrayFlags(t *testing.T) {
	var flag ArrayFlags
	flag.Set("first")
	flag.Set("second")
	if got := flag.String(); got != "first second" {
		t.Errorf("String() = %s; want %s", got, "first second")
	}
}
