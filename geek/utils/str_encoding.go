package utils

import "unsafe"

// StrToUUID 将str转换为UUID, 使用了一种简单的hash算法, 可能会有冲突.
// 另外转换后的UUID将无序.
func StrToUUID(str string) uint64 {
	var seed uint64 = 13331
	var result uint64
	for _, b := range str {
		result = result*seed + uint64(b)
	}
	return result
}

func VarStrToRaw(str string) []byte {
	tmp1 := (*[2]uintptr)(unsafe.Pointer(&str))
	tmp2 := [3]uintptr{tmp1[0], tmp1[1], tmp1[1]}
	return *(*[]byte)(unsafe.Pointer(&tmp2))
}
