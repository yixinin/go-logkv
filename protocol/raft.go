package protocol

import "github.com/hashicorp/raft"

type AddNonvoterReq struct {
	ID            raft.ServerID
	ServerAddress raft.ServerAddress
	PrevIndex     uint64
}
type AddNonvoterAck struct {
	Index uint64
	Error error
}

type AddVoterReq struct {
	ID            raft.ServerID
	ServerAddress raft.ServerAddress
	PrevIndex     uint64
}
type AddVoterAck struct {
	Index uint64
	Error error
}

type AppliedIndexReq struct {
}

type AppliedIndexAck struct {
	Index uint64
}

type BarrierReq struct {
}
type BarrierAck struct {
	Index uint64
	Error error
}

type DemoteVoterReq struct {
	Id            raft.ServerID
	PreviousIndex uint64
}
type DemoteVoterAck struct {
	Index uint64
	Error error
}

type RemoveServerReq struct {
	Id            raft.ServerID
	PreviousIndex uint64
}
type RemoveServerAck struct {
	Index uint64
	Error error
}

type ApplyLogAck struct {
	Index    uint64
	Error    error
	Response interface{}
}

type ApplyAck struct {
	Index    uint64
	Error    error
	Response interface{}
}
