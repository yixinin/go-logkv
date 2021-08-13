package kv

import (
	"sync"
	"time"

	"logkv/skipmap"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type KvIndexer struct {
	sync.RWMutex
	i *skipmap.Skipmap
}

func NewKvIndexer() *KvIndexer {
	return &KvIndexer{
		i: skipmap.New(),
	}
}

func (i *KvIndexer) Get(id primitive.ObjectID) (int64, bool) {
	i.RLock()
	defer i.RUnlock()
	node := i.i.Get(id.Hex())
	if node != nil {
		return node.Val().(int64), true
	}
	return -1, false
}

func (i *KvIndexer) GetMin(key primitive.ObjectID) (offset int64, ok bool) {
	i.RLock()
	defer i.RUnlock()
	node := i.i.FirstInRange(skipmap.Range{
		Min: key.Hex(),
		Max: primitive.NewObjectID().Hex(),
	})

	if node == nil {
		return -1, false
	}
	return node.Val().(int64), true
}

func (i *KvIndexer) GetMax(key primitive.ObjectID) (offset int64, ok bool) {
	i.RLock()
	defer i.RUnlock()

	node := i.i.LastInRange(skipmap.Range{
		Min: primitive.NewObjectIDFromTimestamp(time.Unix(0, 0)).Hex(),
		Max: key.Hex(),
	})
	if node == nil {
		return -1, false
	}
	return node.Val().(int64), true
}

func (i *KvIndexer) Set(id primitive.ObjectID, offset int64) {
	i.Lock()
	defer i.Unlock()
	i.i.Set(id.Hex(), offset)
}

func (i *KvIndexer) Clone() map[string]int64 {
	i.RLock()
	defer i.RUnlock()
	var m = make(map[string]int64)
	var iter = i.i.ToIter()
	for iter.HasNext() {
		node := iter.Next()
		m[node.Key()] = node.Val().(int64)
	}
	return m
}

type Index struct {
	sync.RWMutex
	i map[string][]int64
}

func NewIndex() *Index {
	return &Index{
		i: make(map[string][]int64),
	}
}

func (i *Index) Set(key string, offset int64) {
	i.Lock()
	defer i.Unlock()
	if _, ok := i.i[key]; ok {
		i.i[key] = append(i.i[key], offset)
		return
	}
	i.i[key] = []int64{offset}
}
func (i *Index) Get(key string) []int64 {
	i.RLock()
	defer i.RUnlock()
	return i.i[key]
}

func (i *Index) Clone() map[string][]int64 {
	i.RLock()
	defer i.RUnlock()
	var m = make(map[string][]int64, len(i.i))
	for k, v := range i.i {
		m[k] = v
	}
	return m
}
