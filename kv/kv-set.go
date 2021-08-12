package kv

import (
	"logkv/protocol"
)

func (e *KvEngine) Set(kv protocol.Kv) {
	e.ch <- kv
}

func (e *KvEngine) BatchSet(kvs protocol.Kvs) {
	for _, v := range kvs {
		e.ch <- v
	}
}

func (e *KvEngine) receive() {
	for kv := range e.ch {
		e.cache.Set(kv.Key.Hex(), kv.Data)
	}
}
