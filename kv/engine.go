package kv

import (
	"io"
	"logkv/protocol"
	"os"
	"sync"

	"github.com/hashicorp/raft"
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
	return e
}

func (e *KvEngine) Set(kv protocol.Kv) error {
	e.Lock()
	defer e.Unlock()
	_, err := e.fd.Write(kv.Bytes())
	return err
}

func (e *KvEngine) RawSet(data []byte) error {
	e.Lock()
	defer e.Unlock()
	_, err := e.fd.Write(data)
	return err
}

func (e *KvEngine) rawDataSet(data []byte) error {
	_, err := e.fd.Write(data)
	return err
}

func (e *KvEngine) BatchSet(kvs protocol.Kvs) error {
	e.Lock()
	defer e.Unlock()
	_, err := e.fd.Write(kvs.Bytes())
	return err
}

func (e *KvEngine) Get(index uint64) (protocol.Kv, error) {

	return protocol.Kv{}, nil
}

func (e *KvEngine) BatchGet(indexs []uint64) (protocol.Kvs, error) {

	return nil, nil
}

func (e *KvEngine) Scan(startIndex, endIndex uint64, limits ...int) (protocol.Kvs, error) {
	// var limit = math.MaxInt64
	// if len(limits) > 0 {
	// 	limit = limits[0]
	// }

	var kvs = make(protocol.Kvs, 0)
	return kvs, nil
}

func (e *KvEngine) rawReader() (io.ReadCloser, error) {
	return os.OpenFile(e.meta.filename, os.O_RDONLY, os.ModePerm)
}

func (e *KvEngine) Apply(log *raft.Log) interface{} {

	e.RawSet(log.Data)
	return nil
}

// Snapshot is used to support log compaction. This call should
// return an FSMSnapshot which can be used to save a point-in-time
// snapshot of the FSM. Apply and Snapshot are not called in multiple
// threads, but Apply will be called concurrently with Persist. This means
// the FSM should be implemented in a fashion that allows for concurrent
// updates while a snapshot is happening.
func (e *KvEngine) Snapshot() (raft.FSMSnapshot, error) {
	var r, err = e.rawReader()
	if err != nil {
		return nil, err
	}
	return &snapshot{r}, nil
}

// Restore is used to restore an FSM from a snapshot. It is not called
// concurrently with any other command. The FSM must discard all previous
// state.
func (e *KvEngine) Restore(r io.ReadCloser) error {
	e.Lock()
	defer e.Unlock()
	var buf = make([]byte, 1024*1024)
	for {
		n, err := r.Read(buf)
		if err != nil {
			if err == io.EOF {
				return nil
			}
			return err
		}
		err = e.rawDataSet(buf[:n])
		if err != nil {
			return err
		}
	}
}
