package kv

import (
	"context"
	"log"
	"logkv/protocol"
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

func (e *KvEngine) flushTick(ctx context.Context) {
	var ticker = time.NewTicker(time.Second)
	for {
		select {
		case <-ticker.C:
			e.Lock()
			if e.cache.Len() > 1024*10 {
				if err := e.flush(); err != nil {
					log.Println(err)
				}
			}
			e.Unlock()
		case <-ctx.Done():
			return
		}
	}
}

func (e *KvEngine) flush() error {
	// 将缓存的kv拷贝一份 然后写入磁盘
	e.Lock()
	var bucket = make(protocol.Kvs, 0, e.cache.Len())
	var iter = e.cache.ToIter()
	for iter.HasNext() {
		var node = iter.Next()
		key, _ := primitive.ObjectIDFromHex(node.Key())
		bucket = append(bucket, protocol.NewKv(key, node.Val().([]byte)))
	}
	e.Unlock()

	offset, err := e.fd.Seek(0, 2)
	if err != nil {
		return err
	}
	for _, kv := range bucket {
		n, err := e.fd.Write(kv.Bytes())
		if err != nil {
			return err
		}
		e.indexer.Set(kv.Key, offset)
		offset += int64(n)
	}
	e.Lock()
	for _, kv := range bucket {
		e.cache.Delete(kv.Key.Hex())
	}
	e.Unlock()
	return nil
}
