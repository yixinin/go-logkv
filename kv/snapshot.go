package kv

import (
	"encoding/json"
	"fmt"
	"io"
	bytesutils "logkv/bytes-utils"

	"github.com/hashicorp/raft"
)

type snapshot struct {
	r io.Reader
	i map[string]map[string][]int64
}

func (s *snapshot) Persist(sink raft.SnapshotSink) error {
	if s.r == nil {
		return nil
	}
	defer s.Release()
	var indexBuf, err = json.Marshal(s.i)
	if err != nil {
		return err
	}
	var indexSize = len(indexBuf)
	n, err := sink.Write(bytesutils.IntToBytes(indexSize, 8))
	if err != nil {
		return err
	}
	if n != 8 {
		return fmt.Errorf("write index header size buf error, n=%d", n)
	}
	n, err = sink.Write(indexBuf)
	if err != nil {
		return err
	}
	if n != len(indexBuf) {
		return fmt.Errorf("write index header size buf error, n=%d", n)
	}
	var buf = make([]byte, 1024*1024)
	for {
		n, err := s.r.Read(buf)
		if err != nil {
			if err == io.EOF {
				return nil
			}
			return err
		}

		_, err = sink.Write(buf[:n])
		if err != nil {
			return fmt.Errorf("sink.Write(): %v", err)
		}
	}
}

func (s *snapshot) Release() {

}
