package kv

import (
	"bytes"
	"log"

	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/x/bsonx/bsoncore"
)

func (e *KvEngine) Set(data []byte) {
	e.ch <- data
}

func (e *KvEngine) BatchSet(datas [][]byte) {
	for _, v := range datas {
		e.ch <- v
	}
}

func (e *KvEngine) receive() {
	for data := range e.ch {
		doc, err := bsoncore.NewDocumentFromReader(bytes.NewBuffer(data))
		if err != nil {
			log.Println(err)
			continue
		}
		var key primitive.ObjectID
		_id := doc.Lookup("_id")
		if len(_id.Data) == 0 {
			key = primitive.NewObjectID()
			bsoncore.AppendObjectIDElement(data, "_id", key)
		} else {
			key = _id.ObjectID()
		}
		e.cache.Set(key.Hex(), data)
	}
}
