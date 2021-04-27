package main

import (
	"context"
	"log"
	"net"
	"packetio"
	"time"
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	conn, err := net.Dial("tcp", "127.0.0.1:8000")
	defer cancel()
	if err != nil {
		panic(err)
	}
	pkt := packetio.NewPktIO(conn)
	go write(ctx, pkt)
	for {
		msg, err := pkt.Read(ctx)
		if err != nil {
			return
		}
		log.Println("receive:", msg)
	}
}

func write(ctx context.Context, pkt packetio.PacketIo) {
	var t = time.NewTicker(time.Second * 2)
	for {
		select {
		case <-ctx.Done():
			return
		case <-t.C:
			var msg = &packetio.Message{
				Cmd:     1,
				Content: []byte("hello world"),
			}
			if err := pkt.Write(msg); err != nil {
				return
			}
		}
	}
}
