package protocol

import (
	"errors"
	bytesutils "logkv/bytes-utils"
	"logkv/index"
)

const (
	KeySize    = 8
	HeaderSize = 8
)

type Kv struct {
	Key     uint64
	Data    []byte
	Indexes map[string]interface{}
}

func NewKv(key uint64, data []byte) Kv {
	var kv = Kv{
		Key:  key,
		Data: data,
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
	for len(indexBuf) > 3 {
		keySize := indexBuf[0]
		valSize := indexBuf[1]
		valType := indexBuf[3]
		indexKey := indexBuf[3:keySize]
		indexVal := indexBuf[3+keySize : 3+keySize+valSize]
		var val interface{}
		switch valType {
		case 1:
			val = string(indexVal)
		case 2:
			val, err = bytesutils.BytesToIntU(indexVal)
			if err != nil {
				return kv, err
			}
		case 3:
			val, err = bytesutils.BytesToFloat(indexVal)
			if err != nil {
				return kv, err
			}
		}
		kv.Indexes[string(indexKey)] = val
		indexBuf = indexBuf[3+keySize+valSize:]
	}
	return kv, nil
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
	for k, v := range kv.Indexes {
		indexVal := index.NewIndexVal(v)
		buf = append(buf, byte(len(k)))
		buf = append(buf, byte(indexVal.Size()))
		buf = append(buf, indexVal.Type())
		buf = append(buf, []byte(k)...)
		buf = append(buf, indexVal.Bytes()...)
	}
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
