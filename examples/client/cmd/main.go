package cmd

import (
	"context"
	"fmt"
	"github.com/spf13/viper"
	"go.uber.org/zap"
	"net"
	"time"
	"packetio/examples/client/services"
)

func Start(ctx context.Context) {
	var host = viper.GetString("tcp.host")
	var port = viper.GetString("tcp.port")
	var addr = net.JoinHostPort(host, port)
	zap.L().Info(fmt.Sprintf("tcp connecting: %s", addr))
	worker := services.NewWorker()
	for {
		services.NewClient(ctx, addr, worker)
		time.Sleep(time.Second)
		zap.L().Debug("reconnect..")
	}
}
