package kv

import (
	"sync"
	"time"

	"logkv/skipmap"

	"gopkg.in/mgo.v2/bson"
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

func (i *KvIndexer) Get(id bson.ObjectId) (int64, bool) {
	i.RLock()
	defer i.RUnlock()
	node := i.i.FirstInRange(skipmap.Range{
		Min: id.Hex(),
		Max: id.Hex(),
	})
	if node == nil {
		return -1, false
	}
	return node.Val(), true
}

func (i *KvIndexer) GetMin(key bson.ObjectId) (offset int64, ok bool) {
	i.RLock()
	defer i.RUnlock()
	node := i.i.FirstInRange(skipmap.Range{
		Min: key.Hex(),
		Max: bson.NewObjectId().Hex(),
	})

	if node == nil {
		return -1, false
	}
	return node.Val(), true
}

func (i *KvIndexer) GetMax(key bson.ObjectId) (offset int64, ok bool) {
	i.RLock()
	defer i.RUnlock()

	node := i.i.LastInRange(skipmap.Range{
		Min:        bson.NewObjectIdWithTime(time.Unix(0, 0)).Hex(),
		Max:        key.Hex(),
		ExcludeMin: true,
	})
	if node == nil {
		return -1, false
	}
	return node.Val(), true
}

func (i *KvIndexer) Set(id bson.ObjectId, offset int64) {
	i.Lock()
	defer i.Unlock()
	i.i.Insert(id.Hex(), offset)
}

func (i *KvIndexer) Clone() map[string]int64 {
	i.RLock()
	defer i.RUnlock()
	var m = make(map[string]int64)
	var iter = i.i.ToIter()
	for iter.HasNext() {
		node := iter.Next()
		m[node.Key()] = node.Val()
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
