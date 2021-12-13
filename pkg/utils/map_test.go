package utils

import (
	"reflect"
	"sort"
	"testing"
)

func TestKeys(t *testing.T) {
	m := map[string]string{"Hello": "A", "World": "B"}
	want := []string{"Hello", "World"}
	keys := Keys(m)
	sort.Strings(keys)
	if !reflect.DeepEqual(keys, want) {
		t.Errorf("Keys() = %v; want %v", keys, want)
	}
}
