package kv

import (
	"encoding/json"
	"io"
	bytesutils "logkv/bytes-utils"
	"logkv/protocol"
)

func ReadSnapshots(r io.Reader, set func(kv protocol.Kv), setIndex func(i map[string]*Index)) error {
	var indexSizeBuf = make([]byte, 8)
	_, err := r.Read(indexSizeBuf)
	if err != nil {
		return err
	}
	indexSize, err := bytesutils.BytesToIntU(indexSizeBuf)
	if err != nil {
		return err
	}
	var indexBuf = make([]byte, indexSize)
	_, err = r.Read(indexBuf)
	if err != nil {
		return err
	}
	var i = make(map[string]*Index)
	err = json.Unmarshal(indexBuf, &i)
	if err != nil {
		return err
	}

	setIndex(i)

	for {
		kv, err := ReadSnapshot(r)
		if err != nil {
			if err == io.EOF {
				return nil
			}
			return err
		}
		set(kv)
	}
}

func ReadSnapshot(r io.Reader) (protocol.Kv, error) {
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
