package protocol

import (
	"bytes"
	"encoding/binary"
	"fmt"
)

const (
	KeySize    = 8
	HeaderSize = 8
)

type Kv struct {
	Key  uint64
	Data []byte
}

func NewKv(key uint64, data []byte) Kv {
	var kv = Kv{
		Key:  key,
		Data: data,
	}
	return kv
}

func (kv *Kv) Bytes() []byte {
	var buf = make([]byte, HeaderSize+KeySize+len(kv.Data))
	dataSizeBuf := IntToBytes(len(kv.Data), 8)
	copy(buf[:HeaderSize], dataSizeBuf)
	copy(buf[HeaderSize:HeaderSize+KeySize], IntToBytes(int(kv.Key), KeySize))
	copy(buf[HeaderSize+KeySize:], kv.Data)
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

//字节数(大端)组转成int(无符号的)
func BytesToIntU(b []byte) (uint64, error) {
	if len(b) == 3 {
		b = append([]byte{0}, b...)
	}
	bytesBuffer := bytes.NewBuffer(b)
	switch len(b) {
	case 1:
		var tmp uint8
		err := binary.Read(bytesBuffer, binary.BigEndian, &tmp)
		return uint64(tmp), err
	case 2:
		var tmp uint16
		err := binary.Read(bytesBuffer, binary.BigEndian, &tmp)
		return uint64(tmp), err
	case 4:
		var tmp uint32
		err := binary.Read(bytesBuffer, binary.BigEndian, &tmp)
		return uint64(tmp), err
	case 8:
		var tmp uint64
		err := binary.Read(bytesBuffer, binary.BigEndian, &tmp)
		return uint64(tmp), err
	default:
		return 0, fmt.Errorf("%s", "BytesToInt bytes lenth is invaild!")
	}
}

func IntToBytes(n int, b byte) []byte {
	switch b {
	case 1:
		tmp := int8(n)
		bytesBuffer := bytes.NewBuffer([]byte{})
		binary.Write(bytesBuffer, binary.BigEndian, &tmp)
		return bytesBuffer.Bytes()
	case 2:
		tmp := int16(n)
		bytesBuffer := bytes.NewBuffer([]byte{})
		binary.Write(bytesBuffer, binary.BigEndian, &tmp)
		return bytesBuffer.Bytes()
	case 3, 4:
		tmp := int32(n)
		bytesBuffer := bytes.NewBuffer([]byte{})
		binary.Write(bytesBuffer, binary.BigEndian, &tmp)
		return bytesBuffer.Bytes()
	}
	panic("IntToBytes b param is invaild")
}
