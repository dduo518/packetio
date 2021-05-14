package services

import (
	"context"
	"github.com/opentracing/opentracing-go"
	"go.uber.org/zap"
	"github.com/elvin-zheng/packetio"
)

type Worker interface {
	Dispatch(ctx context.Context, message *packetio.Message)
}

func NewWorker() Worker {
	return &worker{}
}

type worker struct{}

func (w worker) Dispatch(ctx context.Context, message *packetio.Message) {
	switch message.Cmd {
	case 1002:
		sp, _ := opentracing.StartSpanFromContext(ctx, "receive_pong")
		defer sp.Finish()
		zap.L().Info("msg_type", zap.String("TYPE", "PONG"))
	}
}
