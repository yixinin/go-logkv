package server

import (
	"encoding/json"
	"log"
	bytesutils "logkv/bytes-utils"
	"logkv/protocol"

	"github.com/davyxu/cellnet"
)

func (s *Server) Handle(sess cellnet.Session, msg interface{}) {
	switch msg := msg.(type) {
	case *protocol.SetReq:

		indexBuf, err := json.Marshal(msg.Indexes)
		if err != nil {
			log.Println(err)
			return
		}
		var data = make([]byte, protocol.HeaderSize+len(msg.Data)+len(indexBuf))
		var dataSizeBuf = bytesutils.UintToBytes(uint64(len(msg.Data)), 8)
		copy(data[:protocol.HeaderSize], dataSizeBuf)
		copy(data[protocol.HeaderSize:protocol.HeaderSize+len(msg.Data)], msg.Data)
		copy(data[protocol.HeaderSize+len(msg.Data):], indexBuf)

		f := s.raft.Apply(data, s.timeout)
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
