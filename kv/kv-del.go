package kv

import (
	"io"
	"os"
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// 删除ts时间之前的数据  重建索引
func (e *KvEngine) Del(ts uint32) error {
	var key = primitive.NewObjectIDFromTimestamp(time.Unix(int64(ts), 0))
	offset, ok := e.indexer.GetMax(key)
	if !ok {
		return ErrNotFound
	}
	e.Lock()
	defer e.Unlock()
	// 备份文件
	f, err := os.OpenFile(e.meta.filename+".bak", os.O_CREATE|os.O_RDWR|os.O_APPEND, os.ModeAppend|os.ModePerm)
	if err != nil {
		return err
	}
	_, err = e.fd.Seek(offset, 0)
	if err != nil {
		return err
	}
	var buf = make([]byte, 1024)
	for {
		n, err := e.fd.Read(buf)
		if err != nil {
			if err == io.EOF {
				break
			}
			return err
		}
		_, err = f.Write(buf[:n])
		if err != nil {
			return err
		}
	}

	err = e.fd.Truncate(0)
	if err != nil {
		return err
	}
	_, err = e.fd.Seek(0, 0)
	if err != nil {
		return err
	}
	_, err = f.Seek(0, 0)
	if err != nil {
		return err
	}
	for {
		n, err := f.Read(buf)
		if err != nil {
			if err == io.EOF {
				break
			}
			return err
		}
		_, err = e.fd.Write(buf[:n])
		if err != nil {
			return err
		}
	}

	return nil
}
