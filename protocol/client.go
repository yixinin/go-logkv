package protocol

type SetReq struct {
	Data    []byte
	Indexes map[string]string
}
