package service

import (
	"context"
	"github.com/opentracing/opentracing-go"
	"go.uber.org/zap"
	"github.com/elvin-zheng/packetio"
)

type Worker interface {
	Dispatch(ctx context.Context, message *packetio.Message)
	SetServer(s *server)
}

func NewWorker() Worker {
	return &worker{}
}

type worker struct {
	server *server
}

func (w *worker) Dispatch(ctx context.Context, message *packetio.Message) {
	switch message.Cmd {
	case 1001:
		msg := &packetio.Message{
			Cmd:          1002,
			Content:      []byte("pong"),
			EncodingType: 1,
		}
		zap.L().Info("msg_type", zap.String("TYPE", "PING"))
		sp, _ := opentracing.StartSpanFromContext(ctx, "receive_ping")
		defer sp.Finish()
		w.server.Send(ctx, msg)
	}
}

func (w *worker) SetServer(s *server) {
	w.server = s
}
