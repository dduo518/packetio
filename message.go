package packetio

import (
	"crypto/md5"
	"encoding/hex"
	"io"
	"strconv"
	"time"
)

type Message struct {
	Version      string `json:"v"`  // 版本号 1.0  4字节
	EncodingType int8   `json:"et"` // 消息内容类型 [default：0 json] [1：protobuf] 1字节
	Cmd          uint32 `json:"c"`  // 消息类型 4字节
	Sig          []byte `json:"s"`  // 签名 16 字节
	Time         int64  `json:"t"`  // 时间戳 8字节
	Content      []byte `json:"ct"` // 消息内容
}

func (m *Message) sign() {
	h := md5.New()
	m.Time = time.Now().Unix()
	io.WriteString(h, string(m.Content))
	io.WriteString(h, strconv.Itoa(int(m.Time)))
	io.WriteString(h, MessageSign)
	m.Sig = h.Sum(nil)
	return
}

func (m *Message) check() bool {
	h := md5.New()
	io.WriteString(h, string(m.Content))
	io.WriteString(h, strconv.Itoa(int(m.Time)))
	io.WriteString(h, MessageSign)
	return hex.EncodeToString(h.Sum(nil)) == hex.EncodeToString(m.Sig)
}
