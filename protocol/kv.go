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
	key, err = bytesutils.BytesToIntU(buf[HeaderSize : HeaderSize+KeySize])
	if err != nil {
		return Kv{}, err
	}
	data := buf[HeaderSize+KeySize : HeaderSize+KeySize+dataSize]
	if len(data) != int(dataSize) {
		return Kv{}, errors.New("data not match")
	}
	var kv = NewKv(key, data)
	indexBuf := buf[HeaderSize+KeySize+dataSize:]
	err = json.Unmarshal(indexBuf, &kv.Indexes)
	return kv, err
}

func KvFromBytes(buf []byte) (Kv, error) {
	dataSize, err := bytesutils.BytesToIntU(buf[:HeaderSize])
	if err != nil {
		return Kv{}, err
	}
	key, err := bytesutils.BytesToIntU(buf[HeaderSize : HeaderSize+KeySize])
	if err != nil {
		return Kv{}, err
	}
	data := buf[HeaderSize+KeySize:]
	if len(data) != int(dataSize) {
		return Kv{}, errors.New("data not match")
	}
	return NewKv(key, data), nil
}

func (kv *Kv) Bytes() []byte {
	var buf = make([]byte, HeaderSize+KeySize+len(kv.Data))
	dataSizeBuf := bytesutils.IntToBytes(len(kv.Data), HeaderSize)
	copy(buf[:HeaderSize], dataSizeBuf)
	copy(buf[HeaderSize:HeaderSize+KeySize], bytesutils.IntToBytes(int(kv.Key), KeySize))
	copy(buf[HeaderSize+KeySize:], kv.Data)
	return buf
}

func (kv *Kv) BytesWithIndexes() []byte {
	var buf = kv.Bytes()
	var indexBuf, _ = json.Marshal(kv.Indexes)
	buf = append(buf, indexBuf...)
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
