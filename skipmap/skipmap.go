// Package skipmap 基于 Redis 的 zskipmap 实现的一个
// 跳表数据结构。
// Redis zset 底层实现之一就使用到了跳表，当然还包括一个 dict
// 大体定义如下：
// ```go
// type zset struct {
//      dict map[Element]Key
//      skipmap *Skipmap
// }
// 其中，dict 记录了 key -> key 的映射，所以获取一个 key
// 的 key 时间复杂度为 O(1)；而在 skipmap 中可以快速查找
// 到 key 对应的 key(s)，对应的时间复杂度为 O(logN)。
// ```
package skipmap

import (
	"fmt"
	"math/rand"
)

const (
	MaxLevel = 64 // 足以容纳 2^64 个元素
	P        = 0.25
)

type Skipmap struct {
	header, tail *Node
	level        int // 记录跳表的实际高度
	length       int // 记录跳表的长度（不含头节点）
}

func New() *Skipmap {
	return &Skipmap{
		level: 1,
		// 头节点比较特殊，它有 64 层。头节点不会存储具体的元素信息
		header: newNode(MaxLevel, "", -1),
		tail:   nil,
	}
}

func (m *Skipmap) String() string {
	return fmt.Sprintf("<Skipmap level=%d, length=%d, tail=%s>", m.level, m.length, m.tail)
}

func (m *Skipmap) Len() int {
	return m.length
}

func (m *Skipmap) First() (*Node, bool) {
	f := m.header.level[0].forward
	return f, f != nil
}

func (m *Skipmap) Last() (*Node, bool) {
	f := m.tail
	return f, f != nil
}

// Set 向跳表中插入一个新的元素。
// 步骤：
// 1. 查找插入位置
// 2. 创建新节点，并在目标位置插入节点
// 3. 调整跳表 backward 指针等
func (m *Skipmap) Set(key string, elem interface{}) *Node {
	var (
		// update 用于记录每层待更新的节点
		update [MaxLevel]*Node
		// rank 用来记录每层经过的节点记录（可以看成到头节点的距离）
		rank [MaxLevel]int
		// 构建一个新节点，用于下面的大小判断，其 level 在后面设置
		node = &Node{key: key, elem: elem}
	)

	cur := m.header
	for i := m.level - 1; i >= 0; i-- {
		if cur == m.header {
			rank[i] = 0
		} else {
			rank[i] = rank[i+1]
		}

		// 与同层的后一个节点比较，如果后一个比目标值小，则可以继续向后
		// 否则下降到一层查找。注意这里的大小比较是按照 key 和
		// elem 综合计算得到的。
		for cur.level[i].forward != nil && cur.level[i].forward.Lt(node) {
			rank[i] += cur.level[i].span
			// 同层继续往后查找
			cur = cur.level[i].forward
		}
		update[i] = cur
	}

	// 调整跳表高度
	level := m.randomLevel()
	if level > m.level {
		// 初始化每层
		for i := level - 1; i >= m.level; i-- {
			rank[i] = 0
			update[i] = m.header
			update[i].level[i].span = m.length
		}

		m.level = level
	}

	// 更新节点 level，并插入新节点
	node.setLevel(level)
	for i := 0; i < level; i++ {
		// 更新每层的节点指向
		node.level[i].forward = update[i].level[i].forward
		update[i].level[i].forward = node

		// 更新 span 信息
		node.level[i].span = update[i].level[i].span - (rank[0] - rank[i])
		update[i].level[i].span = (rank[0] - rank[i]) + 1
	}

	// 针对新增节点 level < m.level 的情况，需要更新上面没有扫到的层 span
	for i := level; i < m.level; i++ {
		update[i].level[i].span++
	}

	// 调整 backward 指针
	// 如果前一个节点是头节点，则 backward 为 nil
	// 否则 backward 指向之前节点
	if update[0] != m.header {
		// update[0] 就是和新增节点相邻的前一个节点
		node.backward = update[0]
	}

	// 如果新增节点是最后一个，则需要更新 tail 指针
	if node.level[0].forward == nil {
		m.tail = node
	} else {
		// 中间节点，需要更新后一个节点的回退指针
		node.level[0].forward.backward = node
	}

	m.length++
	return node
}

// randomLevel 对于新增节点，返回一个随机的 level
// 返回的 level 范围为 [1, MaxLevel]。并且，采用的
// 算法会保证，更大的 level 返回的概率越低。
// 每个 level 出现的概率计算：(1-p) * p^(level-1)
func (m *Skipmap) randomLevel() int {
	level := 1
	for rand.Float64() < P && level < MaxLevel {
		level++
	}
	return level
}

// Delete 用于删除跳表中指定的节点。
func (m *Skipmap) Delete(key string) *Node {
	// 第一步，找到需要删除节点
	var (
		update     [MaxLevel]*Node
		targetNode = &Node{key: key}
	)

	cur := m.header
	for i := m.level - 1; i >= 0; i-- {
		for cur.level[i].forward != nil && cur.level[i].forward.Lt(targetNode) {
			cur = cur.level[i].forward
		}
		update[i] = cur
	}

	// 目标节点找到后，这里需要判断下 elem 是否相等
	// key 可以重复，所以必须要谨慎
	nodeToBeDeleted := update[0].level[0].forward
	if nodeToBeDeleted == nil || !nodeToBeDeleted.Eq(targetNode) {
		return nil
	}

	m.deleteNode(update, nodeToBeDeleted)
	return nodeToBeDeleted
}

func (m *Skipmap) Get(key string) *Node {
	var (
		update     [MaxLevel]*Node
		targetNode = &Node{key: key}
	)

	cur := m.header
	for i := m.level - 1; i >= 0; i-- {
		for cur.level[i].forward != nil && cur.level[i].forward.Lt(targetNode) {
			cur = cur.level[i].forward
		}
		update[i] = cur
	}

	target := update[0].level[0].forward
	return target
}

func (m *Skipmap) deleteNode(update [64]*Node, nodeToBeDeleted *Node) {
	// 这时我们要删除的节点就是 nodeToBeDeleted
	// 调整每层待更新节点，修改 forward 指向
	for i := 0; i < m.level; i++ {
		if update[i].level[i].forward == nodeToBeDeleted {
			update[i].level[i].forward = nodeToBeDeleted.level[i].forward
			update[i].level[i].span += nodeToBeDeleted.level[i].span - 1
		} else {
			update[i].level[i].span--
		}
	}
	// 调整回退指针：
	// 1. 如果被删除的节点是最后一个节点，需要更新 m.tail
	// 2. 如果被删除的节点位于中间，则直接更新后一个节点 backward 即可
	if m.tail == nodeToBeDeleted {
		m.tail = nodeToBeDeleted.backward
	} else {
		nodeToBeDeleted.level[0].forward.backward = nodeToBeDeleted.backward
	}
	// 调整层数
	for m.header.level[m.level-1].forward == nil {
		m.level--
	}
	// 减少节点计数
	m.length--
	nodeToBeDeleted.backward = nil
	nodeToBeDeleted.level[0].forward = nil
}

// UpdateKey 用于更新节点的分数。该函数会保证更新分数后，
// 节点的有序性依然可以维持。
// 策略如下：
// 1. 快速判断能否原节点修改，如果可以则直接修改并返回；
// 2. 采用更加昂贵的操作：删除再添加。
func (m *Skipmap) UpdateKey(curKey string, elem int64, newKey string) *Node {
	var (
		update     [MaxLevel]*Node
		targetNode = &Node{elem: elem, key: curKey}
	)

	cur := m.header
	// 第一步，找到符合条件的目标节点
	for i := m.level - 1; i >= 0; i-- {
		for cur.level[i].forward != nil && cur.level[i].forward.Lt(targetNode) {
			cur = cur.level[i].forward
		}
		update[i] = cur
	}

	node := cur.level[0].forward
	if node == nil || !node.Eq(targetNode) {
		return nil
	}

	if m.canUpdateKeyFor(node, newKey) {
		node.key = newKey
		return node
	} else {
		// 需要删除旧节点，增加新节点
		m.deleteNode(update, node)
		return m.Set(newKey, node.elem)
	}
}

// canUpdateKeyFor 确定能否直接在原有的节点上进行修改
// 什么条件才可以直接原地更新 key 呢？
// 1. node 是唯一一个数据节点（node.backward == NULL && node->level[0].forward == NULL）
// 2. node 是第一个数据节点，且新的分数要比 node 之后节点分数要小（这样才能保证有序）
//    即：node.backward == NULL && node->level[0].forward->key > newKey）
// 3. node 是最后一个数据节点，且 node 之前节点的分数要比新改的分数小
//    即：node->backward->key < newKey && node->level[0].forward == NULL
// 4. node 是修改的后的分数恰好还能保证位于前一个和后一个节点分数之间
//    即：node->backward->key < newkey && node->level[0].forward->key > newkey
func (m *Skipmap) canUpdateKeyFor(node *Node, newKey string) bool {
	if (node.backward == nil || node.backward.key < newKey) &&
		(node.level[0].forward == nil || node.level[0].forward.key > newKey) {
		return true
	}
	return false
}

type Range struct {
	Min, Max               string
	ExcludeMin, ExcludeMax bool
}

func (r *Range) GteMin(v string) bool {
	if r.ExcludeMin {
		return v > r.Min
	}
	return v >= r.Min
}

func (r *Range) LteMax(v string) bool {
	if r.ExcludeMax {
		return v < r.Max
	}
	return v <= r.Max
}

func (r *Range) isValid() bool {
	if r.Min > r.Max {
		return false
	}

	if r.Min == r.Max && r.ExcludeMin && r.ExcludeMax {
		return false
	}

	return true
}

// FirstInRange 查找符合指定范围内的第一个节点。
func (m *Skipmap) FirstInRange(rng Range) *Node {
	if !m.isInRange(rng) {
		return nil
	}

	cur := m.header
	for i := m.level - 1; i >= 0; i-- {
		// 注意边界情况
		for cur.level[i].forward != nil && !rng.GteMin(cur.level[i].forward.key) {
			cur = cur.level[i].forward
		}
	}

	cur = cur.level[0].forward
	// 检查最终返回结果是否超过最大值了
	if cur == nil || !rng.LteMax(cur.key) {
		return nil
	}

	return cur
}

// LastInRange 查找符合指定范围内的最后一个节点。
func (m *Skipmap) LastInRange(rng Range) *Node {
	if !m.isInRange(rng) {
		return nil
	}

	cur := m.header
	for i := m.level - 1; i >= 0; i-- {
		// 注意边界情况
		for cur.level[0].forward != nil && rng.LteMax(cur.level[0].forward.key) {
			cur = cur.level[0].forward
		}
	}

	if cur == nil || !rng.GteMin(cur.key) {
		return nil
	}

	return cur
}

// isInRange 用来快速判断查找的范围是否有效，避免无效查找。
// 以下是 min 和 max 可能存在的位置
//                [1, 2, 3, 4, 5, 6, 7, 8, 9]
// [min,max]  [min,max]    [min,max]     [min,max] [min,max]
//            [min,                          max]
func (m *Skipmap) isInRange(rng Range) bool {
	if !rng.isValid() || m.length == 0 {
		return false
	}

	if !rng.GteMin(m.tail.key) {
		return false
	}

	if !rng.LteMax(m.header.level[0].forward.key) {
		return false
	}

	return true
}

// Rank 返回一个节点的排名。
// 注意，排名是从 1 开始计算的；如果 <key, elem> 未找到，则返回 0
func (m *Skipmap) Rank(key string, elem int64) int {
	rank := 0

	target := &Node{key: key, elem: elem}
	cur := m.header
	for i := m.level - 1; i >= 0; i-- {
		for cur.level[i].forward != nil && cur.level[i].forward.Lte(target) {
			rank += cur.level[i].span
			cur = cur.level[i].forward
		}

		if cur.Eq(target) {
			return rank
		}
	}

	return 0
}

// 返回指定排名的节点。
func (m *Skipmap) ElementByRank(rank int) *Node {
	if rank <= 0 {
		return nil
	}

	var (
		traversed = 0
		cur       = m.header
	)

	for i := m.level - 1; i >= 0; i-- {
		for cur.level[i].forward != nil && traversed+cur.level[i].span <= rank {
			traversed += cur.level[i].span
			cur = cur.level[i].forward
		}

		if traversed == rank {
			return cur
		}
	}

	return nil
}

// ToIter 生成一个迭代器，用于遍历跳表中的所有节点。
func (m *Skipmap) ToIter() *Iterator {
	return newIterator(m, false)
}

// ToReverseIter 生成一个迭代器，用于倒序遍历跳表中的所有节点。
func (m *Skipmap) ToReverseIter() *Iterator {
	return newIterator(m, true)
}

type Iterator struct {
	m       *Skipmap
	current *Node
	reverse bool
}

func newIterator(m *Skipmap, reverse bool) *Iterator {
	it := &Iterator{
		m:       m,
		reverse: reverse,
	}
	if reverse {
		it.current = m.tail
	} else {
		if m.header != nil {
			it.current = m.header.level[0].forward
		}
	}
	return it
}

func (it *Iterator) HasNext() bool {
	if it.current == nil {
		return false
	}
	return true
}

func (it *Iterator) Next() *Node {
	if it.current == nil {
		return nil
	}

	var next *Node
	if it.reverse {
		next = it.current.backward
	} else {
		next = it.current.level[0].forward
	}

	cur := it.current
	it.current = next
	return cur
}
