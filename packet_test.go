package rudp

import (
	"fmt"
	"testing"
)

func TestPacketResolve(t *testing.T) {
	testcase := []func(){
		func() {
			payload := []byte{1, 2, 3, 100}
			packet := newNormalPacket(payload, 1000)
			fmt.Printf("%+v\n", packet)
			raw := packet.marshal()
			packet, _ = unmarshalRUDPPacket(raw)
			fmt.Printf("%+v\n", packet)
		},
		func() {
			packet := newConPacket(1000)
			fmt.Printf("%+v\n", packet)
			raw := packet.marshal()
			packet, _ = unmarshalRUDPPacket(raw)
			fmt.Printf("%+v\n", packet)
		},
		func() {
			packet := newConAckPacket(1000)
			fmt.Printf("%+v\n", packet)
			raw := packet.marshal()
			packet, _ = unmarshalRUDPPacket(raw)
			fmt.Printf("%+v\n", packet)
		},
		func() {
			packet := newFinPacket()
			fmt.Printf("%+v\n", packet)
			raw := packet.marshal()
			packet, _ = unmarshalRUDPPacket(raw)
			fmt.Printf("%+v\n", packet)
		},
		func() {
			packet := newPinPacket()
			fmt.Printf("%+v\n", packet)
			raw := packet.marshal()
			packet, _ = unmarshalRUDPPacket(raw)
			fmt.Printf("%+v\n", packet)
		},
		func() {
			packet := newAckPacket(1000)
			fmt.Printf("%+v\n", packet)
			raw := packet.marshal()
			packet, _ = unmarshalRUDPPacket(raw)
			fmt.Printf("%+v\n", packet)
		},
	}

	testcase[0]()
}
