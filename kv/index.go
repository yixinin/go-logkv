package kv

import (
	"logkv/index"
	"sync"
)

type KvIndexer struct {
	sync.RWMutex
	i map[string]*Index
}

func NewKvIndexer() *KvIndexer {
	return &KvIndexer{
		i: make(map[string]*Index, 1),
	}
}
func (i *KvIndexer) Get(name string, key index.IndexVal) []int64 {
	i.RLock()
	index, ok := i.i[name]
	i.RUnlock()
	if ok {
		return index.Get(key)
	}
	return nil
}
func (i *KvIndexer) Set(name string, key index.IndexVal, offset int64) {
	i.Lock()
	index, ok := i.i[name]
	if !ok {
		index = NewIndex()
		i.i[name] = index
	}
	i.Unlock()
	index.Set(key, offset)
}

type Index struct {
	sync.RWMutex
	i map[index.IndexVal][]int64
}

func NewIndex() *Index {
	return &Index{
		i: make(map[index.IndexVal][]int64),
	}
}

func (i *Index) Set(key index.IndexVal, offset int64) {
	i.Lock()
	defer i.Unlock()
	if _, ok := i.i[key]; ok {
		i.i[key] = append(i.i[key], offset)
	}
	i.i[key] = []int64{offset}
}
func (i *Index) Get(key index.IndexVal) []int64 {
	i.RLock()
	defer i.RUnlock()
	return i.i[key]
}
