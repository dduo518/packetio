### TCP解析包

##### 使用方式
```GO
pkt := packetio.NewPacketIo(conn)
// OR
pkt := packetio.PktIO(conn)
```
返回一个`PacketIo`接口 
```go
type PacketIo interface {
	Read(ctx context.Context) (*Message, error)
	Write(message *Message) error
}
```
消息数据协议
```go
type Message struct {
	Version      string `json:"v"`  // 版本号 1.0  4字节
	EncodingType int8   `json:"et"` // 消息内容类型 [default：0 json] [1：protobuf] 1字节
	Cmd          uint32 `json:"c"`  // 消息类型 4字节
	Sig          []byte `json:"s"`  // 签名 16 字节
	Time         int64  `json:"t"`  // 时间戳 8字节
	Content      []byte `json:"ct"` // 消息内容
}

```

##### 测试
```go
// 启动服务端
go run ./examples/server/server.go
```

```go
// 启动客户端
go run ./examples/client/client.go
```