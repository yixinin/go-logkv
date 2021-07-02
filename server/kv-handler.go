package server

import (
	"logkv/protocol"

	"github.com/davyxu/cellnet"
)

func (s *Server) Handle(sess cellnet.Session, msg interface{}) {
	switch msg := msg.(type) {
	case *protocol.Kv:
		f := s.raft.Apply(msg.Data, s.timeout)
		if f.Error() != nil {
			return
		}
		if _, ok := f.Response().(error); ok {
			return
		}
		index := f.Index()
		msg.Key = index
		s.RaftSync(msg)
	}
}

func (s *Server) RaftSync(kv *protocol.Kv) {
	s.ForeachRaftSession(func(sess cellnet.Session) {
		sess.Send(kv)
	})
}
