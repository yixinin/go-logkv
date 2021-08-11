package kv

import (
	"io"
	bytesutils "logkv/bytes-utils"
	"logkv/protocol"

	"gopkg.in/mgo.v2/bson"
)

func ReadKvs(r io.Reader, set func(kv protocol.Kv)) error {
	for {
		_, kv, err := ReadKv(r)
		if err != nil {
			if err == io.EOF {
				return nil
			}
			return err
		}
		set(kv)
	}
}

func ReadKv(r io.Reader) (int, protocol.Kv, error) {
	var headerBuf = make([]byte, protocol.HeaderSize+protocol.KeySize)
	n, err := r.Read(headerBuf)
	if err != nil {
		return n, protocol.Kv{}, err
	}
	dataSize, err := bytesutils.BytesToIntU(headerBuf[:protocol.HeaderSize])
	if err != nil {
		return n, protocol.Kv{}, err
	}

	key := bson.ObjectIdHex(string(headerBuf[protocol.HeaderSize:]))

	var data = make([]byte, dataSize)
	n1, err := r.Read(data)
	if err != nil {
		return n + n1, protocol.Kv{}, err
	}
	kv := protocol.NewKv(key, data)

	return n + n1, kv, err
}
