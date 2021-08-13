package kv

import (
	"errors"
	"io"
	"log"
	bytesutils "logkv/bytes-utils"
	"logkv/protocol"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func ReadIndexes(r io.Reader, set func(key primitive.ObjectID, offset int64)) error {
	var offset int64
	for {
		n, key, _, err := ReadIndex(r)
		if err != nil {
			if err == io.EOF {
				return nil
			}
			return err
		}
		log.Println("load", key.Hex(), offset)

		set(key, offset)
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
	n1, err := r.Read(data[4:])
	if err != nil {
		return n + n1, primitive.NilObjectID, nil, err
	}
	copy(data[:4], headerBuf)
	// doc, err := bsoncore.NewDocumentFromReader(bytes.NewBuffer(buf))
	// if err != nil {
	// 	return n + n1, primitive.NilObjectID, data, err
	// }
	// _id := doc.Lookup("_id")

	// key, ok := _id.ObjectIDOK()
	var m = bson.M{}
	err = bson.Unmarshal(data, &m)
	if err != nil {
		return n + n1, primitive.NilObjectID, nil, err
	}
	key, ok := m["_id"].(primitive.ObjectID)
	if !ok {
		return n + n1, primitive.NilObjectID, data, errors.New("not object id")
	}
	return n + n1, key, data, err
}
