package kv

import (
	"io"
	"logkv/protocol"

	"github.com/hashicorp/raft"
)

func (e *KvEngine) Apply(log *raft.Log) interface{} {
	var kv, err = protocol.NewKvWithIndexes(log.Index, log.Data)
	if err != nil {
		return err
	}
	return e.Set(kv)
}

// Snapshot is used to support log compaction. This call should
// return an FSMSnapshot which can be used to save a point-in-time
// snapshot of the FSM. Apply and Snapshot are not called in multiple
// threads, but Apply will be called concurrently with Persist. This means
// the FSM should be implemented in a fashion that allows for concurrent
// updates while a snapshot is happening.
func (e *KvEngine) Snapshot() (raft.FSMSnapshot, error) {
	r, err := e.rawFileReader()
	if err != nil {
		return nil, err
	}
	return &snapshot{r}, nil
}

// Restore is used to restore an FSM from a snapshot. It is not called
// concurrently with any other command. The FSM must discard all previous
// state.
func (e *KvEngine) Restore(r io.ReadCloser) error {
	defer r.Close()
	// 清除缓存
	if err := e.fd.Truncate(0); err != nil {
		return err
	}
	if _, err := e.fd.Seek(0, 0); err != nil {
		return err
	}

	err := ReadSnapshots(r, func(kv protocol.Kv) {
		err := e.Set(kv)
		if err != nil {
			panic(err)
		}
	})
	return err
}