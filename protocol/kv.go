package protocol

import (
	"errors"
	bytesutils "logkv/bytes-utils"
	"logkv/kvid"
)

const (
	KeySize    = 12
	HeaderSize = 8
)

type Kv struct {
	Key  kvid.Id
	Data []byte
}

func NewKv(ts uint32, index uint64, data []byte) Kv {
	var key = [12]byte{}
	copy(key[:4], bytesutils.UintToBytes(uint64(ts), 4))
	copy(key[4:], bytesutils.UintToBytes(index, 8))
	var kv = Kv{
		Key:  key,
		Data: data,
	}
	return kv
}

func KvFromKv(key kvid.Id, data []byte) Kv {
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
	var key = kvid.FromBytes(buf[HeaderSize : HeaderSize+KeySize])
	if err != nil {
		return Kv{}, err
	}
	data := buf[HeaderSize+KeySize:]
	if len(data) != int(dataSize) {
		return Kv{}, errors.New("data not match")
	}

	var kv = KvFromKv(key, data)
	return kv, err
}

func (kv *Kv) Bytes() []byte {
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
		buf = append(buf, kv.Bytes()...)
	}
	return buf
}
