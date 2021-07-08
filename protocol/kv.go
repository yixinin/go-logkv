package protocol

import (
	"encoding/json"
	"errors"
	bytesutils "logkv/bytes-utils"
)

const (
	KeySize    = 8
	HeaderSize = 8
)

type Kv struct {
	Key     uint64
	Data    []byte
	Indexes map[string]string
}

func NewKv(key uint64, data []byte) Kv {
	var kv = Kv{
		Key:     key,
		Data:    data,
		Indexes: make(map[string]string),
	}
	return kv
}

func NewKvWithIndexes(key uint64, buf []byte) (Kv, error) {
	dataSize, err := bytesutils.BytesToIntU(buf[:HeaderSize])
	if err != nil {
		return Kv{}, err
	}

	data := buf[HeaderSize : HeaderSize+dataSize]
	if len(data) != int(dataSize) {
		return Kv{}, errors.New("data not match")
	}
	var kv = NewKv(key, data)
	indexBuf := buf[HeaderSize+dataSize:]
	err = json.Unmarshal(indexBuf, &kv.Indexes)
	return kv, err
}

func KvFromBytes(buf []byte) (Kv, error) {
	dataSize, err := bytesutils.BytesToIntU(buf[:HeaderSize])
	if err != nil {
		return Kv{}, err
	}
	indexSize, err := bytesutils.BytesToIntU(buf[HeaderSize : HeaderSize*2])
	if err != nil {
		return Kv{}, err
	}
	key, err := bytesutils.BytesToIntU(buf[HeaderSize*2 : HeaderSize*2+KeySize])
	if err != nil {
		return Kv{}, err
	}
	data := buf[HeaderSize*2+KeySize : HeaderSize*2+KeySize+dataSize]
	if len(data) != int(dataSize) {
		return Kv{}, errors.New("data not match")
	}
	var kv = NewKv(key, data)
	err = json.Unmarshal(buf[HeaderSize*2+KeySize+dataSize:HeaderSize*2+KeySize+dataSize+indexSize], &kv.Indexes)
	return kv, err
}

func (kv *Kv) Bytes() []byte {
	// [dataSize+indexSize+key+data+index]
	indexBuf, _ := json.Marshal(kv.Indexes)
	var buf = make([]byte, HeaderSize*2+KeySize+len(kv.Data)+len(indexBuf))
	dataSizeBuf := bytesutils.UintToBytes(uint64(len(kv.Data)), HeaderSize)
	indexSizeBuf := bytesutils.UintToBytes(uint64(len(indexBuf)), 8)
	copy(buf[:HeaderSize], dataSizeBuf)
	copy(buf[HeaderSize:HeaderSize*2], indexSizeBuf)

	copy(buf[HeaderSize*2:HeaderSize*2+KeySize], bytesutils.UintToBytes(kv.Key, KeySize))
	copy(buf[HeaderSize*2+KeySize:HeaderSize*2+KeySize+len(kv.Data)], kv.Data)
	copy(buf[HeaderSize*2+KeySize+len(kv.Data):], indexBuf)
	return buf
}

func (kv *Kv) BytesWithIndexes() []byte {
	var indexBuf, _ = json.Marshal(kv.Indexes)
	var buf = make([]byte, HeaderSize+len(kv.Data)+len(indexBuf))
	dataSizeBuf := bytesutils.UintToBytes(uint64(len(kv.Data)), 8)
	copy(buf[:HeaderSize], dataSizeBuf)
	copy(buf[HeaderSize:HeaderSize+len(dataSizeBuf)], kv.Data)
	copy(buf[HeaderSize+len(kv.Data):], indexBuf)
	return buf
}

type Kvs []Kv

func (kvs Kvs) Bytes() []byte {
	var buf = make([]byte, 0, 1024*len(kvs))
	for _, kv := range kvs {
		buf = append(buf, kv.Bytes()...)
	}
	return buf
}
