package hsort_map

import (
	"container/list"
	"strconv"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestHSortMap_GetAndPut(t *testing.T) {
	for i := 0; i < 2; i++ {
		var m HSortMap
		hash := func(a string) string {
			return a
		}
		if i == 0 {
			m = NewGeekMap(hash)
		} else {
			m = NewHSkipList(hash)
		}
		m.Put("1", &list.Element{Value: []byte("1")})
		m.Put("2", &list.Element{Value: []byte("2")})
		m.Put("3", &list.Element{Value: []byte("3")})

		a := assert.New(t)
		v1, b1 := m.Get("1")
		v2, b2 := m.Get("2")
		v3, b3 := m.Get("3")
		a.Equal([]byte("1"), v1.Value)
		a.Equal([]byte("2"), v2.Value)
		a.Equal([]byte("3"), v3.Value)
		a.True(b1 && b2 && b3)

		// 含有相同hash
		hash2 := func(a string) string {
			if a == "3" || a == "4" {
				return "2"
			}
			return a
		}
		if i == 0 {
			m = NewHSkipList(hash2)
		} else {
			m = NewGeekMap(hash2)
		}
		for i := 0; i < 100; i++ {
			s := strconv.Itoa(i)
			m.Put(s, &list.Element{Value: []byte(s)})
		}
		for i := 0; i < 100; i++ {
			s := strconv.Itoa(i)
			v, b := m.Get(s)
			a.Equal([]byte(s), v.Value)
			a.True(b)
		}

		// Put相同的值
		m.Put("2", &list.Element{Value: []byte("5")})
		m.Put("3", &list.Element{Value: []byte("6")})
		m.Put("4", &list.Element{Value: []byte("7")})
		for i := 0; i < 100; i++ {
			s := strconv.Itoa(i)
			if s == "2" || s == "3" || s == "4" {
				num, _ := strconv.Atoi(s)
				s2 := strconv.Itoa(num + 3)
				v, b := m.Get(s)
				a.Equal([]byte(s2), v.Value)
				a.True(b)
				continue
			}
			v, b := m.Get(s)
			a.Equal([]byte(s), v.Value)
			a.True(b)
		}
	}

}

func TestHSkipList_Delete(t *testing.T) {
	a := assert.New(t)
	// 检测删除
	for i := 0; i < 2; i++ {
		var m HSortMap
		hash := func(a string) string {
			n, _ := strconv.Atoi(a)
			if n >= 10 && n < 20 {
				return "10"
			}
			return a
		}
		if i == 0 {
			m = NewGeekMap(hash)
		} else {
			m = NewHSkipList(hash)
		}
		for i := 0; i < 100; i++ {
			s := strconv.Itoa(i)
			m.Put(s, &list.Element{Value: []byte(s)})
		}
		for i := 0; i < 100; i += 3 {
			m.Delete(strconv.Itoa(i))
		}
		for i := 0; i < 100; i++ {
			key := strconv.Itoa(i)
			if i%3 == 0 {
				a.Nil(m.Get(key))
			} else {
				v, b := m.Get(key)
				a.Equal(v.Value, []byte(key))
				a.True(b)
			}
		}
	}
}

func TestHSkipList_Exist(t *testing.T) {
	a := assert.New(t)
	for i := 0; i < 2; i++ {
		var m HSortMap
		hash := func(a string) string {
			return a
		}
		if i == 0 {
			m = NewGeekMap(hash)
		} else {
			m = NewHSkipList(hash)
		}
		for i := 0; i < 100; i += 2 {
			s := strconv.Itoa(i)
			m.Put(s, &list.Element{Value: []byte(s)})
		}
		for i := 0; i < 100; i++ {
			key := strconv.Itoa(i)
			if i%2 != 0 {
				a.False(m.Exist(key))
			} else {
				a.True(m.Exist(key))
			}
		}
	}
}

func TestHSkipList_DeleteByHashRange(t *testing.T) {
	a := assert.New(t)
	for i := 0; i < 2; i++ {
		var m HSortMap
		hash := func(a string) string {
			return a
		}
		if i == 0 {
			m = NewGeekMap(hash)
		} else {
			m = NewHSkipList(hash)
		}
		for i := 0; i < 100; i++ {
			s := strconv.Itoa(i)
			m.Put(s, &list.Element{Value: []byte(s)})
		}
		l := m.DeleteByHashRange("10", "20")
		for i := 0; i < 100; i++ {
			s := strconv.Itoa(i)
			if s >= "10" && s < "20" {
				l--
				a.False(m.Exist(s))
			} else {
				a.True(m.Exist(s))
			}
		}
		a.Equal(l, 0)
	}
}
