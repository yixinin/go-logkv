package kv

import (
	"bytes"
	"io"
	"logkv/protocol"
	"os"
	"sync"
)

type EngineMeta struct {
	filename string
}

type KvEngine struct {
	sync.Mutex
	meta    EngineMeta
	fd      *os.File
	indexer *KvIndexer
}

func NewKvEngine(filename string) *KvEngine {
	e := &KvEngine{
		meta: EngineMeta{
			filename: filename,
		},
		indexer: NewKvIndexer(),
	}
	var err error
	e.fd, err = os.OpenFile(filename, os.O_CREATE|os.O_RDWR|os.O_APPEND, os.ModeAppend|os.ModePerm)
	if err != nil {
		panic(err)
	}
	e.load()
	return e
}

func (e *KvEngine) load() {
	ReadKvs(e.fd, func(kv protocol.Kv) {
		e.Set(kv)
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
