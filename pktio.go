package packetio

import (
	"bufio"
	"bytes"
	"context"
	"encoding/binary"
	"fmt"
	"github.com/opentracing/opentracing-go"
	"github.com/uber/jaeger-client-go"
	"io"
	"net"
	"time"
)

const defaultReaderSize = 16 * 1024
const defaultWriterSize = 16 * 1024
const defaultReaderTimeout = 60
const defaultHeaderLen = 33

type PackIO struct {
	net.Conn
	reader *bufio.Reader
	writer *bufio.Writer
	opts   *Options
}

func NewPackIO(conn net.Conn, options ...Option) PacketIO {
	pkt := &PackIO{
		Conn:   conn,
		reader: bufio.NewReaderSize(conn, defaultReaderSize),
		writer: bufio.NewWriterSize(conn, defaultWriterSize),
		opts: &Options{
			trace:       false,
			readTimeout: defaultReaderTimeout,
		},
	}
	for _, option := range options {
		option(pkt.opts)
	}
	return pkt
}

func (p *PackIO) Read(ctx context.Context) (context.Context, *Message, error) {
	for {
		select {
		case <-ctx.Done():
			return nil, nil, fmt.Errorf("read end")
		default:
			var header = make([]byte, defaultHeaderLen)
			if p.opts.readTimeout > 0 {
				_ = p.Conn.(*net.TCPConn).SetReadDeadline(time.Now().Add(p.opts.readTimeout * time.Second))
			}
			if _, err := io.ReadFull(p.reader, header[:]); err != nil {
				return nil, nil, err
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
				return nil, nil, err
			}
			var spanCtx opentracing.SpanContext
			var err error
			if p.opts.trace {
				tra := opentracing.GlobalTracer()
				tracer, ok := tra.(*jaeger.Tracer)
				if ok {
					spanCtx, err = tracer.Extract(opentracing.Binary, p.reader)
					if err != nil {
						return nil, nil, err
					}
				}
			}
			
			if !m.check() {
				return nil, nil, fmt.Errorf("uncheck sig message")
			}
			
			if spanCtx != nil {
				sp := opentracing.StartSpan("receive_packet", opentracing.FollowsFrom(spanCtx))
				sp.Finish()
				return opentracing.ContextWithSpan(ctx, sp), m, nil
			}
			return context.TODO(), m, nil
		}
	}
}

func (p *PackIO) Write(ctx context.Context, message *Message) error {
	message.sign()
	var header = make([]byte, 17)
	binary.BigEndian.PutUint32(header[:4], uint32(len(message.Content)))
	binary.BigEndian.PutUint32(header[4:8], message.Cmd)
	binary.BigEndian.PutUint64(header[8:16], uint64(message.Time))
	header[16] = uint8(message.EncodingType)
	
	var lenNum = make([]byte, 0, defaultHeaderLen)
	var buf = bytes.NewBuffer(lenNum)
	if _, err := buf.Write(header); err != nil {
		return err
	}
	
	if _, err := buf.Write(message.Sig); err != nil {
		return err
	}
	if _, err := buf.Write(message.Content); err != nil {
		return err
	}
	if _, err := p.writer.Write(buf.Bytes()); err != nil {
		return err
	}
	
	if p.opts.trace {
		sp := opentracing.SpanFromContext(ctx)
		tra := opentracing.GlobalTracer()
		tracer, ok := tra.(*jaeger.Tracer)
		if ok && sp != nil {
			b := bytes.NewBuffer(make([]byte, 0))
			err := tracer.Inject(sp.Context(), opentracing.Binary, b)
			if err != nil {
				return err
			}
			if _, err := p.writer.Write(b.Bytes()); err != nil {
				return err
			}
		}
	}
	return p.writer.Flush()
}
