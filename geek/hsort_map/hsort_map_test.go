package hsort_map

import (
	"container/list"
	"strconv"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestHSortMap_GetAndPut(t *testing.T) {
	// 检测普通情况下Put方法
	var m HSortMap = NewHSkipList(func(a string) string {
		return a
	})
	m.Put("1", &list.Element{Value: []byte("1")})
	m.Put("2", &list.Element{Value: []byte("2")})
	m.Put("3", &list.Element{Value: []byte("3")})

	a := assert.New(t)
	a.Equal([]byte("1"), m.Get("1").Value)
	a.Equal([]byte("2"), m.Get("2").Value)
	a.Equal([]byte("3"), m.Get("3").Value)

	// 含有相同hash
	m = NewHSkipList(func(a string) string {
		if a == "3" || a == "4" {
			return "2"
		}
		return a
	})
	for i := 0; i < 100; i++ {
		s := strconv.Itoa(i)
		m.Put(s, &list.Element{Value: []byte(s)})
	}
	for i := 0; i < 100; i++ {
		s := strconv.Itoa(i)
		a.Equal([]byte(s), m.Get(s).Value)
	}

	// Put相同的值
	m.Put("2", &list.Element{Value: []byte("5")})
	m.Put("3", &list.Element{Value: []byte("6")})
	m.Put("4", &list.Element{Value: []byte("7")})
	for i := 0; i < 100; i++ {
		s := strconv.Itoa(i)
		if s == "2" {
			a.Equal([]byte("5"), m.Get(s).Value)
			continue
		}
		if s == "3" {
			a.Equal([]byte("6"), m.Get(s).Value)
			continue
		}
		if s == "4" {
			a.Equal([]byte("7"), m.Get(s).Value)
			continue
		}
		a.Equal([]byte(s), m.Get(s).Value)
	}
}

func TestHSkipList_Delete(t *testing.T) {
	a := assert.New(t)
	// 检测删除
	var m HSortMap = NewHSkipList(func(a string) string {
		n, _ := strconv.Atoi(a)
		if n >= 10 && n < 20 {
			return "10"
		}
		return a
	})
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
			a.Equal(m.Get(key).Value, []byte(key))
		}
	}
}

func TestHSkipList_Exist(t *testing.T) {
	a := assert.New(t)
	var m HSortMap = NewHSkipList(func(a string) string {
		return a
	})
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

func TestHSkipList_DeleteByHashRange(t *testing.T) {
	a := assert.New(t)
	var m HSortMap = NewHSkipList(func(a string) string {
		return a
	})
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
