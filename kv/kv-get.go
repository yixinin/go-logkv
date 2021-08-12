package kv

import (
	"errors"
	"io"
	bytesutils "logkv/bytes-utils"
	"logkv/protocol"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

var (
	ErrNotFound = errors.New("not found")
)

func (e *KvEngine) Get(id primitive.ObjectID) (protocol.Kv, error) {
	// 先查询cache
	node := e.cache.Get(id.Hex())
	if node != nil {
		data := node.Val().([]byte)
		return protocol.NewKv(id, data), nil
	}
	offset, _ := e.indexer.Get(id)
	return e.get(offset)
}

func (e *KvEngine) get(offset int64) (protocol.Kv, error) {
	if offset == -1 {
		return protocol.Kv{}, ErrNotFound
	}
	_, err := e.fd.Seek(offset, 0)
	if err != nil {
		return protocol.Kv{}, err
	}
	var headerBuf = make([]byte, protocol.HeaderSize+protocol.KeySize)
	_, err = e.fd.Read(headerBuf)
	if err != nil {
		return protocol.Kv{}, err
	}

	dataSize, err := bytesutils.BytesToIntU(headerBuf[:protocol.HeaderSize])
	if err != nil {
		return protocol.Kv{}, err
	}
	var key primitive.ObjectID
	copy(key[:], headerBuf[protocol.HeaderSize:])
	if err != nil {
		return protocol.Kv{}, err
	}
	var data = make([]byte, dataSize)
	_, err = e.fd.Read(data)
	if err != nil {
		return protocol.Kv{}, err
	}
	kv := protocol.NewKv(key, data)
	return kv, nil
}

func (e *KvEngine) BatchGet(indexes []primitive.ObjectID) (protocol.Kvs, error) {
	var kvs = make(protocol.Kvs, 0, len(indexes))
	for _, index := range indexes {
		kv, err := e.Get(index)
		if err != nil {
			return kvs, err
		}
		kvs = append(kvs, kv)
	}
	return kvs, nil
}

func (e *KvEngine) Scan(startIndex, endIndex primitive.ObjectID, limits ...int) (protocol.Kvs, error) {
	var limit = 10 * 1000
	if len(limits) > 0 {
		limit = limits[0]
	}
	offset, _ := e.indexer.GetMin(startIndex)

	return e.scan(offset, limit, endIndex, -1)
}

func (e *KvEngine) scan(offset int64, limit int, endIndex primitive.ObjectID, max int) (protocol.Kvs, error) {
	if offset == -1 {
		return nil, ErrNotFound
	}
	_, err := e.fd.Seek(int64(offset), 0)
	if err != nil {
		return nil, err
	}

	var endKey = endIndex.Hex()
	var readSize = 0
	var kvs = make(protocol.Kvs, 10)
	for i := 0; i < limit; i++ {
		n, kv, err := ReadKv(e.fd)
		if err != nil {
			if err == io.EOF {
				return kvs, nil
			}
			return kvs, err
		}
		if len(kv.Data) == 0 {
			return kvs, nil
		}

		var key = kv.Key.Hex()
		if endKey > "" && key > endKey {
			return kvs, nil
		}
		kvs = append(kvs, kv)
		if endKey > "" && key == endKey {
			return kvs, nil
		}
		readSize += n
		if max > 0 && readSize >= max {
			return kvs, nil
		}
	}
	return kvs, nil
}
