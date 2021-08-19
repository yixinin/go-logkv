package main

type Node struct {
	Val  interface{}
	Next *Node
	Prev *Node
}
type LRUCache struct {
	m    map[string]*Node
	head *Node
	cap  int
}

func NewLRUCache(cap int) *LRUCache {
	return &LRUCache{
		cap:  cap,
		head: &Node{},
		m:    make(map[string]*Node, cap),
	}
}

func (c *LRUCache) Set(key string, val interface{}) {
	old := c.head.Next
	c.head.Next = &Node{
		Prev: c.head,
		Next: old,
		Val:  val,
	}
}

func (c *LRUCache) Get(key string) (interface{}, bool) {
	node, ok := c.m[key]
	old := c.head.Next
	if node != old {
		node.Prev.Next = node.Next
		c.head.Next = node
		node.Next = old
	}
	return node, ok
}
