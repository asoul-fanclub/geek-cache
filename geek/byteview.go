package geek

// ByteView 只读的字节视图，用于缓存数据
type ByteView struct {
	B []byte
}

func (b ByteView) Len() int {
	return len(b.B)
}

func (b ByteView) ByteSLice() []byte {
	return cloneBytes(b.B)
}

func (b ByteView) String() string {
	return string(b.B)
}

func cloneBytes(b []byte) []byte {
	c := make([]byte, len(b))
	copy(c, b)
	return c
}
