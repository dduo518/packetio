package main

import (
	"context"
	"log"
	"net"
	"packetio"
)

func main() {

	ln, err := net.Listen("tcp", "0.0.0.0:8000")
	if err != nil {
		panic(err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	for {
		select {
		case <-ctx.Done():
			return
		default:
			log.Println("listening:tcp 8000")
			conn, err := ln.Accept()
			log.Println("connected 8000")
			if err != nil {
				return
			}

			go func() {
				pkt := packetio.NewPktIO(conn)
				for {
					msg, err := pkt.Read(ctx)
					if err != nil {
						return
					}
					log.Println("receive:", msg)
					pkt.Write(msg)
				}
			}()
		}
	}
}
