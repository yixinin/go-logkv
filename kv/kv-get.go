package kv

import (
	"errors"
	"io"
	bytesutils "logkv/bytes-utils"
	"logkv/protocol"
	"strconv"
)

var (
	ErrNotFound = errors.New("not found")
)

func (e *KvEngine) Get(i uint64) (protocol.Kv, error) {
	offsets := e.indexer.Get("index", strconv.FormatUint(i, 10))
	var offset int64 = -1
	if len(offsets) > 0 {
		offset = offsets[0]
	}
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
	var headerBuf = make([]byte, protocol.HeaderSize*2+protocol.KeySize)
	_, err = e.fd.Read(headerBuf)
	if err != nil {
		return protocol.Kv{}, err
	}

	dataSize, err := bytesutils.BytesToIntU(headerBuf[:protocol.HeaderSize])
	if err != nil {
		return protocol.Kv{}, err
	}
	// indexSize, err := bytesutils.BytesToIntU(headerBuf[protocol.HeaderSize : protocol.HeaderSize*2])
	// if err != nil {
	// 	return protocol.Kv{}, err
	// }
	key, err := bytesutils.BytesToIntU(headerBuf[protocol.HeaderSize*2:])
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

func (e *KvEngine) BatchGet(indexes []uint64) (protocol.Kvs, error) {
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

func (e *KvEngine) Scan(startIndex, endIndex uint64, limits ...int) (protocol.Kvs, error) {
	var limit = 10 * 1000
	if len(limits) > 0 {
		limit = limits[0]
	}
	offsets := e.indexer.Get("index", strconv.FormatUint(startIndex, 10))
	var offset int64 = -1
	if len(offsets) > 0 {
		offset = offsets[0]
	}
	return e.scan(offset, limit, endIndex, -1)
}

func (e *KvEngine) scan(offset int64, limit int, endIndex uint64, max int) (protocol.Kvs, error) {
	if offset == -1 {
		return nil, ErrNotFound
	}
	_, err := e.fd.Seek(int64(offset), 0)
	if err != nil {
		return nil, err
	}

	var readSize = 0
	var kvs = make(protocol.Kvs, 10)
	for i := 0; i < limit; i++ {
		n, kv, err := ReadSnapshot(e.fd)
		if err != nil {
			if err == io.EOF {
				return kvs, nil
			}
			return kvs, err
		}
		if len(kv.Data) == 0 {
			return kvs, nil
		}

		if endIndex > 0 && kv.Key > endIndex {
			return kvs, nil
		}
		kvs = append(kvs, kv)
		if endIndex > 0 && kv.Key == endIndex {
			return kvs, nil
		}
		readSize += n
		if max > 0 && readSize >= max {
			return kvs, nil
		}
	}
	return kvs, nil
}

func (e *KvEngine) GetWithIndex(name string, val string) (protocol.Kvs, error) {
	offsets := e.indexer.Get(name, val)
	if len(offsets) == 0 {
		return protocol.Kvs{}, ErrNotFound
	}
	var kvs = make(protocol.Kvs, 0, len(offsets))
	for _, v := range offsets {
		kv, err := e.get(v)
		if err != nil {
			return kvs, err
		}
		kvs = append(kvs, kv)
	}
	return kvs, nil
}

func (e *KvEngine) ScanIndex(name string, startVal, endVal string, limits ...int) (protocol.Kvs, error) {
	var limit = 10 * 1000
	if len(limits) > 0 {
		limit = limits[0]
	}
	offsets := e.indexer.Get(name, startVal)
	var offset int64 = -1
	for _, v := range offsets {
		offset = v
		break
	}
	endOffsets := e.indexer.Get(name, endVal)
	var endOffset int64 = -1
	for _, v := range endOffsets {
		endOffset = v
	}
	return e.scan(offset, limit, 0, int(endOffset))
}
