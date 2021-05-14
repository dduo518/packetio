package services

import (
	"context"
	"fmt"
	"github.com/opentracing/opentracing-go"
	"go.uber.org/zap"
	"net"
	"time"
	"github.com/elvin-zheng/packetio"
)

type client struct {
	conn   net.Conn
	pkgio  packetio.PacketIO
	worker Worker
}

func NewClient(ctx context.Context, addr string, w Worker) {
	ctx, cancel := context.WithCancel(ctx)
	conn, err := net.Dial("tcp", addr)
	defer cancel()
	zap.L().Info("connecting:", zap.String("addr", addr))
	if err != nil {
		zap.L().Info("connect err", zap.Error(err))
		return
	}
	zap.L().Info("connected")
	var cl = &client{
		conn:   conn,
		//pkgio:  packetio.NewPackIO(conn, packetio.Trace(true)),
		pkgio:  packetio.NewPacketIo(conn),
		worker: w,
	}
	go cl.sendHeartBeat(ctx)
	if err := cl.read(ctx); err != nil {
		zap.L().Error("conn disconnect", zap.Error(err))
	}
}

func (c *client) sendHeartBeat(ctx context.Context) {
	zap.L().Info("send heart beat")
	var t = time.NewTicker(time.Second * 10)
	defer t.Stop()
	for {
		select {
		case <-ctx.Done():
			zap.L().Info("send heart beat cancel")
			return
		case <-t.C:
			message := &packetio.Message{
				Cmd:          1001,
				Content:      []byte("ping"),
				EncodingType: 1,
			}
			sp := opentracing.StartSpan("client_heartbeat")
			if err := c.pkgio.Write(opentracing.ContextWithSpan(ctx, sp), message); err != nil {
				zap.L().Error("write err", zap.Error(err))
			}
			sp.Finish()
		}
	}
}

func (c *client) read(ctx context.Context) error {
	for {
		select {
		case <-ctx.Done():
			return fmt.Errorf("read close")
		default:
			if ctx, message, err := c.pkgio.Read(ctx); err != nil {
				zap.L().Error("client read err", zap.Error(err))
				return err
			} else {
				c.worker.Dispatch(ctx, message)
			}
		}
	}
}
