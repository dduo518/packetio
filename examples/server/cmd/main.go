package cmd

import (
	"context"
	"fmt"
	"github.com/spf13/viper"
	"go.uber.org/zap"
	"net"
	
	"packetio/examples/server/service"
)

func Start(ctx context.Context) {
	var host = "0.0.0.0"
	var port = viper.GetString("tcp.port")
	var addr = net.JoinHostPort(host, port)
	zap.L().Info(fmt.Sprintf("tcp listening: %s", addr))
	w := service.NewWorker()
	var s = service.NewServer(w)
	if err := s.Start(ctx, addr); err != nil {
		return
	}
	select {}
}
