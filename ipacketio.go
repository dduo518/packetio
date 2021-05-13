package packetio

import (
	"context"
	"time"
)

type PacketIO interface {
	Read(ctx context.Context) (context.Context, *Message, error)
	Write(ctx context.Context, message *Message) error
}

type Option func(c *Options)

type Options struct {
	trace       bool
	readTimeout time.Duration
}

func Trace(trace bool) Option {
	return func(c *Options) {
		c.trace = trace
	}
}

func ReadTimeout(readTimeout time.Duration) Option {
	return func(c *Options) {
		c.readTimeout = readTimeout
	}
}
