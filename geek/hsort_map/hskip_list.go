package hsort_map

import (
	"container/list"
	"math"
	"math/rand"
	"strings"
	"sync"
	"time"
)

const (
	maxLevel    int     = 18 // 跳表默认最大高度
	probability float64 = 1 / math.E
)

// Hash 用于计算key的Hash函数
type Hash func(string) string

// Node 跳表的一个节点
type Node struct {
	next  []*Node
	key   string
	hash  string
	value *list.Element
}

// HSkipList 跳表
// 通过hash排序的跳表，用于存储k-v对，并且可以根据hash值去删除一个区间
// todo: 需要保证线程安全
// 注意：该跳表并不是根据key进行排序的，而是根据key的hash值进行排序的
// 推荐理由：1. 删除区间的时间复杂log(n)
// 不推荐理由：1. 使用读写的复杂度为log(n) 2. 目前没有实现线程安全，用一个较大的锁
type HSkipList struct {
	lock        sync.RWMutex
	head        *Node // 链表头节点
	maxLevel    int   // 链表最大高度
	Len         int   // 跳表元素长度
	randSource  rand.Source
	probability float64   // 没高一层所减少的概率，0.4左右
	probTable   []float64 // 用于存储每一层高的概率，一般是1，0.4， 0.4^2
	hash        Hash      // 用于计算hash值的hash函数，需要外部传递
}

// 创建跳表
// NewHSkipList create a new skip list.
func NewHSkipList(hash Hash) *HSkipList {
	return &HSkipList{
		head: &Node{
			next: make([]*Node, maxLevel),
		},
		maxLevel:    maxLevel,
		randSource:  rand.New(rand.NewSource(time.Now().UnixNano())),
		probability: probability,
		probTable:   probabilityTable(probability, maxLevel),
		hash:        hash,
	}
}

// Next 读取当前节点的下一个节点
func (e *Node) Next() *Node {
	return e.next[0]
}

// Front 获取跳表的第一元素
func (t *HSkipList) Front() *Node {
	return t.head
}

func (t *HSkipList) Get(key string) (*list.Element, bool) {
	t.lock.RLock()
	node, b := t.get(key)
	t.lock.RUnlock()
	if node != nil {
		return node.value, b
	}
	return nil, b
}

// get find value by the key, returns nil if not found.
func (t *HSkipList) get(key string) (*Node, bool) {

	hkey := t.hash(key)

	// node用于存储当前节点
	var node = t.head
	var next *Node

	if node == nil {
		return nil, false
	}

	// 从最高层一直向后遍历
	for i := t.maxLevel - 1; i >= 0; i-- {
		next = node.next[i]

		for next != nil && strings.Compare(hkey, next.hash) > 0 {
			node = next
			next = next.next[i]
		}
	}

	//循环遍历接下来的节点，找到目标key
	node = node.Next()
	for node != nil && strings.Compare(node.hash, hkey) <= 0 {
		if strings.Compare(node.key, key) == 0 {
			return node, true
		}
		node = node.Next()
	}
	return nil, false
}

// Exist 判断key是否存在
func (t *HSkipList) Exist(key string) bool {
	_, answer := t.Get(key)
	return answer
}

// backNodes 查找为目标hash值节点的前面的所有节点
func (t *HSkipList) backNodes(hash string) []*Node {
	node := t.head
	var next *Node

	//记录寻找节点中所经过的节点
	prevs := make([]*Node, t.maxLevel)

	for i := t.maxLevel - 1; i >= 0; i-- {
		next = node.next[i]
		//寻找过程与get保持一致
		for next != nil && strings.Compare(hash, next.hash) > 0 {
			node = next
			next = next.next[i]
		}
		//记录寻找过程
		prevs[i] = node
	}

	return prevs
}

func (t *HSkipList) nextNodes(hash string) []*Node {
	node := t.head
	var next *Node

	//记录寻找节点中所经过的节点
	nexts := make([]*Node, t.maxLevel)

	for i := t.maxLevel - 1; i >= 0; i-- {
		next = node.next[i]
		//寻找过程与get保持一致
		for next != nil && strings.Compare(hash, next.hash) >= 0 {
			node = next
			next = next.next[i]
		}
		//记录寻找过程
		nexts[i] = node
	}

	return nexts
}

// Delete Remove element by the key.
func (t *HSkipList) Delete(key string) bool {
	t.lock.Lock()
	// 判断节点是否存在
	_, f := t.get(key)
	if !f {
		return false
	}
	//寻找key，并记录寻找过程每层经过的最后节点
	hash := t.hash(key)
	prev := t.backNodes(hash)
	// 删除节点
	var answer *list.Element
	for i, node := range prev {
		for node != nil && node.next[i] != nil &&
			strings.Compare(node.next[i].hash, hash) == 0 && strings.Compare(node.next[i].key, key) != 0 {
			node = node.next[i]
		}
		if node != nil && node.next[i] != nil && strings.Compare(node.next[i].key, key) == 0 {
			if answer == nil {
				answer = node.next[i].value
			}
			node.next[i] = node.next[i].next[i]
		}
	}
	t.lock.Unlock()
	if answer != nil {
		t.Len--
	}
	return true
}

// Put an element into skip list, replace the value if key already exists.
func (t *HSkipList) Put(key string, value *list.Element) {
	t.lock.Lock()
	// key已经存在则直接设置为目标值
	node, _ := t.get(key)
	if node != nil {
		node.value = value
	}

	//寻找key相同的hash值的所经过的所有节点
	prev := t.backNodes(t.hash(key))

	//插入节点
	node = &Node{
		// 随机高度
		next:  make([]*Node, t.randomLevel()),
		key:   key,
		hash:  t.hash(key),
		value: value,
	}
	for i := range node.next {
		node.next[i] = prev[i].next[i]
		prev[i].next[i] = node
	}
	t.lock.Unlock()
	t.Len++
}

// DeleteByHashRange 根据一个hash范围进行删除
// [lhash, rhash) 左闭右开
func (t *HSkipList) DeleteByHashRange(lhash string, rhash string) int {
	t.lock.Lock()
	prevs := t.backNodes(lhash)
	prevs2 := t.backNodes(rhash)
	node := prevs[0].next[0]
	tail := prevs2[0].next[0]
	for k := range prevs {
		prevs[k].next[k] = prevs2[k].next[k]
	}
	t.lock.Unlock()
	answer := 0
	for node != tail {
		answer++
		node = node.Next()
	}
	return answer
}

// 一个简单的数学概率问题，假设向上不建立索引的概率为0.4
// （代码实际的概率为1/e）
// 首先第一层（索引为0）是一定会插入的
// 从第二层次（索引为1）开始，向上不建立索引的概率为0.4^1
// 第n层（索引为n-1）则是0.4^（n-1）
// 这个函数就是建立了这样一个表。之后只需要查表插入第几层只需要一次随机数
// 然后拿着这个随机数对表比较即可，随机数比表的数小，就插入
// 比表的数大则不插入
func probabilityTable(probability float64, maxLevel int) (table []float64) {
	for i := 1; i <= maxLevel; i++ {
		prob := math.Pow(probability, float64(i-1))
		table = append(table, prob)
	}
	return table
}

// 随机决定生成多少层索引
// generate random index level.
func (t *HSkipList) randomLevel() (level int) {
	//生成一个随机数
	r := float64(t.randSource.Int63()) / (1 << 63)
	//第一层一定生成
	level = 1
	//如果没到最高层并且随机数比表对应层的概率小，则新建索引
	for level < t.maxLevel && r < t.probTable[level] {
		level++
	}
	return
}
