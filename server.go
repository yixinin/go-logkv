package main

import (
	"fmt"
	"log"
	"logkv/protocol"
	"net"
	"sync"
	"time"

	"github.com/davyxu/cellnet"
	"github.com/davyxu/cellnet/peer"
	"github.com/davyxu/cellnet/proc"
	"github.com/hashicorp/raft"
)

type Server struct {
	sync.RWMutex
	sessions map[string]cellnet.Session
	raft     *raft.Raft
	timeout  time.Duration
}

func NewServer(r *raft.Raft, addrs []string) *Server {
	var s = &Server{
		sessions: make(map[string]cellnet.Session, len(addrs)),
		raft:     r,
		timeout:  1 * time.Second,
	}

	return s
}

func (s *Server) Run(addrs []string) {
	go s.InitPeers(addrs)
	go s.Listen(8080)
}

func (s *Server) GetSession(addr string) (cellnet.Session, bool) {
	s.RLock()
	defer s.RUnlock()
	c, ok := s.sessions[addr]
	return c, ok
}

func (s *Server) Listen(port int16) {
	defer func() {
		if r := recover(); r != nil {
			log.Println(r)
		}
	}()
	queue := cellnet.NewEventQueue()
	peerIns := peer.NewGenericPeer("tcp.Acceptor", "server", fmt.Sprintf("0.0.0.0:%d", port), queue)
	proc.BindProcessorHandler(peerIns, "tcp.ltv", func(ev cellnet.Event) {
		switch msg := ev.Message().(type) {
		case *cellnet.SessionAccepted:
			if conn, ok := ev.Session().Raw().(net.Conn); ok {
				addr := conn.RemoteAddr().String()
				ok, err := s.AddSession(addr, ev.Session())
				if err != nil {
					log.Println(err)
					return
				}
				if ok {
					ev.Session().Close()
				}
			}

		case *cellnet.SessionClosed:
			if conn, ok := ev.Session().Raw().(net.Conn); ok {
				addr := conn.RemoteAddr().String()
				s.CloseSession(addr)
			}
		default:
			s.Handle(ev.Session(), msg)
		}
	})
	peerIns.Start()
	queue.StartLoop().Wait()
}

func (s *Server) Connect(addr string) {
	queue := cellnet.NewEventQueue()
	peerIns := peer.NewGenericPeer("tcp.Acceptor", "client", addr, queue)
	proc.BindProcessorHandler(peerIns, "tcp.ltv", func(ev cellnet.Event) {
		switch msg := ev.Message().(type) {
		case *cellnet.SessionConnected, *cellnet.SessionAccepted:
			ok, err := s.AddSession(addr, ev.Session())
			if err != nil {
				log.Println(err)
				return
			}
			if ok {
				ev.Session().Close()
			}
		case *cellnet.SessionClosed, *cellnet.SessionConnectError:
			s.CloseSession(addr)
		default:
			s.Handle(ev.Session(), msg)
		}
	})
	peerIns.Start()
	queue.StartLoop().Wait()
}

func (s *Server) AddSession(addr string, sess cellnet.Session) (bool, error) {
	s.Lock()
	defer s.Unlock()
	_, ok := s.sessions[addr]
	if ok {
		return ok, nil
	}
	s.sessions[addr] = sess
	return false, nil
}

func (s *Server) CloseSession(addr string) {
	s.Lock()
	defer s.Unlock()
	if sess, ok := s.sessions[addr]; ok {
		sess.Close()
	}
	delete(s.sessions, addr)
}

func (s *Server) Foreach(f func(sess cellnet.Session)) {
	s.RLock()
	defer s.RUnlock()
	for _, session := range s.sessions {
		go func(sess cellnet.Session) {
			defer func() {
				if r := recover(); r != nil {
					log.Println(r)
				}
			}()
			f(sess)
		}(session)
	}
}

func (s *Server) ForeachAddrs(addrs []string, f func(addr string, sess cellnet.Session)) {
	s.RLock()
	defer s.RUnlock()
	for _, addr := range addrs {
		go func(addr string, sess cellnet.Session) {
			defer func() {
				if r := recover(); r != nil {
					log.Println(r)
				}
			}()
			f(addr, sess)
		}(addr, s.sessions[addr])
	}

}

func (s *Server) InitPeers(addrs []string) {
	defer func() {
		if r := recover(); r != nil {
			log.Println(r)
		}
	}()
	for {
		var inited = true
		s.ForeachAddrs(addrs, func(addr string, sess cellnet.Session) {
			if sess == nil {
				inited = false
				s.Connect(addr)
			}
		})
		if inited {
			break
		}
	}
}

func (s *Server) Handle(sess cellnet.Session, msg interface{}) {
	var resp interface{}
	defer func() {
		if resp != nil {
			sess.Send(resp)
		}
	}()
	switch msg := msg.(type) {
	case *protocol.AddVoterReq:
		ack := &protocol.AddVoterAck{}
		resp = ack
		f := s.raft.AddVoter(msg.ID, msg.ServerAddress, msg.PrevIndex, s.timeout)
		if err := f.Error(); err != nil {
			ack.Error = err
			return
		}
		ack.Index = f.Index()
	case *protocol.AddNonvoterReq:
		ack := &protocol.AddNonvoterAck{}
		resp = ack
		f := s.raft.AddNonvoter(msg.ID, msg.ServerAddress, msg.PrevIndex, s.timeout)
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
