package kv

import (
	"context"
	"log"
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
	var keys = make([]primitive.ObjectID, 0, e.cache.Len())
	var bucket = make([][]byte, 0, e.cache.Len())
	var iter = e.cache.ToIter()
	for iter.HasNext() {
		var node = iter.Next()
		keys = append(keys, node.Key())
		bucket = append(bucket, node.Val().([]byte))
	}
	e.Unlock()

	offset, err := e.fd.Seek(0, 2)
	if err != nil {
		return err
	}
	for i, data := range bucket {
		n, err := e.fd.Write(data)
		if err != nil {
			return err
		}
		e.indexer.Set(keys[i], offset)
		offset += int64(n)
	}
	e.Lock()
	for _, key := range keys {
		e.cache.Del(key)
	}
	e.Unlock()
	return nil
}
