package kv

import (
	"logkv/protocol"
	"strconv"
)

func (e *KvEngine) Set(kv protocol.Kv) error {
	e.Lock()
	defer e.Unlock()
	var offset int64
	stat, err := e.fd.Stat()
	if err != nil {
		return err
	}
	offset = stat.Size()
	_, err = e.fd.Write(kv.Bytes())
	if err != nil {
		return err
	}
	var indexes = make(map[string]string, len(kv.Indexes))
	indexes["index"] = strconv.FormatUint(kv.Key, 10)
	for k, v := range kv.Indexes {
		indexes[k] = v
	}
	for k, v := range indexes {
		e.indexer.Set(k, v, offset)
	}
	return nil
}

func (e *KvEngine) BatchSet(kvs protocol.Kvs) error {
	for _, v := range kvs {
		if err := e.Set(v); err != nil {
			return err
		}
	}
	return nil
}
