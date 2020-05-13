package rudp

import (
	"fmt"
	"math/rand"
	"testing"
	"time"
)

var seq uint32 = 0
var r = rand.New(rand.NewSource(time.Now().Unix()))
var dataList = []string{"fh", "and", "dx", "go", "to", "eat", "the", "kfc"}

func TestSorted(t *testing.T) {
	l := newPacketList(packetListOrderBySeqNb)
	for cnt := 0; cnt < 15; cnt++ {
		l.putPacket(productRandomPacket())
	}
	l.debug()
}

func TestPacketList(t *testing.T) {
	l := newPacketList(packetListOrderBySeqNb)

	go func() {
		for {
			l.putPacket(productRandomPacket())
			time.Sleep(time.Millisecond * 20)
		}
	}()

	go func() {
		for {
			fmt.Printf("%d\n", l.consume().seqNumber)
			l.debug()
			time.Sleep(time.Millisecond * 200)
		}
	}()

	select {}
}

func TestConsumeExpirePacket(t *testing.T) {
	l := newPacketList(packetListOrderBySeqNb)

	go func() {
		for {
			l.putPacket(productPacket())
			time.Sleep(time.Millisecond * 50)
		}
	}()

	go func() {
		for {
			resendList := l.consumePacketSinceNMs(600)
			fmt.Printf("%d packet need to resend\n", len(resendList))
			for _, apacket := range resendList {
				fmt.Printf("resend %d packet: %s\n", apacket.seqNumber, string(apacket.payload))
			}
			time.Sleep(time.Millisecond * 1000)
		}
	}()

	select {}

}

func productPacket() *packet {
	r := &packet{
		segmentType: rudpSegmentTypeNormal,
		seqNumber:   seq,
		payload:     []byte(dataList[seq%uint32(len(dataList))]),
	}
	seq++
	return r
}

func productRandomPacket() *packet {
	nb := r.Uint32() % 256
	r := &packet{
		segmentType: rudpSegmentTypeNormal,
		seqNumber:   nb,
		payload:     []byte(fmt.Sprintf("%d", nb)),
	}
	return r
}
