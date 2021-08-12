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

func (e *KvEngine) Get(id primitive.ObjectID) ([]byte, error) {
	// 先查询cache
	node := e.cache.Get(id.Hex())
	if node != nil {
		data := node.Val().([]byte)
		return data, nil
	}
	offset, _ := e.indexer.Get(id)
	return e.get(offset)
}

func (e *KvEngine) get(offset int64) ([]byte, error) {
	if offset == -1 {
		return nil, ErrNotFound
	}
	_, err := e.fd.Seek(offset, 0)
	if err != nil {
		return nil, err
	}
	var headerBuf = make([]byte, protocol.HeaderSize)
	_, err = e.fd.Read(headerBuf)
	if err != nil {
		return nil, err
	}

	dataSize, err := bytesutils.BytesToIntU(headerBuf)
	if err != nil {
		return nil, err
	}

	var data = make([]byte, dataSize)
	_, err = e.fd.Read(data)
	if err != nil {
		return nil, err
	}
	return data, nil
}

func (e *KvEngine) BatchGet(indexes []primitive.ObjectID) ([][]byte, error) {
	var kvs = make([][]byte, 0, len(indexes))
	for _, index := range indexes {
		kv, err := e.Get(index)
		if err != nil {
			return kvs, err
		}
		kvs = append(kvs, kv)
	}
	return kvs, nil
}

func (e *KvEngine) Scan(startIndex, endIndex primitive.ObjectID, limits ...int) ([][]byte, error) {
	var limit = 10 * 1000
	if len(limits) > 0 {
		limit = limits[0]
	}
	offset, _ := e.indexer.GetMin(startIndex)

	return e.scan(offset, limit, endIndex, -1)
}

func (e *KvEngine) scan(offset int64, limit int, endIndex primitive.ObjectID, max int) ([][]byte, error) {
	if offset == -1 {
		return nil, ErrNotFound
	}
	_, err := e.fd.Seek(int64(offset), 0)
	if err != nil {
		return nil, err
	}

	var endKey = endIndex.Hex()
	var readSize = 0
	var kvs = make([][]byte, 0, limit)
	for i := 0; i < limit; i++ {
		n, key, data, err := ReadIndex(e.fd)
		if err != nil {
			if err == io.EOF {
				return kvs, nil
			}
			return kvs, err
		}
		if len(data) == 0 {
			return kvs, nil
		}

		if endKey > "" && key.Hex() > endKey {
			return kvs, nil
		}
		kvs = append(kvs, data)
		if endKey > "" && key.Hex() == endKey {
			return kvs, nil
		}
		readSize += n
		if max > 0 && readSize >= max {
			return kvs, nil
		}
	}
	return kvs, nil
}
