package server

import (
	"log"
	"logkv/protocol"

	"github.com/davyxu/cellnet"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func (s *Server) Handle(sess cellnet.Session, msg interface{}) {
	switch req := msg.(type) {
	//set
	case *protocol.SetReq:
		var ack = protocol.SetAck{}
		key, err := primitive.ObjectIDFromHex(req.Key)
		if err != nil {
			panic(err)
		}
		err = s.engine.Set(protocol.NewKv(key, req.Data))
		if err != nil {
			ack.Code = 400
			ack.Message = err.Error()
		}
		sess.Send(&ack)

	//get
	case *protocol.GetReq:
		var ack = &protocol.GetAck{}
		key, err := primitive.ObjectIDFromHex(req.Key)
		if err != nil {
			panic(err)
		}
		v, err := s.engine.Get(key)
		if err != nil {
			ack.Code = 400
			ack.Message = err.Error()
		}
		ack.Data = v.Bytes()
		sess.Send(ack)
	//delete
	case *protocol.DeleteReq:
		var ack protocol.DeleteAck
		err := s.engine.Del(req.Time)
		if err != nil {
			ack.Code = 400
			ack.Message = err.Error()
		}
		sess.Send(&ack)
	//batchget
	case *protocol.BatchGetReq:
		// var ack protocol.BatchGetAck
		// vs, err := s.engine.BatchGet(req.Keys)
		// if err != nil {
		// 	ack.Code = 400
		// 	ack.Message = err.Error()
		// }
		// ack.Datas = vs.Bytes()
	//batchset
	case *protocol.BatchSetReq:
		var ack protocol.BatchSetAck
		err := s.engine.BatchSet(req.Sets)
		if err != nil {
			ack.Code = 400
			ack.Message = err.Error()
		}
		sess.Send(&ack)
	//scan
	case *protocol.ScanReq:

	default:
		log.Println("unkown msg", req)
		return
	}
}
