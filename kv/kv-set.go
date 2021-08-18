package kv

import (
	"bytes"
	"log"

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
		_id := doc.Lookup("_id").ObjectID()
		trace := doc.Lookup("trace").String()
		if trace != "" {
			e.indexer.SetTrace(trace, _id)
		}
		e.cache.Set(_id, data)
	}
}
