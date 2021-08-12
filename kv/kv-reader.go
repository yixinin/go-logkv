package kv

import (
	"bytes"
	"io"
	bytesutils "logkv/bytes-utils"
	"logkv/protocol"

	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/x/bsonx/bsoncore"
)

func ReadIndexes(r io.Reader, set func(key primitive.ObjectID, offset int64)) error {
	var offset int64
	for {
		n, kv, _, err := ReadIndex(r)
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

func ReadIndex(r io.Reader) (int, primitive.ObjectID, []byte, error) {
	var headerBuf = make([]byte, protocol.HeaderSize)
	n, err := r.Read(headerBuf)
	if err != nil {
		return n, primitive.NilObjectID, nil, err
	}
	dataSize, err := bytesutils.BytesToIntU(headerBuf)
	if err != nil {
		return n, primitive.NilObjectID, nil, err
	}

	var data = make([]byte, dataSize)
	n1, err := r.Read(data)
	if err != nil {
		return n + n1, primitive.NilObjectID, nil, err
	}
	doc, err := bsoncore.NewDocumentFromReader(bytes.NewBuffer(data))
	if err != nil {
		return n + n1, primitive.NilObjectID, data, err
	}
	key := doc.Lookup("_id").ObjectID()

	return n + n1, key, data, err
}
