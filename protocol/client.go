package protocol

type SetReq struct {
	Data    []byte
	Indexes map[string]string
}
type SetAck struct {
	Index uint64
}

type BatchSetReq struct {
	Sets []SetReq
}
type BatchSetAck struct {
	Index []uint64
}

type GetReq struct {
	Index uint64
	Time  uint64 //unix second
}

type GetAck struct {
	Key  []byte
	Data []byte
}

type BatchGetReq struct {
	Indexes GetReq
}

type BatchGetAck struct {
	Datas []GetAck
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
	Index uint64
	Time  uint64
}

type DeleteAck struct {
	Deleted uint64
}
