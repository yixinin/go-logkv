package server

import (
	"fmt"
	"log"

	"github.com/davyxu/cellnet"
	"github.com/davyxu/cellnet/peer"
	"github.com/davyxu/cellnet/proc"
)

func (s *Server) AddSession(sess cellnet.Session) {
	s.Lock()
	defer s.Unlock()
	if old, ok := s.session[sess.ID()]; ok {
		old.Close()
	}
	s.session[sess.ID()] = sess
}

func (s *Server) CloseSession(id int64) {
	s.Lock()
	defer s.Unlock()
	if old, ok := s.session[id]; ok {
		old.Close()
	}
	delete(s.session, id)
}
func (s *Server) GetSession(id int64) cellnet.Session {
	s.RLock()
	defer s.RUnlock()
	return s.session[id]
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
			s.AddSession(ev.Session())

		case *cellnet.SessionClosed:
			s.CloseSession(ev.Session().ID())
		default:
			s.Handle(ev.Session(), msg)
		}
	})
	peerIns.Start()
	queue.StartLoop().Wait()
}
