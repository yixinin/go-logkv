package kv

import (
	"encoding/hex"
	bytesutils "logkv/bytes-utils"
)

// 删除ts时间之前的数据  重建索引
func (e *KvEngine) Del(ts uint64, index uint64) error {
	tsb := bytesutils.UintToBytes(ts, 8)
	idxb := bytesutils.UintToBytes(index, 8)
	var keysb = make([]byte, 16)
	copy(keysb[:8], tsb)
	copy(keysb[8:], idxb)
	var key = hex.EncodeToString(keysb)
	offset := e.indexer.Get("index", key)

	return nil
}
