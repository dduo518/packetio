package main

import (
	"context"
	"fmt"
	"github.com/joho/godotenv"
	"github.com/opentracing/opentracing-go"
	"github.com/spf13/viper"
	"github.com/uber/jaeger-client-go"
	jaegercfg "github.com/uber/jaeger-client-go/config"
	"github.com/uber/jaeger-lib/metrics"
	"go.uber.org/zap"
	"strings"
	"time"
	
	client "packetio/examples/client/cmd"
	server "packetio/examples/server/cmd"
)

var serverName = "tcp_server"

func init() {
	godotenv.Load()
	viper.SetConfigType("toml")
	viper.AddConfigPath("./config")
	viper.SetConfigName("config")
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	viper.AutomaticEnv()
	if err := viper.ReadInConfig(); err != nil {
		panic(fmt.Errorf("Fatal error config file: %s \n", err))
	}
	InitLogger()
	serverName = viper.GetString("service")
	_ = InitTrace(serverName)
}

func main() {
	if serverName == "tcp_client" {
		client.Start(context.TODO())
	} else {
		server.Start(context.TODO())
	}
}

func InitLogger() {
	logger, _ := zap.NewProduction(zap.AddCaller())
	defer logger.Sync()
	zap.ReplaceGlobals(logger)
}

func InitTrace(serviceName string) error {
	addr := viper.GetString("COLLECTOR_ADDR")
	if addr != "" {
		cfg := jaegercfg.Configuration{
			ServiceName: serviceName,
			Sampler: &jaegercfg.SamplerConfig{
				Type:  jaeger.SamplerTypeConst,
				Param: 1,
			},
			Reporter: &jaegercfg.ReporterConfig{
				LogSpans:            true,
				BufferFlushInterval: 1 * time.Second,
				LocalAgentHostPort:  addr,
			},
		}
		jLogger := jaeger.StdLogger
		jMetricsFactory := metrics.NullFactory
		tracer, _, err := cfg.NewTracer(
			jaegercfg.Logger(jLogger),
			jaegercfg.Metrics(jMetricsFactory),
		)
		if err != nil {
			return err
		}
		opentracing.SetGlobalTracer(tracer)
	}
	
	return nil
}
