package server

import (
	"log"
	"logkv/protocol"
	"time"

	"github.com/davyxu/cellnet"
	"github.com/hashicorp/raft"
)

func (s *Server) HandleRaft(sess cellnet.Session, msg interface{}) {
	var resp interface{}
	defer func() {
		if resp != nil {
			sess.Send(resp)
		}
	}()
	switch msg := msg.(type) {
	case *protocol.AddNonvoterReq:
		ack := &protocol.AddNonvoterAck{}
		resp = ack
		f := s.raft.AddNonvoter(msg.ID, msg.ServerAddress, msg.PrevIndex, s.timeout)
		if err := f.Error(); err != nil {
			ack.Error = err
			return
		}
		ack.Index = f.Index()

	case *protocol.AddVoterReq:
		ack := &protocol.AddVoterAck{}
		resp = ack
		f := s.raft.AddVoter(msg.ID, msg.ServerAddress, msg.PrevIndex, s.timeout)
		if err := f.Error(); err != nil {
			ack.Error = err
			return
		}
		ack.Index = f.Index()

	case *protocol.AppliedIndexReq:
		index := s.raft.AppliedIndex()
		resp = &protocol.AppliedIndexAck{Index: index}
	case *protocol.BarrierReq:
		err := s.raft.Barrier(s.timeout).Error()
		resp = &protocol.BarrierAck{Error: err}
	case *protocol.DemoteVoterReq:
		ack := &protocol.DemoteVoterAck{}
		resp = ack
		f := s.raft.DemoteVoter(msg.Id, msg.PreviousIndex, s.timeout)
		if err := f.Error(); err != nil {
			ack.Error = err
			return
		}
		ack.Index = f.Index()
	case *protocol.RemoveServerReq:
		ack := &protocol.RemoveServerAck{}
		resp = ack
		f := s.raft.RemoveServer(msg.Id, msg.PreviousIndex, s.timeout)
		if err := f.Error(); err != nil {
			ack.Error = err
			return
		}
		ack.Index = f.Index()
	case *raft.Log:
		ack := &protocol.ApplyLogAck{}
		resp = ack
		f := s.raft.ApplyLog(*msg, s.timeout)
		if err := f.Error(); err != nil {
			ack.Error = err
			return
		}
		ack.Response = f.Response()
		ack.Index = f.Index()
	case *protocol.Kv:
		ack := &protocol.ApplyAck{}
		resp = ack
		f := s.raft.Apply(msg.Bytes(), s.timeout)
		if err := f.Error(); err != nil {
			ack.Error = err
			return
		}
		ack.Response = f.Response()
	case *protocol.Kvs:
		ack := &protocol.ApplyAck{}
		resp = ack
		f := s.raft.Apply(msg.Bytes(), time.Duration(len(*msg))*s.timeout)
		if err := f.Error(); err != nil {
			ack.Error = err
			return
		}
		ack.Response = f.Response()
	default:
		s.HandleAck(sess, msg)
	}
}

func (s *Server) HandleAck(sess cellnet.Session, msg interface{}) {
	switch msg := msg.(type) {
	case protocol.AddNonvoterAck:
		log.Println(msg)
	}
}
