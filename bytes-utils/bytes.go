package bytesutils

import (
	"bytes"
	"encoding/binary"
	"fmt"
)

//字节数(大端)组转成int(无符号的)
func BytesToIntU(b []byte) (uint64, error) {
	if len(b) == 3 {
		b = append([]byte{0}, b...)
	}
	bytesBuffer := bytes.NewBuffer(b)
	switch len(b) {
	case 1:
		var tmp uint8
		err := binary.Read(bytesBuffer, binary.LittleEndian, &tmp)
		return uint64(tmp), err
	case 2:
		var tmp uint16
		err := binary.Read(bytesBuffer, binary.LittleEndian, &tmp)
		return uint64(tmp), err
	case 4:
		var tmp uint32
		err := binary.Read(bytesBuffer, binary.LittleEndian, &tmp)
		return uint64(tmp), err
	case 8:
		var tmp uint64
		err := binary.Read(bytesBuffer, binary.LittleEndian, &tmp)
		return uint64(tmp), err
	default:
		return 0, fmt.Errorf("%s", "BytesToInt bytes lenth is invaild!")
	}
}

func BytesToFloat(b []byte) (float64, error) {
	if len(b) == 3 {
		b = append([]byte{0}, b...)
	}
	bytesBuffer := bytes.NewBuffer(b)
	switch len(b) {
	case 4:
		var tmp float32
		err := binary.Read(bytesBuffer, binary.LittleEndian, &tmp)
		return float64(tmp), err
	case 8:
		var tmp float64
		err := binary.Read(bytesBuffer, binary.LittleEndian, &tmp)
		return tmp, err
	default:
		return 0, fmt.Errorf("%s", "BytesToInt bytes lenth is invaild!")
	}
}

func UintToBytes(n uint64, b byte) []byte {
	switch b {
	case 1:
		tmp := int8(n)
		bytesBuffer := bytes.NewBuffer([]byte{})
		binary.Write(bytesBuffer, binary.LittleEndian, &tmp)
		return bytesBuffer.Bytes()
	case 2:
		tmp := int16(n)
		bytesBuffer := bytes.NewBuffer([]byte{})
		binary.Write(bytesBuffer, binary.LittleEndian, &tmp)
		return bytesBuffer.Bytes()
	case 3, 4:
		tmp := int32(n)
		bytesBuffer := bytes.NewBuffer([]byte{})
		binary.Write(bytesBuffer, binary.LittleEndian, &tmp)
		return bytesBuffer.Bytes()
	case 8:
		tmp := int64(n)
		bytesBuffer := bytes.NewBuffer([]byte{})
		binary.Write(bytesBuffer, binary.LittleEndian, &tmp)
		return bytesBuffer.Bytes()
	}
	panic("IntToBytes b param is invaild")
}

func FloatToBytes(n float64, b byte) []byte {
	switch b {
	case 4:
		tmp := float64(n)
		bytesBuffer := bytes.NewBuffer([]byte{})
		binary.Write(bytesBuffer, binary.LittleEndian, &tmp)
		return bytesBuffer.Bytes()
	case 8:
		tmp := n
		bytesBuffer := bytes.NewBuffer([]byte{})
		binary.Write(bytesBuffer, binary.LittleEndian, &tmp)
		return bytesBuffer.Bytes()
	}
	panic("IntToBytes b param is invaild")
}
