package main

import (
	"fmt"
	"log"
	"net"

	"github.com/l-f-h/rudp"
)

func main() {
	rudp.Debug()
	listener, err := rudp.ListenRUDP(&net.UDPAddr{
		Port: 9999,
	})

	if err != nil {
		log.Fatalf("rudp.ListenRUDP error: %v", err)
	}

	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Fatalf("listener.Accept error: %v", err)
		}

		go func() {
			buf := make([]byte, 10)
			for {
				n, err := conn.Read(buf)
				if err != nil {
					log.Fatalf("conn.Read error: %v", err)
				}
				buf = buf[:n]
				fmt.Printf("%s", string(buf))
			}
		}()
	}

}
