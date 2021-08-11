package protocol

import (
	"errors"
	bytesutils "logkv/bytes-utils"

	"gopkg.in/mgo.v2/bson"
)

const (
	KeySize    = 12
	HeaderSize = 8
)

type Kv struct {
	Key  bson.ObjectId
	Data []byte
}

func NewKv(key bson.ObjectId, data []byte) Kv {
	return Kv{
		Key:  key,
		Data: data,
	}
}

func KvFromBytes(buf []byte) (Kv, error) {
	dataSize, err := bytesutils.BytesToIntU(buf[:HeaderSize])
	if err != nil {
		return Kv{}, err
	}

	var key = bson.ObjectIdHex(string(buf[HeaderSize : HeaderSize+KeySize]))
	if err != nil {
		return Kv{}, err
	}
	data := buf[HeaderSize+KeySize:]
	if len(data) != int(dataSize) {
		return Kv{}, errors.New("data not match")
	}

	var kv = NewKv(key, data)
	return kv, err
}

func (kv *Kv) Bytes() []byte {
	if len(kv.Data) == 0 {
		return nil
	}
	// [dataSize+key+data]
	var buf = make([]byte, HeaderSize+KeySize+len(kv.Data))
	dataSizeBuf := bytesutils.UintToBytes(uint64(len(kv.Data)), HeaderSize)
	copy(buf[:HeaderSize], dataSizeBuf)
	copy(buf[HeaderSize:HeaderSize+KeySize], kv.Key[:])
	copy(buf[HeaderSize+KeySize:], kv.Data)
	return buf
}

type Kvs []Kv

func (kvs Kvs) Bytes() []byte {
	var buf = make([]byte, 0, 100*len(kvs))
	for _, kv := range kvs {
		b := kv.Bytes()
		if len(b) > 0 {
			buf = append(buf, b...)
		}
	}
	return buf
}
