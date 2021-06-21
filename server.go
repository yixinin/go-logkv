package main

import (
	"context"
	"log"
	pb "logkv/proto"
	"sync"
	"time"

	"github.com/hashicorp/raft"
	"google.golang.org/grpc"
	"google.golang.org/grpc/peer"
)

type Server struct {
	sync.RWMutex
	clients map[string]pb.PeerClient
	conns   map[string]*grpc.ClientConn
	raft    *raft.Raft
}

func (s *Server) GetClient(addr string) (pb.PeerClient, bool) {
	s.RLock()
	defer s.RUnlock()
	c, ok := s.clients[addr]
	return c, ok
}
func (s *Server) AddClient(addr string) (bool, error) {
	s.Lock()
	defer s.Unlock()
	_, ok := s.clients[addr]
	if ok {
		conn, err := grpc.Dial(addr, grpc.WithInsecure(), grpc.WithBlock())
		if err != nil {
			return false, err
		}
		s.clients[addr] = pb.NewPeerClient(conn)
		s.conns[addr] = conn
		return true, nil
	}
	return ok, nil
}

func (s *Server) ClientFor(f func(pb.PeerClient)) {
	s.RLock()
	defer s.RUnlock()
	for _, client := range s.clients {
		go func(c pb.PeerClient) {
			if r := recover(); r != nil {
				log.Println(r)
			}

			f(c)
		}(client)
	}
}

func NewServer(addrs []string) *Server {
	var s = &Server{}
	go s.InitPeers(addrs)
	return s
}

func (s *Server) InitPeers(addrs []string) {
	defer func() {
		if r := recover(); r != nil {
			log.Println(r)
		}
	}()
	inited := false
	for !inited {
		inited = true
		for _, addr := range addrs {
			ok, err := s.AddClient(addr)
			if err != nil {
				log.Println("add client error", err)
				continue
			}
			if !ok {
				inited = false
			}
		}
	}
}

func (s *Server) AddPeer(ctx context.Context, req *pb.AddPeerReq) (*pb.AddPeerAck, error) {
	p, ok := peer.FromContext(ctx)
	if ok {
		var addr = p.Addr.String()

		if s.raft.State() == raft.Leader {
			s.raft.AddPeer(raft.ServerAddress(addr))
		} else {
			s.raft.AddVoter(req.Id, raft.ServerAddress(addr), 0, time.Second)
		}

		if _, ok := s.clients[addr]; !ok {
			conn, err := grpc.Dial(addr, grpc.WithInsecure(), grpc.WithBlock())
			if err != nil {
				log.Println("did not connect: %v", err)
				return nil, err
			}
			s.Lock()
			s.clients[addr] = pb.NewPeerClient(conn)
			s.conns[addr] = conn
			s.Unlock()
		}
	} else {
		log.Println("get peer not ok")
	}
	return nil, nil
}
func (s *Server) GetPeers(ctxx context.Context, req *pb.GetPeersReq) (*pb.GetPeersAck, error) {
	return nil, nil
}
