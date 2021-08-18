package kv

import (
	"bytes"
	"errors"
	"io"
	"log"
	bytesutils "logkv/bytes-utils"
	"logkv/protocol"

	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/x/bsonx/bsoncore"
)

func ReadIndexes(r io.Reader, traceKey string, set func(key primitive.ObjectID, trace string, offset int64)) error {
	var offset int64
	for {
		n, key, trace, _, err := ReadIndex(r, traceKey)
		if err != nil {
			if err == io.EOF {
				return nil
			}
			return err
		}
		log.Println("load", key.Hex(), offset)

		set(key, trace, offset)
		offset += int64(n)
	}
}

func ReadIndex(r io.Reader, traceKey string) (int, primitive.ObjectID, string, []byte, error) {
	var headerBuf = make([]byte, protocol.HeaderSize)
	n, err := r.Read(headerBuf)
	if err != nil {
		return n, primitive.NilObjectID, "", nil, err
	}
	dataSize, err := bytesutils.BytesToIntU(headerBuf)
	if err != nil {
		return n, primitive.NilObjectID, "", nil, err
	}

	var data = make([]byte, dataSize)
	n1, err := r.Read(data[4:])
	if err != nil {
		return n + n1, primitive.NilObjectID, "", nil, err
	}
	copy(data[:4], headerBuf)
	doc, err := bsoncore.NewDocumentFromReader(bytes.NewBuffer(data))
	if err != nil {
		return n + n1, primitive.NilObjectID, "", data, err
	}
	_id := doc.Lookup("_id")

	key, ok := _id.ObjectIDOK()
	if !ok {
		return n + n1, primitive.NilObjectID, "", data, errors.New("not object id")
	}
	var trace string
	if traceKey != "" {
		trace = doc.Lookup(traceKey).String()
	}
	return n + n1, key, trace, data, err
}
