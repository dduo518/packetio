package packetio

import (
	"bufio"
	"bytes"
	"context"
	"encoding/binary"
	"fmt"
	"io"
	"net"
	"time"
)

const defaultReaderSize = 16 * 1024
const defaultWriterSize = 16 * 1024
const defaultReaderTimeout = time.Second * 30
const headerLength = 41

type PktIO struct {
	net.Conn
	reader      *bufio.Reader
	writer      *bufio.Writer
	readTimeout time.Duration
}

func NewPktIO(conn net.Conn) PacketIo {
	return &PktIO{
		Conn:        conn,
		reader:      bufio.NewReader(conn),
		writer:      bufio.NewWriterSize(conn, defaultWriterSize),
		readTimeout: defaultReaderTimeout,
	}
}

func (p *PktIO) SetReadTimeout(timeout time.Duration) {
	p.readTimeout = timeout
}

func (p *PktIO) SetBufferedReadConn(conn net.Conn) {
	p.Conn = conn
	p.reader = bufio.NewReaderSize(conn, defaultReaderSize)
	p.writer = bufio.NewWriterSize(conn, defaultWriterSize)
	conn.(*net.TCPConn).SetNoDelay(false) // 不使用Nagle算法 禁用Nagle算法后，数据将尽可能快的被发送出去。
}

func (p *PktIO) Read(ctx context.Context) (*Message, error) {
	for {
		select {
		case <-ctx.Done():
			return nil, fmt.Errorf("read end")
		default:
			var header = make([]byte, headerLength)
			if p.readTimeout > 0 {
				_ = p.Conn.(*net.TCPConn).SetReadDeadline(time.Now().Add(p.readTimeout))
			}
			if _, err := io.ReadFull(p.reader, header[:]); err != nil {
				return nil, err
			}

			msgLen := int(binary.BigEndian.Uint32(header[:4]))
			var m = &Message{
				Cmd:          binary.BigEndian.Uint32(header[4:8]),
				EncodingType: int8(header[16]),
				Sig:          header[17:33],
				Time:         int64(binary.BigEndian.Uint64(header[8:16])),
				Content:      make([]byte, msgLen),
			}

			if _, err := io.ReadFull(p.reader, m.Content[:]); err != nil {
				return nil, err
			}

			if !m.check() {
				return nil, fmt.Errorf("uncheck sig message")
			}
			return m, nil
		}
	}
}

func (p *PktIO) Write(message *Message) error {
	message.sign()
	var total = len(message.Content) + headerLength
	var header = make([]byte, 17)
	binary.BigEndian.PutUint32(header[:4], uint32(len(message.Content)))
	binary.BigEndian.PutUint32(header[4:8], message.Cmd)
	binary.BigEndian.PutUint64(header[8:16], uint64(message.Time))
	header[16] = uint8(message.EncodingType)

	var lenNum = make([]byte, 0, total)
	var buf = bytes.NewBuffer(lenNum)
	if _, err := buf.Write(header); err != nil {
		return err
	}

	if _, err := buf.Write(message.Sig); err != nil {
		return err
	}

	if _, err := buf.Write(make([]byte, 8)); err != nil {
		return err
	}

	if _, err := buf.Write(message.Content); err != nil {
		return err
	}

	if count, err := p.writer.Write(buf.Bytes()); err != nil {
		return err
	} else if count != total {
		return fmt.Errorf("write err")
	}
	return p.writer.Flush()
}
