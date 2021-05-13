package packetio

import (
	"bufio"
	"bytes"
	"context"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"github.com/opentracing/opentracing-go"
	"io"
	"net"
)

const (
	HeaderLen   = 4
	Version     = "1.0.0"
	MessageSign = "!@QESEFDSAID#$134"
)

func NewPacketIo(conn net.Conn) PacketIO {
	p := &PacketIo{
		scan: bufio.NewScanner(conn),
		w:    bufio.NewWriter(conn),
	}
	p.scan.Split(p.split)
	return p
}

type PacketIo struct {
	scan *bufio.Scanner
	w    *bufio.Writer
}

func (p *PacketIo) Read(ctx context.Context) (context.Context, *Message, error) {
	for p.scan.Scan() {
		select {
		case <-ctx.Done():
			return nil, nil, fmt.Errorf("read closed")
		default:
			err := p.scan.Err()
			if err != nil && err != io.EOF {
				return nil, nil, err
			}
			
			bs := p.scan.Bytes()
			var msg = &Message{}
			if err := json.Unmarshal(bs, msg); err != nil {
				return nil, nil, err
			}
			
			if msg.Metadata != nil {
				spanCtx, err := Extract(msg.Metadata)
				if err != nil {
					return nil, nil, err
				}
				if spanCtx != nil {
					sp := opentracing.StartSpan("receive_packet", opentracing.FollowsFrom(spanCtx))
					sp.Finish()
					return opentracing.ContextWithSpan(ctx, sp), msg, nil
				}
			}
			return context.TODO(), msg, nil
		}
	}
	return nil, nil, fmt.Errorf("read err")
}

func (p *PacketIo) Write(ctx context.Context, m *Message) error {
	
	if sp := opentracing.SpanFromContext(ctx); sp != nil {
		m.Metadata = make(map[string]interface{})
		_ = Inject(sp, m.Metadata)
	}
	if bs, err := json.Marshal(m); err != nil {
		return err
	} else {
		var lenNum = make([]byte, HeaderLen)
		binary.BigEndian.PutUint32(lenNum, uint32(len(bs)))
		var buf = bytes.NewBuffer(lenNum)
		_, _ = buf.Write(bs)
		if _, err := p.w.Write(buf.Bytes()); err != nil {
			return err
		}
		return p.w.Flush()
	}
}

func (p *PacketIo) split(data []byte, atEOF bool) (advance int, token []byte, err error) {
	if len(data) < 4 {
		return
	}
	length := binary.BigEndian.Uint32(data[:HeaderLen])
	if !atEOF && length == uint32(len(data[HeaderLen:])) {
		return len(data), data[HeaderLen:], nil
	}
	return
}
