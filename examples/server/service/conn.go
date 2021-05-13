package service

import (
	"context"
	"crypto/tls"
	"go.uber.org/zap"
	"net"
	"sync/atomic"
	"time"
	"packetio"
)

var connId uint64

func NewConn(srv *server, conn net.Conn) *ClientConn {
	return &ClientConn{
		conn:      conn,
		pkt:       nil,
		tlsConn:   nil,
		connectId: atomic.AddUint64(&connId, 1),
		server:    srv,
	}
}

type ClientConn struct {
	conn      net.Conn
	pkt       packetio.PacketIO
	tlsConn   *tls.Conn
	connectId uint64
	Metadata  interface{}
	server    *server
	status    uint8
}

func (cc *ClientConn) setConn(conn net.Conn) {
	cc.conn = conn
}

func (cc *ClientConn) upgradeToTLS(tlsConfig *tls.Config) error {
	tlsConn := tls.Server(cc.conn, tlsConfig)
	if err := tlsConn.Handshake(); err != nil {
		return err
	}
	cc.setConn(tlsConn)
	cc.tlsConn = tlsConn
	return nil
}

func (cc *ClientConn) clientRun(ctx context.Context) {
	ctx, cancel := context.WithCancel(ctx)
	defer func() {
		zap.L().Info("disconnect")
		cancel()
		_ = cc.conn.Close()
	}()
	
	if c, ok := cc.conn.(*net.TCPConn); ok {
		_ = c.SetReadBuffer(8 * 1024)
		_ = c.SetWriteBuffer(8 * 1024)
	}
	
	//cc.pkt = packetio.NewPackIO(cc.conn, packetio.Trace(true))
	cc.pkt = packetio.NewPacketIo(cc.conn)
	
	for {
		_ = cc.conn.SetReadDeadline(time.Now().Add(time.Second * 30))
		ctx, msg, err := cc.pkt.Read(ctx)
		if err != nil {
			zap.L().Error("read error", zap.Error(err))
			return
		}
		cc.server.worker.Dispatch(ctx, msg)
	}
}

func (cc *ClientConn) Send(ctx context.Context, m *packetio.Message) error {
	return cc.pkt.Write(ctx, m)
}
