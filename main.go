package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"logkv/kv"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/hashicorp/raft"
	boltdb "github.com/hashicorp/raft-boltdb"
)

var (
	myAddr = flag.String("addr", "localhost:50051", "TCP host+port for this node")
	raftId = flag.String("id", "", "Node id used by Raft")

	raftDir       = flag.String("data_dir", "./data/", "Raft data dir")
	raftBootstrap = flag.Bool("bootstrap", false, "Whether to bootstrap the Raft cluster")
	peers         = flag.String("peers", "", "host:port,")
)

func main() {
	flag.Parse()

	var addrs = strings.Split(*peers, ",")
	if *raftId == "" {
		log.Fatalf("flag --raft_id is required")
	}

	ctx := context.Background()

	var engine = kv.NewKvEngine("")
	r, err := NewRaft(ctx, *raftId, *myAddr, engine)
	if err != nil {
		log.Fatalf("failed to start raft: %v", err)
	}
	s := NewServer(r, addrs)
	s.Run(addrs)
}

func NewRaft(ctx context.Context, myID, myAddress string, fsm raft.FSM) (*raft.Raft, error) {
	c := raft.DefaultConfig()
	c.LocalID = raft.ServerID(myID)

	baseDir := filepath.Join(*raftDir, myID)

	ldb, err := boltdb.NewBoltStore(filepath.Join(baseDir, "logs.dat"))
	if err != nil {
		return nil, fmt.Errorf(`boltdb.NewBoltStore(%q): %v`, filepath.Join(baseDir, "logs.dat"), err)
	}

	sdb, err := boltdb.NewBoltStore(filepath.Join(baseDir, "stable.dat"))
	if err != nil {
		return nil, fmt.Errorf(`boltdb.NewBoltStore(%q): %v`, filepath.Join(baseDir, "stable.dat"), err)
	}

	fss, err := raft.NewFileSnapshotStore(baseDir, 3, os.Stderr)
	if err != nil {
		return nil, fmt.Errorf(`raft.NewFileSnapshotStore(%q, ...): %v`, baseDir, err)
	}

	// tm := transport.New(raft.ServerAddress(myAddress), []grpc.DialOption{grpc.WithInsecure()})
	tm, err := raft.NewTCPTransport(myAddress, nil, 100, time.Second, nil)
	if err != nil {
		return nil, fmt.Errorf("raft.NewTCPTransport: %v", err)
	}
	r, err := raft.NewRaft(c, fsm, ldb, sdb, fss, tm)
	if err != nil {
		return nil, fmt.Errorf("raft.NewRaft: %v", err)
	}

	if *raftBootstrap {
		cfg := raft.Configuration{
			Servers: []raft.Server{
				{
					Suffrage: raft.Voter,
					ID:       raft.ServerID(myID),
					Address:  raft.ServerAddress(myAddress),
				},
			},
		}
		f := r.BootstrapCluster(cfg)
		if err := f.Error(); err != nil {
			return nil, fmt.Errorf("raft.Raft.BootstrapCluster: %v", err)
		}
	}

	return r, nil
}
