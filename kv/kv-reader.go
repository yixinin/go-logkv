package kv

import (
	"io"
	bytesutils "logkv/bytes-utils"
	"logkv/protocol"
)

func Read(r io.Reader, f func(kv protocol.Kv)) error {
	for {
		kv, err := ReadKv(r)
		if err != nil {
			if err == io.EOF {
				return nil
			}
			return err
		}
		f(kv)
	}
}

func ReadKv(r io.Reader) (protocol.Kv, error) {
	var headerBuf = make([]byte, protocol.HeaderSize+protocol.KeySize)
	_, err := r.Read(headerBuf)
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
	_, err = r.Read(data)
	if err != nil {
		return protocol.Kv{}, err
	}
	kv := protocol.NewKv(key, data)
	return kv, nil
}
