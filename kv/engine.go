package kv

import (
	"bytes"
	"io"
	bytesutils "logkv/bytes-utils"
	"logkv/protocol"
	"os"
	"strconv"
	"sync"
)

type EngineMeta struct {
	filename string
}

type KvEngine struct {
	sync.Mutex
	meta    EngineMeta
	fd      *os.File
	indexer *KvIndexer
}

func NewKvEngine(filename string) *KvEngine {
	e := &KvEngine{
		meta: EngineMeta{
			filename: filename,
		},
		indexer: NewKvIndexer(),
	}
	var err error
	e.fd, err = os.OpenFile(filename, os.O_CREATE|os.O_RDWR|os.O_APPEND, os.ModeAppend|os.ModePerm)
	if err != nil {
		panic(err)
	}
	return e
}

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

func (e *KvEngine) Get(i uint64) (protocol.Kv, error) {
	offsets := e.indexer.Get("index", strconv.FormatUint(i, 10))
	var offset int64
	if len(offsets) > 0 {
		offset = offsets[0]
	}
	_, err := e.fd.Seek(int64(offset), 0)
	if err != nil {
		return protocol.Kv{}, err
	}
	var headerBuf = make([]byte, protocol.HeaderSize+protocol.KeySize)
	_, err = e.fd.Read(headerBuf)
	if err != nil {
		return protocol.Kv{}, err
	}
	key, err := bytesutils.BytesToIntU(headerBuf[protocol.HeaderSize:])
	if err != nil {
		return protocol.Kv{}, err
	}
	dataSize, err := bytesutils.BytesToIntU(headerBuf[:protocol.HeaderSize])
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
	// var limit = math.MaxInt64
	// if len(limits) > 0 {
	// 	limit = limits[0]
	// }

	var kvs = make(protocol.Kvs, 0)
	return kvs, nil
}

func (e *KvEngine) rawFileReader() (io.Reader, error) {
	var reader bytes.Buffer
	f, err := os.OpenFile(e.meta.filename, os.O_RDONLY, os.ModePerm)
	if err != nil {
		return nil, err
	}
	_, err = io.Copy(&reader, f)
	if err != nil {
		return nil, err
	}
	return &reader, nil
}
