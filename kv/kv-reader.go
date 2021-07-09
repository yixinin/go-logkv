package kv

import (
	"encoding/json"
	"io"
	bytesutils "logkv/bytes-utils"
	"logkv/protocol"
)

func ReadSnapshots(r io.Reader, set func(kv protocol.Kv)) error {
	for {
		_, kv, err := ReadSnapshot(r)
		if err != nil {
			if err == io.EOF {
				return nil
			}
			return err
		}
		set(kv)
	}
}

func ReadSnapshot(r io.Reader) (int, protocol.Kv, error) {
	var headerBuf = make([]byte, protocol.HeaderSize*2+protocol.KeySize)
	n, err := r.Read(headerBuf)
	if err != nil {
		return n, protocol.Kv{}, err
	}
	dataSize, err := bytesutils.BytesToIntU(headerBuf[:protocol.HeaderSize])
	if err != nil {
		return n, protocol.Kv{}, err
	}
	indexSize, err := bytesutils.BytesToIntU(headerBuf[protocol.HeaderSize : protocol.HeaderSize*2])
	if err != nil {
		return n, protocol.Kv{}, err
	}
	key, err := bytesutils.BytesToIntU(headerBuf[protocol.HeaderSize*2:])
	if err != nil {
		return n, protocol.Kv{}, err
	}

	var data = make([]byte, dataSize)
	n1, err := r.Read(data)
	if err != nil {
		return n + n1, protocol.Kv{}, err
	}
	kv := protocol.NewKv(key, data)
	var indexBuf = make([]byte, indexSize)
	n2, err := r.Read(indexBuf)
	if err != nil {
		return n + n1 + n2, protocol.Kv{}, err
	}
	err = json.Unmarshal(indexBuf, &kv.Indexes)
	return n + n1 + n2, kv, err
}
