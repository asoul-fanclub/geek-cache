package hsort_map

import (
	"strconv"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestHSortMap_GetAndPut(t *testing.T) {
	// 检测普通情况下Put方法
	var m HSortMap = NewHSkipList(func(a string) string {
		return a
	})
	m.Put("1", []byte("1"))
	m.Put("2", []byte("2"))
	m.Put("3", []byte("3"))

	a := assert.New(t)
	a.Equal([]byte("1"), m.Get("1"))
	a.Equal([]byte("2"), m.Get("2"))
	a.Equal([]byte("3"), m.Get("3"))

	// 含有相同hash
	m = NewHSkipList(func(a string) string {
		if a == "3" || a == "4" {
			return "2"
		}
		return a
	})
	for i := 0; i < 100; i++ {
		s := strconv.Itoa(i)
		m.Put(s, []byte(s))
	}
	for i := 0; i < 100; i++ {
		s := strconv.Itoa(i)
		a.Equal([]byte(s), m.Get(s))
	}

	// Put相同的值
	m.Put("2", []byte("5"))
	m.Put("3", []byte("6"))
	m.Put("4", []byte("7"))
	for i := 0; i < 100; i++ {
		s := strconv.Itoa(i)
		if s == "2" {
			a.Equal([]byte("5"), m.Get(s))
			continue
		}
		if s == "3" {
			a.Equal([]byte("6"), m.Get(s))
			continue
		}
		if s == "4" {
			a.Equal([]byte("7"), m.Get(s))
			continue
		}
		a.Equal([]byte(s), m.Get(s))
	}
}
