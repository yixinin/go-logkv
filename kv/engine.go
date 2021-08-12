package kv

import (
	"bytes"
	"context"
	"io"
	"log"
	"logkv/protocol"
	"logkv/skipmap"
	"os"
	"sync"
	"time"
)

type EngineMeta struct {
	filename string
}

type KvEngine struct {
	sync.Mutex
	meta    EngineMeta
	fd      *os.File
	indexer *KvIndexer

	cache *skipmap.Skipmap

	ch chan protocol.Kv
}

func (e *KvEngine) Close() {
	close(e.ch)
	for len(e.ch) > 0 {
		time.Sleep(1 * time.Millisecond)
	}
	if err := e.flush(); err != nil {
		log.Println(err)
	}
}

func NewKvEngine(ctx context.Context, filename string) *KvEngine {
	e := &KvEngine{
		meta: EngineMeta{
			filename: filename,
		},
		indexer: NewKvIndexer(),
		cache:   skipmap.New(),
		ch:      make(chan protocol.Kv, 1024*1024),
	}
	var err error
	e.fd, err = os.OpenFile(filename, os.O_CREATE|os.O_RDWR|os.O_APPEND, os.ModeAppend|os.ModePerm)
	if err != nil {
		panic(err)
	}
	e.initIndexes()
	go e.flushTick(ctx)
	go e.receive()
	return e
}

func (e *KvEngine) initIndexes() {
	ReadKvs(e.fd, func(kv protocol.Kv, offset int64) {
		e.indexer.Set(kv.Key, offset)
	})
}

func (e *KvEngine) rawFileReader() (io.Reader, error) {
	var reader bytes.Buffer
	f, err := os.OpenFile(e.meta.filename, os.O_RDONLY, os.ModePerm)
	if err != nil {
		return nil, err
	}
	_, err = io.Copy(&reader, f)
	if err != nil {
		return nil, err
	}
	return &reader, nil
}
