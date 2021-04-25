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
const defaultReaderTimeout = 60

type PackIO struct {
	net.Conn
	reader      *bufio.Reader
	writer      *bufio.Writer
	readTimeout time.Duration
}

func NewPackIO(conn net.Conn) *PackIO {
	return &PackIO{
		Conn:        conn,
		reader:      bufio.NewReader(conn),
		writer:      bufio.NewWriterSize(conn, defaultWriterSize),
		readTimeout: defaultReaderTimeout,
	}
}

func (p *PackIO) SetReadTimeout(timeout time.Duration) {
	p.readTimeout = timeout
}

func (p *PackIO) SetBufferedReadConn(conn net.Conn) {
	p.Conn = conn
	p.reader = bufio.NewReaderSize(conn, defaultReaderSize)
	p.writer = bufio.NewWriterSize(conn, defaultWriterSize)
	//Nagle算法的做法是：
	//将要发送的小包合并，并延缓发送。
	//延缓后的发送策略是，收到前一个发送出去的包的ACK确认包，
	//或者一定时间后，收集了足够数量的小数据包。
	//Nagle算法的目的是减少发送小包的数量，从而减小带宽，并提高网络吞吐量，
	//付出的代价是有时会增加服务的延时。
	//补充解释一下为什么减少小包的数量可以减小带宽。
	//因为每个TCP包，除了包体中包含的应用层数据外，外层还要套上TCP包头和IP包头。
	//由于应用层要发送的业务数据量是固定的，所以包数量越多，包头占用的带宽也越多
	// 从策略上我们可以使用bufio.write 使用bufio.Writer还有一个好处，就是减少了调用write系统调用的次数，
	//但是相应的，增加了数据拷贝的开销
	conn.(*net.TCPConn).SetNoDelay(false) // 不使用Nagle算法 禁用Nagle算法后，数据将尽可能快的被发送出去。
}

func (p *PackIO) Read(ctx context.Context) (*Message, error) {
	for {
		select {
		case <-ctx.Done():
			return nil, fmt.Errorf("read end")
		default:
			var header = make([]byte, HeaderLen)
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

func (p *PackIO) Write(message *Message) error {
	message.sign()
	var total = len(message.Content) + HeaderLen
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
