package hsort_map

import "testing"

func GetTest(t *testing.T) {
	var m HSortMap = NewHSkipList(func(a string) string {
		return a
	})
}
