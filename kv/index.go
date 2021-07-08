package kv

import (
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
func (i *KvIndexer) Get(name string, key string) []int64 {
	i.RLock()
	index, ok := i.i[name]
	i.RUnlock()
	if ok {
		return index.Get(key)
	}
	return nil
}
func (i *KvIndexer) Set(name string, key string, offset int64) {
	i.Lock()
	index, ok := i.i[name]
	if !ok {
		index = NewIndex()
		i.i[name] = index
	}
	i.Unlock()
	index.Set(key, offset)
}

func (i *KvIndexer) Clone() map[string]map[string][]int64 {
	i.RLock()
	defer i.RUnlock()
	var m = make(map[string]map[string][]int64, len(i.i))
	for k, v := range i.i {
		m[k] = v.Clone()
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
