package packetio

import (
	"github.com/opentracing/opentracing-go"
)

type tcpMetadataCarrier map[string]interface{}

func (tmd tcpMetadataCarrier) ForeachKey(handler func(key, val string) error) error {
	for k, val := range tmd {
		v, ok := val.(string)
		if !ok {
			continue
		}
		if err := handler(k, v); err != nil {
			return err
		}
	}
	return nil
}

func (tmd tcpMetadataCarrier) Set(key, val string) {
	tmd[key] = val
}

func Inject(span opentracing.Span, tmd map[string]interface{}) error {
	c := tcpMetadataCarrier(tmd)
	return span.Tracer().Inject(span.Context(), opentracing.TextMap, c)
}

func Extract(tmd map[string]interface{}) (opentracing.SpanContext, error) {
	c := tcpMetadataCarrier(tmd)
	return opentracing.GlobalTracer().Extract(opentracing.TextMap, c)
}
