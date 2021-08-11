package main

import (
	"context"
	"flag"
	"logkv/kv"
	"logkv/server"

	_ "github.com/davyxu/cellnet/peer/tcp"
	_ "github.com/davyxu/cellnet/proc/tcp"
)

var port int

func main() {
	flag.IntVar(&port, "p", 3210, "port")
	flag.Parse()

	ctx := context.Background()

	var engine = kv.NewKvEngine("log.kv")

	s := server.NewServer(ctx, engine)
	s.Run(int16(port))
}
