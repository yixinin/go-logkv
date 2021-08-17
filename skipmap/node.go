package skipmap

import (
	"fmt"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Node struct {
	elem     interface{}
	key      [12]byte
	backward *Node
	level    []skipLevel
}

type skipLevel struct {
	// forward 每层都要有指向下一个节点的指针
	forward *Node
	// span 间隔定义为：从当前节点到 forward 指向的下个节点之间间隔的节点数
	span int
}

func newNode(maxLevel int, key primitive.ObjectID, elem interface{}) *Node {
	return &Node{
		elem:  elem,
		key:   key,
		level: make([]skipLevel, maxLevel),
	}
}

func (node *Node) setLevel(l int) {
	node.level = make([]skipLevel, l)
}

func (node *Node) Compare(other *Node) int {
	return compareSlice(node.key, other.key)
}

func compareSlice(b1, b2 [12]byte) int {
	for i := range b1 {
		if b1[i] < b2[i] {
			return -1
		}
		if b1[i] > b2[i] {
			return 1
		}
	}
	return 0
}

func (node *Node) Lt(other *Node) bool {
	return node.Compare(other) < 0
}

func (node *Node) Lte(other *Node) bool {
	return node.Compare(other) <= 0
}

func (node *Node) Gt(other *Node) bool {
	return node.Compare(other) > 0
}

func (node *Node) Eq(other *Node) bool {
	return node.Compare(other) == 0
}

func (node *Node) String() string {
	return fmt.Sprintf("<Node key=%s, elem='%d'>", node.key, node.elem)
}

func (node *Node) Key() primitive.ObjectID {
	return node.key
}

func (node *Node) Val() interface{} {
	return node.elem
}
