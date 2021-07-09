package kvid

import (
	"encoding/hex"
	"errors"
	bytesutils "logkv/bytes-utils"
	"time"
)

var IdErr = errors.New("not a valid kvid")

type Id [12]byte

func NewId(index uint64) Id {
	var ts = time.Now().Unix()
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
