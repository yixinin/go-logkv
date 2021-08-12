package server

import (
	"context"
	"logkv/kv"
	"sync"
	"time"

	"github.com/davyxu/cellnet"
)

type Server struct {
	sync.RWMutex
	session  map[int64]cellnet.Session
	engine   *kv.KvEngine
	timeout  time.Duration
	tcpQueue cellnet.EventQueue
}

func NewServer(ctx context.Context, engine *kv.KvEngine) *Server {
	var s = &Server{
		session: make(map[int64]cellnet.Session),
		engine:  engine,
		timeout: 1 * time.Second,
	}

	return s
}

func (s *Server) Run(port int16) {
	s.Listen(port)
}

func (s *Server) Close() {
	s.tcpQueue.StopLoop()
	time.Sleep(1 * time.Second)
	s.engine.Close()
}
