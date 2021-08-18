package kv

import (
	"sync"
	"time"

	"logkv/skipmap"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type KvIndexer struct {
	sync.RWMutex
	pk    *skipmap.Skipmap
	trace map[string][]primitive.ObjectID
}

func NewKvIndexer() *KvIndexer {
	return &KvIndexer{
		pk:    skipmap.New(),
		trace: make(map[string][]primitive.ObjectID),
	}
}

func (i *KvIndexer) Get(id primitive.ObjectID) (int64, bool) {
	i.RLock()
	defer i.RUnlock()
	node := i.pk.Get(id)
	if node != nil {
		return node.Val().(int64), true
	}
	return -1, false
}

func (i *KvIndexer) GetMin(key primitive.ObjectID) (offset int64, ok bool) {
	i.RLock()
	defer i.RUnlock()
	node := i.pk.FirstInRange(skipmap.Range{
		Min: key,
		Max: primitive.NewObjectID(),
	})

	if node == nil {
		return -1, false
	}
	return node.Val().(int64), true
}

func (i *KvIndexer) GetMax(key primitive.ObjectID) (offset int64, ok bool) {
	i.RLock()
	defer i.RUnlock()

	node := i.pk.LastInRange(skipmap.Range{
		Min: primitive.NewObjectIDFromTimestamp(time.Unix(0, 0)),
		Max: key,
	})
	if node == nil {
		return -1, false
	}
	return node.Val().(int64), true
}

func (i *KvIndexer) Set(id primitive.ObjectID, offset int64) {
	i.Lock()
	defer i.Unlock()
	i.pk.Set(id, offset)
}

func (i *KvIndexer) SetTrace(trace string, ids ...primitive.ObjectID) {
	i.Lock()
	defer i.Unlock()
	i.trace[trace] = append(i.trace[trace], ids...)
}

func (i *KvIndexer) GetTrace(trace string) []primitive.ObjectID {
	i.RLock()
	defer i.RUnlock()
	return i.trace[trace]
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
