package service

import (
	"context"
	"go.uber.org/zap"
	"log"
	"net"
	_ "net/http/pprof"
	"github.com/elvin-zheng/packetio"
)

type Service interface {
	Start(ctx context.Context, addr string) error
	Restart() error
	Stop() error
}

func NewServer(worker Worker) *server {
	s := &server{
		worker: worker,
	}
	worker.SetServer(s)
	return s
}

type server struct {
	ln       net.Listener
	addr     string
	address  string
	Metadata interface{}
	worker   Worker
	cc       *ClientConn
}

func (s *server) Start(ctx context.Context, addr string) error {
	var err error
	s.addr = addr
	s.ln, err = net.Listen("tcp", addr)
	if err != nil {
		log.Println(err)
		return err
	}
	go func() {
		defer s.ln.Close()
		for {
			select {
			case <-ctx.Done():
				return
			default:
				conn, err := s.ln.Accept()
				if err != nil {
					return
				}
				zap.L().Info("connect")
				s.cc = NewConn(s, conn)
				go s.cc.clientRun(ctx)
			}
		}
	}()
	return nil
}

func (s *server) Restart() error {
	
	return nil
}

func (s *server) Stop() error {
	return nil
}

func (s *server) Send(ctx context.Context, message *packetio.Message) {
	_ = s.cc.Send(ctx, message)
}
