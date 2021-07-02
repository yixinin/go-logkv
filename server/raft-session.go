package server

import (
	"fmt"
	"log"
	"net"

	"github.com/davyxu/cellnet"
	"github.com/davyxu/cellnet/peer"
	"github.com/davyxu/cellnet/proc"
)

func (s *Server) GetRaftSession(addr string) (cellnet.Session, bool) {
	s.RLock()
	defer s.RUnlock()
	c, ok := s.raftSessions[addr]
	return c, ok
}

func (s *Server) ListenRaft(port int16) {
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
				ok, err := s.AddRaftSession(addr, ev.Session())
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
				s.CloseRaftSession(addr)
			}
		default:
			s.HandleRaft(ev.Session(), msg)
		}
	})
	peerIns.Start()
	queue.StartLoop().Wait()
}

func (s *Server) ConnectRaft(addr string) {
	queue := cellnet.NewEventQueue()
	peerIns := peer.NewGenericPeer("tcp.Acceptor", "client", addr, queue)
	proc.BindProcessorHandler(peerIns, "tcp.ltv", func(ev cellnet.Event) {
		switch msg := ev.Message().(type) {
		case *cellnet.SessionConnected, *cellnet.SessionAccepted:
			ok, err := s.AddRaftSession(addr, ev.Session())
			if err != nil {
				log.Println(err)
				return
			}
			if ok {
				ev.Session().Close()
			}
		case *cellnet.SessionClosed, *cellnet.SessionConnectError:
			s.CloseRaftSession(addr)
		default:
			s.HandleRaft(ev.Session(), msg)
		}
	})
	peerIns.Start()
	queue.StartLoop().Wait()
}

func (s *Server) AddRaftSession(addr string, sess cellnet.Session) (bool, error) {
	s.Lock()
	defer s.Unlock()
	_, ok := s.raftSessions[addr]
	if ok {
		return ok, nil
	}
	s.raftSessions[addr] = sess
	return false, nil
}

func (s *Server) CloseRaftSession(addr string) {
	s.Lock()
	defer s.Unlock()
	if sess, ok := s.raftSessions[addr]; ok {
		sess.Close()
	}
	delete(s.raftSessions, addr)
}

func (s *Server) ForeachRaftSession(f func(sess cellnet.Session)) {
	s.RLock()
	defer s.RUnlock()
	for _, session := range s.raftSessions {
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

func (s *Server) ForeachRaftSessionAddrs(addrs []string, f func(addr string, sess cellnet.Session)) {
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
		}(addr, s.raftSessions[addr])
	}

}

func (s *Server) InitRaftPeers(addrs []string) {
	defer func() {
		if r := recover(); r != nil {
			log.Println(r)
		}
	}()
	for {
		var inited = true
		s.ForeachRaftSessionAddrs(addrs, func(addr string, sess cellnet.Session) {
			if sess == nil {
				inited = false
				s.ConnectRaft(addr)
			}
		})
		if inited {
			break
		}
	}
}
