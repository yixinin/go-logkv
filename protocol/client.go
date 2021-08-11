package protocol

import (
	"reflect"

	"github.com/davyxu/cellnet"
	"github.com/davyxu/cellnet/codec"
	_ "github.com/davyxu/cellnet/codec/binary"
	"github.com/davyxu/cellnet/util"
)

type CodeAck struct {
	Code    uint32
	Message string
}

type SetReq struct {
	Key  string
	Data []byte
}
type SetAck struct {
	CodeAck
}

type BatchSetReq struct {
	Sets []Kv
}
type BatchSetAck struct {
	CodeAck
	Index []uint64
}

type GetReq struct {
	Key string
}

type GetAck struct {
	CodeAck
	Data []byte
}

type BatchGetReq struct {
	Keys []string
}

type BatchGetAck struct {
	CodeAck
	Datas []byte
}

type ScanReq struct {
	StartIndex uint64
	EndIndex   uint64
	StartTime  uint64
	EndTime    uint64
}

type ScanAck struct {
	Datas []GetAck
}

type GetWithIndexReq struct {
	FieldName string
	FieldVal  string
}

type GetWithIndexAck struct {
	Datas []GetAck
}

type ScanWithIndexReq struct {
	FieldName     string
	FieldValStart string
	FieldValEnd   string
}

type ScanWithIndexAck struct {
	Datas []GetAck
}

type DeleteReq struct {
	Time uint32
}

type DeleteAck struct {
	CodeAck
}

type NextReq struct {
	Offset int64
}

type NextAck struct {
	Offset int64
}

func init() {

	cellnet.RegisterMessageMeta(&cellnet.MessageMeta{
		Codec: codec.MustGetCodec("binary"),
		Type:  reflect.TypeOf((*SetReq)(nil)).Elem(),
		ID:    int(util.StringHash("proto.SetReq")),
	})

	cellnet.RegisterMessageMeta(&cellnet.MessageMeta{
		Codec: codec.MustGetCodec("binary"),
		Type:  reflect.TypeOf((*SetAck)(nil)).Elem(),
		ID:    int(util.StringHash("proto.SetAck")),
	})
	cellnet.RegisterMessageMeta(&cellnet.MessageMeta{
		Codec: codec.MustGetCodec("binary"),
		Type:  reflect.TypeOf((*GetReq)(nil)).Elem(),
		ID:    int(util.StringHash("proto.GetReq")),
	})

	cellnet.RegisterMessageMeta(&cellnet.MessageMeta{
		Codec: codec.MustGetCodec("binary"),
		Type:  reflect.TypeOf((*ScanReq)(nil)).Elem(),
		ID:    int(util.StringHash("proto.ScanReq")),
	})
	cellnet.RegisterMessageMeta(&cellnet.MessageMeta{
		Codec: codec.MustGetCodec("binary"),
		Type:  reflect.TypeOf((*ScanAck)(nil)).Elem(),
		ID:    int(util.StringHash("proto.ScanAck")),
	})

}
