package server

import (
	"log"
	"logkv/protocol"

	"github.com/davyxu/cellnet"
)

func (s *Server) Handle(sess cellnet.Session, msg interface{}) {
	switch req := msg.(type) {
	case *protocol.SetReq:

		f := s.raft.Apply(req.Data, s.timeout)
		if err := f.Error(); err != nil {
			log.Println(err)
			return
		}

		switch resp := f.Response().(type) {
		case error:
			log.Println(resp)
			return
		case protocol.Kv:
			s.RaftSync(&resp)
			sess.Send(&resp)
		default:
			log.Println("unknown resp:", resp)
		}
	}
}

func (s *Server) RaftSync(kv *protocol.Kv) {
	s.ForeachRaftSession(func(sess cellnet.Session) {
		sess.Send(kv)
	})
}
