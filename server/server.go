package server

import (
	"logkv/kv"
	"sync"
	"time"

	"github.com/davyxu/cellnet"
	"github.com/hashicorp/raft"
)

type Server struct {
	sync.RWMutex
	raftSessions map[string]cellnet.Session
	session      map[int64]cellnet.Session
	raft         *raft.Raft
	engine       *kv.KvEngine
	timeout      time.Duration
}

func NewServer(r *raft.Raft, engine *kv.KvEngine) *Server {
	var s = &Server{
		raftSessions: make(map[string]cellnet.Session),
		raft:         r,
		engine:       engine,
		timeout:      1 * time.Second,
	}

	return s
}

func (s *Server) Run(addrs []string) {
	go s.InitRaftPeers(addrs)
	go s.ListenRaft(8080)
	go s.Listen(3210)
}
