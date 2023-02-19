package consistenthash

import (
	"strconv"
	"testing"
)

func TestMap_Get(t *testing.T) {
	hash := New(Replicas(3), HashFunc(func(data []byte) uint32 {
		i, _ := strconv.Atoi(string(data))
		return uint32(i)
	}))
	// add vir-node 06, 16, 26, 04, 14, 24, 02, 12, 22
	hash.Add("6", "4", "2")
	testCases := map[string]string{
		"2":  "2", // 02 - 2
		"11": "2", // 12 - 2
		"23": "4", // 24 - 4
		"26": "6", // 24 - 4
		"24": "4", // 24 - 4
		"27": "2", // 02 - 2
	}
	for k, v := range testCases {
		if hash.Get(k) != v {
			t.Errorf("hash.Get(%s) expeted %s, but %s", k, v, hash.Get(k))
		}
	}
	// add vir-node 08, 18, 28
	hash.Add("8")
	testCases["27"] = "8" // 28 - 8
	for k, v := range testCases {
		if hash.Get(k) != v {
			t.Errorf("hash.Get(%s) expeted %s, but %s", k, v, hash.Get(k))
		}
	}
	// remove vir-node 08, 18, 28
	hash.Remove("8")
	testCases["27"] = "2" // 02 - 2
	for k, v := range testCases {
		if hash.Get(k) != v {
			t.Errorf("hash.Get(%s) expeted %s, but %s", k, v, hash.Get(k))
		}
	}
}
