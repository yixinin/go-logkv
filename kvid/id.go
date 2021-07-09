package kvid

import (
	"encoding/hex"
	"errors"
	bytesutils "logkv/bytes-utils"
)

var IdErr = errors.New("not a valid kvid")

type Id [12]byte

func NewId(ts uint32, index uint64) Id {
	var id = Id{}
	copy(id[:4], bytesutils.UintToBytes(uint64(ts), 4))
	copy(id[4:], bytesutils.UintToBytes(index, 8))
	return id
}

func (id Id) Hex() string {
	return hex.EncodeToString(id[:])
}

func (id Id) Meta() (uint32, uint64) {
	ts, _ := bytesutils.BytesToIntU(id[:4])
	index, _ := bytesutils.BytesToIntU(id[4:])
	return uint32(ts), index
}

func (id Id) Ts() uint32 {
	ts, _ := bytesutils.BytesToIntU(id[:4])
	return uint32(ts)
}

func (id Id) Index() uint64 {
	index, _ := bytesutils.BytesToIntU(id[4:])
	return index
}

func FromHex(s string) (Id, error) {
	if len(s) != 24 {
		return [12]byte{}, IdErr
	}
	bs, err := hex.DecodeString(s)
	if err != nil {
		return [12]byte{}, err
	}
	if len(bs) != 12 {
		return [12]byte{}, IdErr
	}
	var id = [12]byte{}
	copy(id[:], bs)
	return id, nil
}

func FromBytes(buf []byte) Id {
	var id = Id{}
	copy(id[:], buf)
	return id
}

func TsHex(ts uint32) string {
	return hex.EncodeToString(bytesutils.UintToBytes(uint64(ts), 4))
}
