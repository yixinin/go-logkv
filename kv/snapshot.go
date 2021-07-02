package kv

import (
	"fmt"
	"io"

	"github.com/hashicorp/raft"
)

type snapshot struct {
	r io.Reader
}

func (s *snapshot) Persist(sink raft.SnapshotSink) error {
	if s.r == nil {
		return nil
	}
	defer s.Release()
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
			sink.Cancel()
			return fmt.Errorf("sink.Write(): %v", err)
		}
	}
}

func (s *snapshot) Release() {

}
