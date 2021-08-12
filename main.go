package main

import (
	"context"
	"flag"
	"logkv/kv"
	"logkv/server"
	"os"
	"os/signal"

	_ "github.com/davyxu/cellnet/peer/tcp"
	_ "github.com/davyxu/cellnet/proc/tcp"
)

var port int

func main() {
	flag.IntVar(&port, "p", 3210, "port")
	flag.Parse()

	ctx, cancel := context.WithCancel(context.Background())

	var engine = kv.NewKvEngine(ctx, "log.kv")

	s := server.NewServer(ctx, engine)
	go s.Run(int16(port))
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	<-c
	cancel()
	s.Close()
}
