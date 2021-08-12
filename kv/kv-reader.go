package kv

import (
	"io"
	bytesutils "logkv/bytes-utils"
	"logkv/protocol"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

func ReadKvs(r io.Reader, set func(kv protocol.Kv, offset int64)) error {
	var offset int64
	for {
		n, kv, err := ReadKv(r)
		if err != nil {
			if err == io.EOF {
				return nil
			}
			return err
		}

		set(kv, offset)
		offset += int64(n)
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

	var key primitive.ObjectID
	copy(key[:], headerBuf[protocol.HeaderSize:])

	var data = make([]byte, dataSize)
	n1, err := r.Read(data)
	if err != nil {
		return n + n1, protocol.Kv{}, err
	}
	kv := protocol.NewKv(key, data)

	return n + n1, kv, err
}
